package engine

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"anthology-live/config"
)

var (
	DefaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

	resolverHTTP = &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	flotfRegex  = regexp.MustCompile(`flotfUrl\s*=\s*[` + "`" + `'""](https:\/\/[^` + "`" + `'""]+index\.m3u8)`)
	m3u8Regex   = regexp.MustCompile(`https?:\/\/[^"'\s\\]+\.m3u8`)
	mahsunRegex = regexp.MustCompile(`["'](https:\/\/andro\.[^"']+\/checklist\/)["']`)
	b64StreamRe = regexp.MustCompile(`([A-Za-z0-9+/=]{40,})`)

	mahsunBaseCache string
	mahsunBaseMu    sync.RWMutex
	mahsunBaseTime  time.Time
)

// NormalizeProvider maps provider aliases to canonical names.
func NormalizeProvider(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	switch p {
	case "selcuk", "arda", "ardasports", "selcuksport", "selcukspor":
		return "ardasports"
	case "alemdar", "alemdarsport":
		return "alemdarsport"
	case "taraftarium", "taraftarium24":
		return "taraftarium"
	case "mahsun", "mahsunsports":
		return "mahsunsports"
	case "papaz", "papazsports":
		return "papazsports"
	default:
		return p
	}
}

// ProviderInfo represents metadata about an available stream source.
type ProviderInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// GetAvailableProviders returns the list of all supported sources for a channel.
func GetAvailableProviders(channelKey string) []ProviderInfo {
	ch, ok := FindChannelByKey(channelKey)
	if !ok {
		return nil
	}

	providers := []ProviderInfo{
		{ID: "auto", Name: "⚡️ Otomatik (En Hızlı)", Available: true},
	}
	if ch.AlemdarSport != "" {
		providers = append(providers, ProviderInfo{ID: "alemdarsport", Name: "Alemdar Sport", Available: true})
	}
	if ch.ArdaSports != "" {
		providers = append(providers, ProviderInfo{ID: "ardasports", Name: "Selçuk / Arda", Available: true})
	}
	if ch.Taraftarium != "" {
		providers = append(providers, ProviderInfo{ID: "taraftarium", Name: "Taraftarium24", Available: true})
	}
	if ch.MahsunSports != "" {
		providers = append(providers, ProviderInfo{ID: "mahsunsports", Name: "Mahsun Sports", Available: true})
	}
	if ch.PapazSports != "" {
		providers = append(providers, ProviderInfo{ID: "papazsports", Name: "Papaz Sports", Available: true})
	}
	return providers
}

// ResolveStream queries a specific provider or defaults to auto-selection.
func ResolveStream(ctx context.Context, channelKey, provider string) (*StreamResult, error) {
	provider = NormalizeProvider(provider)
	if provider == "" || provider == "auto" {
		return ResolveBestStream(ctx, channelKey)
	}

	ch, ok := FindChannelByKey(channelKey)
	if !ok {
		return nil, nil
	}

	var resolver func(context.Context, ChannelMapping) (*StreamResult, error)
	switch provider {
	case "ardasports":
		resolver = resolveArda
	case "taraftarium":
		resolver = resolveTaraftarium
	case "mahsunsports":
		resolver = resolveMahsun
	case "alemdarsport":
		resolver = resolveAlemdar
	case "papazsports":
		resolver = resolvePapaz
	default:
		return ResolveBestStream(ctx, channelKey)
	}

	res, err := resolver(ctx, ch)
	if err == nil && res != nil && res.URL != "" {
		if testStreamAlive(ctx, res) {
			res.ChannelKey = ch.Key
			res.ChannelName = ch.Name
			return res, nil
		}
	}

	// If requested provider failed, fall back to best available working stream
	return ResolveBestStream(ctx, channelKey)
}

