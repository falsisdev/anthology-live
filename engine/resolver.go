package engine

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	flotfRegex  = regexp.MustCompile(`flotfUrl\s*=\s*[` + "`" + `'""](https:\/\/[^` + "`" + `'""]+index\.m3u8)`)
	m3u8Regex   = regexp.MustCompile(`https?:\/\/[^"'\s\\]+\.m3u8`)
	ardaIframeRe = regexp.MustCompile(`src="(/loadstream\.php\?[^"]+)"`)
	ardaB64Re   = regexp.MustCompile(`['"](d2lu[A-Za-z0-9+/=]+)\|`)

	mahsunBaseCache string
	mahsunBaseMu    sync.RWMutex
	mahsunBaseTime  time.Time
)

// StreamResult holds resolved stream metadata.
type StreamResult struct {
	ChannelKey  string            `json:"channel_key"`
	ChannelName string            `json:"channel_name"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Provider    string            `json:"provider"`
}

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
	case "direct", "resmi", "official":
		return "direct"
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
	if ch.DirectURL != "" {
		providers = append(providers, ProviderInfo{ID: "direct", Name: "📡 Resmi / Doğrudan Akış", Available: true})
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
	ch, ok := FindChannelByKey(channelKey)
	if !ok {
		return nil, fmt.Errorf("kanal bulunamadı: %s", channelKey)
	}

	// 1. Direct stream request
	if provider == "direct" {
		if ch.DirectURL != "" {
			return &StreamResult{
				ChannelKey:  ch.Key,
				ChannelName: ch.Name,
				URL:         ch.DirectURL,
				Headers: map[string]string{
					"User-Agent": DefaultUA,
				},
				Provider: "direct",
			}, nil
		}
		return nil, fmt.Errorf("bu kanal için doğrudan resmi akış bulunmuyor")
	}

	// 2. Auto selection
	if provider == "" || provider == "auto" {
		return ResolveBestStream(ctx, channelKey)
	}

	// 3. Explicit provider requested -> DO NOT FALLBACK, strictly return that provider's stream!
	var resolver func(context.Context, ChannelMapping) (*StreamResult, error)
	switch provider {
	case "alemdarsport":
		resolver = resolveAlemdar
	case "ardasports":
		resolver = resolveArda
	case "taraftarium":
		resolver = resolveTaraftarium
	case "mahsunsports":
		resolver = resolveMahsun
	case "papazsports":
		resolver = resolvePapaz
	default:
		return nil, fmt.Errorf("desteklenmeyen kaynak: %s", provider)
	}

	res, err := resolver(ctx, ch)
	if err != nil || res == nil || res.URL == "" {
		return nil, fmt.Errorf("%s kaynağında bu kanal şu an aktif değil", provider)
	}

	if !testStreamAlive(ctx, res) {
		return nil, fmt.Errorf("%s kaynağı yanıt vermiyor (HTTP 403/404)", provider)
	}

	res.ChannelKey = ch.Key
	res.ChannelName = ch.Name
	return res, nil
}

// ResolveBestStream queries all available providers concurrently and returns the best working stream.
func ResolveBestStream(ctx context.Context, channelKey string) (*StreamResult, error) {
	ch, ok := FindChannelByKey(channelKey)
	if !ok {
		return nil, nil
	}

	// For non-sports channels with direct official stream, return immediately
	if ch.Taraftarium == "" && ch.ArdaSports == "" && ch.AlemdarSport == "" && ch.MahsunSports == "" {
		if ch.DirectURL != "" {
			return &StreamResult{
				ChannelKey:  ch.Key,
				ChannelName: ch.Name,
				URL:         ch.DirectURL,
				Headers: map[string]string{
					"User-Agent": DefaultUA,
				},
				Provider: "direct",
			}, nil
		}
	}

	type providerJob struct {
		name     string
		resolver func(context.Context, ChannelMapping) (*StreamResult, error)
	}

	jobs := []providerJob{
		{"alemdarsport", resolveAlemdar},
		{"ardasports", resolveArda},
		{"taraftarium", resolveTaraftarium},
		{"mahsunsports", resolveMahsun},
		{"papazsports", resolvePapaz},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()

	resultChan := make(chan *StreamResult, len(jobs)+1)
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

	// Also check direct URL if available
	if ch.DirectURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resultChan <- &StreamResult{
				URL:      ch.DirectURL,
				Headers:  map[string]string{"User-Agent": DefaultUA},
				Provider: "direct",
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	priorityWeight := map[string]int{
		"alemdarsport": 5,
		"ardasports":   4,
		"taraftarium":  3,
		"mahsunsports": 2,
		"papazsports":  1,
		"direct":       6, // Direct official streams have highest priority if valid
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

	for _, cand := range candidates {
		if testStreamAlive(ctx, cand) {
			cand.ChannelKey = ch.Key
			cand.ChannelName = ch.Name
			return cand, nil
		}
	}

	return nil, fmt.Errorf("hiçbir yayından akış alınamadı")
}

// ── 1. AlemdarSport Resolver ────────────────────────────────────────────────
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

// ── 2. ArdaSports / SelcukSports Resolver ────────────────────────────────────
func resolveArda(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.ArdaSports == "" {
		return nil, nil
	}
	ardaBase := "https://selcuksporkral2.cc"
	watchURL := ardaBase + "/channel/watch/" + url.PathEscape(ch.ArdaSports)

	// Step 1: Request watch page to retrieve iframe URL
	req1, err := http.NewRequestWithContext(ctx, "GET", watchURL, nil)
	if err != nil {
		return nil, err
	}
	req1.Header.Set("User-Agent", DefaultUA)
	req1.Header.Set("Referer", ardaBase+"/canli-izle/"+ch.ArdaSports)
	req1.Header.Set("Cookie", "beget=begetok;")

	resp1, err := resolverHTTP.Do(req1)
	if err != nil {
		return nil, err
	}
	defer resp1.Body.Close()

	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		return nil, err
	}

	iframeMatch := ardaIframeRe.FindStringSubmatch(string(body1))
	if len(iframeMatch) < 2 {
		return nil, nil
	}

	// Step 2: Request iframe /loadstream.php?...
	loadURL := ardaBase + iframeMatch[1]
	req2, err := http.NewRequestWithContext(ctx, "GET", loadURL, nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("User-Agent", DefaultUA)
	req2.Header.Set("Referer", watchURL)
	req2.Header.Set("Cookie", "beget=begetok;")

	resp2, err := resolverHTTP.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, err
	}

	// Step 3: Extract and decode Base64 config
	b64Match := ardaB64Re.FindStringSubmatch(string(body2))
	if len(b64Match) > 1 {
		b64Str := b64Match[1]
		if rem := len(b64Str) % 4; rem != 0 {
			b64Str += strings.Repeat("=", 4-rem)
		}
		decoded, err := base64.StdEncoding.DecodeString(b64Str)
		if err == nil {
			if streamM := m3u8Regex.FindString(string(decoded)); streamM != "" {
				return &StreamResult{
					URL: streamM,
					Headers: map[string]string{
						"User-Agent": DefaultUA,
						"Referer":    ardaBase + "/",
						"Origin":     ardaBase,
					},
					Provider: "ardasports",
				}, nil
			}
		}
	}

	return nil, nil
}

// ── 3. Taraftarium24 Resolver ────────────────────────────────────────────────
func resolveTaraftarium(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.Taraftarium == "" {
		return nil, nil
	}
	siteURL := "https://taraftarium24-246.top"
	apiBase := "https://teletv15.top"

	apiURL := fmt.Sprintf("%s/load/yayinlink.php?id=%s", apiBase, url.QueryEscape(ch.Taraftarium))
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Referer", siteURL+"/")
	req.Header.Set("Origin", siteURL)

	resp, err := resolverHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		DeismaKanal string `json:"deismackanal"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil && data.DeismaKanal != "" {
		finalURL := strings.ReplaceAll(data.DeismaKanal, "edge4.", "edge3.")
		return &StreamResult{
			URL: finalURL,
			Headers: map[string]string{
				"User-Agent": DefaultUA,
				"Referer":    siteURL + "/",
				"Origin":     siteURL,
			},
			Provider: "taraftarium",
		}, nil
	}

	return nil, nil
}

