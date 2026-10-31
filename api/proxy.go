package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"anthology-live/auth"
	"anthology-live/config"
	"anthology-live/engine"
)

// HandleLiveStream serves the HLS manifest for a given channel.
func HandleLiveStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/live/")
	path = strings.TrimSuffix(path, ".m3u8")
	path = strings.TrimSuffix(path, ".ts")

	// Support Xtream URL pattern: /live/:username/:password/:streamId
	parts := strings.Split(path, "/")
	channelIdentifier := parts[len(parts)-1]

	source := r.URL.Query().Get("source")
	if source == "" {
		source = r.URL.Query().Get("provider")
	}

	// Allow URL patterns like /live/beinsports1_alemdar.m3u8
	if strings.Contains(channelIdentifier, "_") {
		subParts := strings.SplitN(channelIdentifier, "_", 2)
		channelIdentifier = subParts[0]
		if source == "" {
			source = subParts[1]
		}
	}

	var channelKey string
	if streamID, err := strconv.Atoi(channelIdentifier); err == nil {
		baseID, virtProv := engine.ProviderFromVirtualID(streamID)
		if source == "" && virtProv != "" {
			source = virtProv
		}
		if ch, ok := engine.FindChannelByID(baseID); ok {
			channelKey = ch.Key
		}
	} else {
		channelKey = strings.ToLower(channelIdentifier)
	}

	if channelKey == "" {
		http.Error(w, "Geçersiz kanal", http.StatusBadRequest)
		return
	}

	cfg := config.Get()
	host := r.Host
	if host == "" {
		host = fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	// Preserve credentials for segment links
	authQuery := ""
	u, p, token := auth.ExtractCredentials(r)
	if u != "" && p != "" {
		authQuery = fmt.Sprintf("&username=%s&password=%s", u, p)
	} else if token != "" {
		authQuery = fmt.Sprintf("&token=%s", token)
	}

	restreamer := engine.GetRestreamer()
	manifest, err := restreamer.RewriteManifest(r.Context(), channelKey, source, baseURL, authQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Yayın çözülemedi: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(manifest))
}

// HandleSubManifest proxies HLS variant sub-manifests.
func HandleSubManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	targetEnc := r.URL.Query().Get("u")
	refEnc := r.URL.Query().Get("r")
	origEnc := r.URL.Query().Get("o")
	tokenEnc := r.URL.Query().Get("t")

	targetBytes, err := base64.RawURLEncoding.DecodeString(targetEnc)
	if err != nil || len(targetBytes) == 0 {
		http.Error(w, "Geçersiz hedef URL", http.StatusBadRequest)
		return
	}

	targetURL := string(targetBytes)
	referer := decodeB64(refEnc)
	origin := decodeB64(origEnc)
	token := decodeB64(tokenEnc)

	cfg := config.Get()
	host := r.Host
	if host == "" {
		host = fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	authQuery := ""
	u, p, t := auth.ExtractCredentials(r)
	if u != "" && p != "" {
		authQuery = fmt.Sprintf("&username=%s&password=%s", u, p)
	} else if t != "" {
		authQuery = fmt.Sprintf("&token=%s", t)
	}

	restreamer := engine.GetRestreamer()
	manifest, err := restreamer.RewriteSubManifest(r.Context(), targetURL, referer, origin, token, baseURL, authQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Alt manifest hatası: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(manifest))
}

// HandleSegment serves video chunks (.ts / .jpg disguised TS) from cache or upstream.
func HandleSegment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	targetEnc := r.URL.Query().Get("u")
	refEnc := r.URL.Query().Get("r")
	origEnc := r.URL.Query().Get("o")
	tokenEnc := r.URL.Query().Get("t")

	targetBytes, err := base64.RawURLEncoding.DecodeString(targetEnc)
	if err != nil || len(targetBytes) == 0 {
		http.Error(w, "Geçersiz segment URL", http.StatusBadRequest)
		return
	}

	targetURL := string(targetBytes)
	referer := decodeB64(refEnc)
	origin := decodeB64(origEnc)
	token := decodeB64(tokenEnc)

	restreamer := engine.GetRestreamer()
	data, err := restreamer.FetchSegment(r.Context(), targetURL, referer, origin, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Segment hatası: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=25")

	w.Write(data)
}

func decodeB64(s string) string {
	if s == "" {
		return ""
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(b)
}

// HandleChannelSources returns the list of available providers for a channel.
func HandleChannelSources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	channelKey := r.URL.Query().Get("channel")
	if channelKey == "" {
		channelKey = r.URL.Query().Get("key")
	}
	if channelKey == "" {
		http.Error(w, `{"error":"missing channel parameter"}`, http.StatusBadRequest)
		return
	}

	providers := engine.GetAvailableProviders(channelKey)
	json.NewEncoder(w).Encode(providers)
}