// ResolveBestStream queries all available providers concurrently and returns the fastest valid stream.
func ResolveBestStream(ctx context.Context, channelKey string) (*StreamResult, error) {
	ch, ok := FindChannelByKey(channelKey)
	if !ok {
		return nil, nil
	}

	type providerJob struct {
		name     string
		resolver func(context.Context, ChannelMapping) (*StreamResult, error)
	}

	jobs := []providerJob{
		{"ardasports", resolveArda},
		{"taraftarium", resolveTaraftarium},
		{"mahsunsports", resolveMahsun},
		{"alemdarsport", resolveAlemdar},
		{"papazsports", resolvePapaz},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()

	resultChan := make(chan *StreamResult, len(jobs))
	var wg sync.WaitGroup

	for _, j := range jobs {
		job := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := job.resolver(ctxTimeout, ch)
			if err == nil && res != nil && res.URL != "" {
				select {
				case resultChan <- res:
				default:
				}
			}
		}()
	}

	// Close channel when all routines finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Provider priority order: Alemdar > Arda > Taraftarium > Mahsun > Papaz
	priorityWeight := map[string]int{
		"alemdarsport": 5,
		"ardasports":   4,
		"taraftarium":  3,
		"mahsunsports": 2,
		"papazsports":  1,
	}

	var candidates []*StreamResult
	for res := range resultChan {
		candidates = append(candidates, res)
	}

	// Sort candidates by priority
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if priorityWeight[candidates[j].Provider] > priorityWeight[candidates[i].Provider] {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Pick the first candidate whose stream manifest is actually live (HTTP 200)
	for _, cand := range candidates {
		if testStreamAlive(ctx, cand) {
			cand.ChannelKey = ch.Key
			cand.ChannelName = ch.Name
			return cand, nil
		}
	}

	return nil, nil
}

func testStreamAlive(ctx context.Context, s *StreamResult) bool {
	if s == nil || s.URL == "" {
		return false
	}
	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(tCtx, "GET", s.URL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Range", "bytes=0-200")
	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}

	resp, err := resolverHTTP.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// ── 1. ArdaSports / SelcukSports Resolver ────────────────────────────────────
func resolveArda(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.ArdaSports == "" {
		return nil, nil
	}
	cfg := config.Get().GetPlatforms()
	reqURL := strings.TrimRight(cfg.ArdaBase, "/") + "/loadstream.php?id=" + url.QueryEscape(ch.ArdaSports)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Referer", cfg.ArdaBase+"/")
	req.Header.Set("Cookie", "beget=begetok;")

	resp, err := resolverHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyStr := string(body)

	// Find base64 chunk
	matches := b64StreamRe.FindAllStringSubmatch(bodyStr, -1)
	for _, m := range matches {
		decoded, err := base64.StdEncoding.DecodeString(m[1])
		if err == nil {
			if streamM := m3u8Regex.FindString(string(decoded)); streamM != "" {
				return &StreamResult{
					URL: streamM,
					Headers: map[string]string{
						"User-Agent": DefaultUA,
						"Referer":    cfg.ArdaBase + "/",
						"Origin":     cfg.ArdaBase,
					},
					Provider: "ardasports",
				}, nil
			}
		}
	}

	// Try direct m3u8 in html
	if direct := m3u8Regex.FindString(bodyStr); direct != "" {
		return &StreamResult{
			URL: direct,
			Headers: map[string]string{
				"User-Agent": DefaultUA,
				"Referer":    cfg.ArdaBase + "/",
				"Origin":     cfg.ArdaBase,
			},
			Provider: "ardasports",
		}, nil
	}

	return nil, nil
}

// ── 2. Taraftarium24 Resolver ────────────────────────────────────────────────
func resolveTaraftarium(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.Taraftarium == "" {
		return nil, nil
	}
	cfg := config.Get().GetPlatforms()
	apiURL := strings.TrimRight(cfg.TaraftariumAPIBase, "/") + "/api/channel/" + url.PathEscape(ch.Taraftarium)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Referer", cfg.TaraftariumBase)
	req.Header.Set("Origin", strings.TrimRight(cfg.TaraftariumBase, "/"))

	resp, err := resolverHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Success bool `json:"success"`
		Data    struct {
			StreamURL string `json:"stream_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil && data.Success && data.Data.StreamURL != "" {
		return &StreamResult{
			URL: data.Data.StreamURL,
			Headers: map[string]string{
				"User-Agent": DefaultUA,
				"Referer":    cfg.TaraftariumBase,
				"Origin":     strings.TrimRight(cfg.TaraftariumBase, "/"),
			},
			Provider: "taraftarium",
		}, nil
	}
	return nil, nil
}

// ── 3. MahsunSports Resolver ─────────────────────────────────────────────────
func resolveMahsun(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.MahsunSports == "" {
		return nil, nil
	}

	mahsunBaseMu.RLock()
	cachedBase := mahsunBaseCache
	cachedTime := mahsunBaseTime
	mahsunBaseMu.RUnlock()

	base := cachedBase
	if base == "" || time.Since(cachedTime) > 10*time.Minute {
		cfg := config.Get().GetPlatforms()
		probeURL := strings.TrimRight(cfg.MahsunBase, "/") + "/event.html?id=androstreamlivebs1"
		req, _ := http.NewRequestWithContext(ctx, "GET", probeURL, nil)
		if req != nil {
			req.Header.Set("User-Agent", DefaultUA)
			if resp, err := resolverHTTP.Do(req); err == nil {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				if match := mahsunRegex.FindStringSubmatch(string(bodyBytes)); len(match) > 1 {
					base = match[1]
					mahsunBaseMu.Lock()
					mahsunBaseCache = base
					mahsunBaseTime = time.Now()
					mahsunBaseMu.Unlock()
				}
			}
		}
	}

	if base == "" {
		return nil, nil
	}

	return &StreamResult{
		URL: base + ch.MahsunSports + ".m3u8",
		Headers: map[string]string{
			"User-Agent": DefaultUA,
		},
		Provider: "mahsunsports",
	}, nil
}

// ── 4. AlemdarSport Resolver ────────────────────────────────────────────────
func resolveAlemdar(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.AlemdarSport == "" {
		return nil, nil
	}
	cfg := config.Get().GetPlatforms()
	channelURL := strings.TrimRight(cfg.AlemdarBase, "/") + "/channel.php?id=" + url.QueryEscape(ch.AlemdarSport)

	req, err := http.NewRequestWithContext(ctx, "GET", channelURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Referer", cfg.AlemdarBase+"/")

	resp, err := resolverHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	match := flotfRegex.FindStringSubmatch(string(bodyBytes))
	if len(match) > 1 {
		template := match[1]
		streamURL := strings.ReplaceAll(template, "${cid}", ch.AlemdarSport)
		return &StreamResult{
			URL: streamURL,
			Headers: map[string]string{
				"User-Agent": DefaultUA,
				"Referer":    cfg.AlemdarBase + "/",
				"Origin":     cfg.AlemdarBase,
			},
			Provider: "alemdarsport",
		}, nil
	}
	return nil, nil
}

// ── 5. PapazSports Resolver ─────────────────────────────────────────────────
func resolvePapaz(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.PapazSports == "" {
		return nil, nil
	}
	cfg := config.Get().GetPlatforms()
	authURL := strings.TrimRight(cfg.PapazBase, "/") + "/auth.php"

	data := url.Values{}
	data.Set("channel", ch.PapazSports)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Referer", cfg.PapazBase+"/")
	req.Header.Set("Origin", cfg.PapazBase)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := resolverHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pRes struct {
		URL   string `json:"URL"`
		Token string `json:"TOKEN"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pRes); err == nil && pRes.URL != "" && pRes.Token != "" {
		return &StreamResult{
			URL: pRes.URL,
			Headers: map[string]string{
				"User-Agent": DefaultUA,
				"Referer":    cfg.PapazBase + "/",
				"Origin":     cfg.PapazBase,
				"usertoken":  pRes.Token,
			},
			Provider: "papazsports",
		}, nil
	}
	return nil, nil
}
