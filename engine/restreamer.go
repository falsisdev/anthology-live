package engine

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type CachedSegment struct {
	Data      []byte
	CreatedAt time.Time
}

type CachedManifest struct {
	Stream    *StreamResult
	CreatedAt time.Time
}

type RestreamerEngine struct {
	mu           sync.RWMutex
	segmentCache map[string]CachedSegment
	streamCache  map[string]CachedManifest
	httpClient   *http.Client
}

var (
	GlobalRestreamer *RestreamerEngine
	restreamerOnce   sync.Once
)

// GetRestreamer returns the singleton HLS Restreamer engine.
func GetRestreamer() *RestreamerEngine {
	restreamerOnce.Do(func() {
		r := &RestreamerEngine{
			segmentCache: make(map[string]CachedSegment),
			streamCache:  make(map[string]CachedManifest),
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		}
		go r.cleanupLoop()
		GlobalRestreamer = r
	})
	return GlobalRestreamer
}

func (r *RestreamerEngine) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		// Clean segments older than 25 seconds
		for k, v := range r.segmentCache {
			if now.Sub(v.CreatedAt) > 25*time.Second {
				delete(r.segmentCache, k)
			}
		}
		// Clean stream manifests older than 60 seconds
		for k, v := range r.streamCache {
			if now.Sub(v.CreatedAt) > 60*time.Second {
				delete(r.streamCache, k)
			}
		}
		r.mu.Unlock()
	}
}

// InvalidateStream clears a channel's cached stream on failure.
func (r *RestreamerEngine) InvalidateStream(channelKey string) {
	r.mu.Lock()
	prefix := channelKey + ":"
	for k := range r.streamCache {
		if k == channelKey || strings.HasPrefix(k, prefix) {
			delete(r.streamCache, k)
		}
	}
	r.mu.Unlock()
}

// GetOrResolveStream retrieves a cached stream or initiates resolution for a specific provider.
func (r *RestreamerEngine) GetOrResolveStream(ctx context.Context, channelKey, provider string) (*StreamResult, error) {
	provider = NormalizeProvider(provider)
	if provider == "" {
		provider = "auto"
	}
	cacheKey := channelKey + ":" + provider

	r.mu.RLock()
	cached, ok := r.streamCache[cacheKey]
	r.mu.RUnlock()

	if ok && time.Since(cached.CreatedAt) < 60*time.Second {
		return cached.Stream, nil
	}

	stream, err := ResolveStream(ctx, channelKey, provider)
	if err != nil || stream == nil {
		return nil, err
	}

	r.mu.Lock()
	r.streamCache[cacheKey] = CachedManifest{
		Stream:    stream,
		CreatedAt: time.Now(),
	}
	r.mu.Unlock()

	return stream, nil
}

// FetchSegment retrieves a video segment with memory caching.
func (r *RestreamerEngine) FetchSegment(ctx context.Context, targetURL, referer, origin, userToken string) ([]byte, error) {
	r.mu.RLock()
	cached, ok := r.segmentCache[targetURL]
	r.mu.RUnlock()

	if ok && time.Since(cached.CreatedAt) < 20*time.Second {
		return cached.Data, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Accept", "*/*")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if userToken != "" {
		req.Header.Set("usertoken", userToken)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upstream segment status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache segment in RAM
	r.mu.Lock()
	r.segmentCache[targetURL] = CachedSegment{
		Data:      data,
		CreatedAt: time.Now(),
	}
	r.mu.Unlock()

	return data, nil
}

// RewriteManifest fetches and translates an upstream HLS playlist into local proxy URLs.
func (r *RestreamerEngine) RewriteManifest(ctx context.Context, channelKey, provider, baseURL, authQuery string) (string, error) {
	stream, err := r.GetOrResolveStream(ctx, channelKey, provider)
	if err != nil || stream == nil {
		return "", fmt.Errorf("yayın bulunamadı")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", stream.URL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range stream.Headers {
		req.Header.Set(k, v)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.InvalidateStream(channelKey)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		r.InvalidateStream(channelKey)
		return "", fmt.Errorf("upstream manifest status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	upstreamBase, err := url.Parse(stream.URL)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(bodyBytes), "\n")
	rewritten := make([]string, 0, len(lines))

	refEnc := base64.RawURLEncoding.EncodeToString([]byte(stream.Headers["Referer"]))
	origEnc := base64.RawURLEncoding.EncodeToString([]byte(stream.Headers["Origin"]))
	tokenEnc := base64.RawURLEncoding.EncodeToString([]byte(stream.Headers["usertoken"]))

	// Forward specific source if chosen
	if provider != "" && provider != "auto" {
		authQuery += "&source=" + url.QueryEscape(provider)
	}

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			rewritten = append(rewritten, lines[i])
			continue
		}

		if strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
				rewritten = append(rewritten, line)
				i++
				if i < len(lines) {
					subLine := strings.TrimSpace(lines[i])
					subTarget, _ := upstreamBase.Parse(subLine)
					if subTarget != nil {
						subEnc := base64.RawURLEncoding.EncodeToString([]byte(subTarget.String()))
						proxiedSub := fmt.Sprintf("%s/sub_manifest?u=%s&r=%s&o=%s&t=%s%s",
							baseURL, subEnc, refEnc, origEnc, tokenEnc, authQuery)
						rewritten = append(rewritten, proxiedSub)
					}
				}
				continue
			}
			rewritten = append(rewritten, line)
		} else {
			// Media segment (.ts, .jpg, .m4s)
			segTarget, _ := upstreamBase.Parse(line)
			if segTarget != nil {
				segEnc := base64.RawURLEncoding.EncodeToString([]byte(segTarget.String()))
				proxiedSeg := fmt.Sprintf("%s/segment?u=%s&r=%s&o=%s&t=%s%s",
					baseURL, segEnc, refEnc, origEnc, tokenEnc, authQuery)
				rewritten = append(rewritten, proxiedSeg)
			}
		}
	}

	return strings.Join(rewritten, "\n"), nil
}

// RewriteSubManifest proxies sub-manifests (rendition variants).
func (r *RestreamerEngine) RewriteSubManifest(ctx context.Context, targetURL, referer, origin, userToken, baseURL, authQuery string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", DefaultUA)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if userToken != "" {
		req.Header.Set("usertoken", userToken)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	upstreamBase, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(bodyBytes), "\n")
	rewritten := make([]string, 0, len(lines))

	refEnc := base64.RawURLEncoding.EncodeToString([]byte(referer))
	origEnc := base64.RawURLEncoding.EncodeToString([]byte(origin))
	tokenEnc := base64.RawURLEncoding.EncodeToString([]byte(userToken))

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			rewritten = append(rewritten, rawLine)
		} else {
			segTarget, _ := upstreamBase.Parse(line)
			if segTarget != nil {
				segEnc := base64.RawURLEncoding.EncodeToString([]byte(segTarget.String()))
				proxiedSeg := fmt.Sprintf("%s/segment?u=%s&r=%s&o=%s&t=%s%s",
					baseURL, segEnc, refEnc, origEnc, tokenEnc, authQuery)
				rewritten = append(rewritten, proxiedSeg)
			}
		}
	}

	return strings.Join(rewritten, "\n"), nil
}
