package api

import (
	"fmt"
	"net/http"
	"strings"

	"anthology-live/auth"
	"anthology-live/config"
	"anthology-live/engine"
)

// HandleM3UPlaylist generates an M3U8 / M3U Plus playlist for all channels.
func HandleM3UPlaylist(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	user, _ := auth.UserFromContext(r.Context())

	// Build auth query string to forward to stream URLs
	authQuery := ""
	u, p, token := auth.ExtractCredentials(r)
	if u != "" && p != "" {
		authQuery = fmt.Sprintf("?username=%s&password=%s", u, p)
	} else if token != "" {
		authQuery = fmt.Sprintf("?token=%s", token)
	} else if user != nil {
		authQuery = fmt.Sprintf("?username=%s", user.Username)
	}

	host := r.Host
	if host == "" {
		host = fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("#EXTM3U x-tvg-url=\"%s\"\n\n", cfg.EPGURL))

	for _, ch := range engine.ChannelList {
		groupTitle := "Spor Kanalları"
		if ch.Category == "Ulusal" {
			groupTitle = "Ulusal Kanallar"
		}

		sb.WriteString(fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s.tr\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\",%s\n",
			ch.Key, ch.Name, ch.LogoURL, groupTitle, ch.Name))
		sb.WriteString(fmt.Sprintf("%s/live/%s.m3u8%s\n\n", baseURL, ch.Key, authQuery))
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\"anthology_live.m3u\"")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	w.Write([]byte(sb.String()))
}