// ── 4. MahsunSports Resolver ─────────────────────────────────────────────────
func resolveMahsun(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.MahsunSports == "" {
		return nil, nil
	}

	bases := []string{
		"https://andro.evrenesoglu211.click/checklist/",
		"https://andro.evrenesoglu215.click/checklist/",
		"https://andro.evrenesoglu217.click/checklist/",
	}

	for _, base := range bases {
		streamURL := base + ch.MahsunSports + ".m3u8"
		res := &StreamResult{
			URL: streamURL,
			Headers: map[string]string{
				"User-Agent": DefaultUA,
				"Referer":    "https://mahsunsports57.xyz/",
			},
			Provider: "mahsunsports",
		}
		if testStreamAlive(ctx, res) {
			return res, nil
		}
	}

	return nil, nil
}

// ── 5. PapazSports Resolver ─────────────────────────────────────────────────
func resolvePapaz(ctx context.Context, ch ChannelMapping) (*StreamResult, error) {
	if ch.PapazSports == "" {
		return nil, nil
	}
	papazBase := "https://www.papazsports1033.pro"
	authURL := papazBase + "/auth.php"

	data := url.Values{}
	data.Set("channel", ch.PapazSports)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Referer", papazBase+"/")
	req.Header.Set("Origin", papazBase)
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
				"Referer":    papazBase + "/",
				"Origin":     papazBase,
				"usertoken":  pRes.Token,
			},
			Provider: "papazsports",
		}, nil
	}
	return nil, nil
}

func testStreamAlive(ctx context.Context, s *StreamResult) bool {
	if s == nil || s.URL == "" {
		return false
	}
	tCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(tCtx, "GET", s.URL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Range", "bytes=0-300")
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
