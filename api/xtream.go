package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"anthology-live/auth"
	"anthology-live/config"
	"anthology-live/engine"
)

type XtreamUserInfo struct {
	Username             string   `json:"username"`
	Password             string   `json:"password"`
	Message              string   `json:"message"`
	Auth                 int      `json:"auth"`
	Status               string   `json:"status"`
	ExpDate              string   `json:"exp_date"`
	IsTrial              string   `json:"is_trial"`
	ActiveCons           string   `json:"active_cons"`
	CreatedAt            string   `json:"created_at"`
	MaxConnections       string   `json:"max_connections"`
	AllowedOutputFormats []string `json:"allowed_output_formats"`
}

type XtreamServerInfo struct {
	URL            string `json:"url"`
	Port           string `json:"port"`
	HTTPSPort      string `json:"https_port"`
	ServerProtocol string `json:"server_protocol"`
	RTMPPort       string `json:"rtmp_port"`
	Timezone       string `json:"timezone"`
	TimestampNow   int64  `json:"timestamp_now"`
	TimeNow        string `json:"time_now"`
}

type XtreamAuthResponse struct {
	UserInfo   XtreamUserInfo   `json:"user_info"`
	ServerInfo XtreamServerInfo `json:"server_info"`
}

type XtreamCategory struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	ParentID     int    `json:"parent_id"`
}

type XtreamLiveStream struct {
	Num                 int    `json:"num"`
	Name                string `json:"name"`
	StreamType          string `json:"stream_type"`
	StreamID            int    `json:"stream_id"`
	StreamIcon          string `json:"stream_icon"`
	EPGChannelID        string `json:"epg_channel_id"`
	Added               string `json:"added"`
	CategoryID          string `json:"category_id"`
	CustomSID           string `json:"custom_sid"`
	TVArchive           int    `json:"tv_archive"`
	DirectSource        string `json:"direct_source"`
	TVArchiveDuration   int    `json:"tv_archive_duration"`
}

// HandleXtreamAPI implements Xtream Codes API (player_api.php).
func HandleXtreamAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	user, ok := auth.UserFromContext(r.Context())
	username := "guest"
	if ok && user != nil {
		username = user.Username
	}

	cfg := config.Get()
	action := r.URL.Query().Get("action")

	now := time.Now()
	expTime := now.Add(365 * 24 * time.Hour).Unix()

	// 1. Authentication Info (No action or action == "")
	if action == "" {
		res := XtreamAuthResponse{
			UserInfo: XtreamUserInfo{
				Username:             username,
				Password:             "******",
				Message:              "Anthology Live Aktif",
				Auth:                 1,
				Status:               "Active",
				ExpDate:              strconv.FormatInt(expTime, 10),
				IsTrial:              "0",
				ActiveCons:           "1",
				CreatedAt:            strconv.FormatInt(now.Unix(), 10),
				MaxConnections:       "5",
				AllowedOutputFormats: []string{"m3u8", "ts"},
			},
			ServerInfo: XtreamServerInfo{
				URL:            cfg.Host,
				Port:           strconv.Itoa(cfg.Port),
				HTTPSPort:      "",
				ServerProtocol: "http",
				RTMPPort:       "",
				Timezone:       "Europe/Istanbul",
				TimestampNow:   now.Unix(),
				TimeNow:        now.Format("2006-01-02 15:04:05"),
			},
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	// Category mapping
	getCategoryDetails := func(catName string) (id string, displayName string) {
		switch strings.ToUpper(catName) {
		case "SPOR KANALLARI", "SPOR":
			return "1", "⚽️ Spor Kanalları"
		case "ULUSAL KANALLAR", "ULUSAL":
			return "2", "📺 Ulusal Kanallar"
		case "HABER KANALLARI", "HABER":
			return "3", "📰 Haber Kanalları"
		case "SİNEMA KANALLARI", "SINEMA KANALLARI", "SİNEMA", "SINEMA":
			return "4", "🎬 Sinema Kanalları"
		case "BELGESEL & ÇOCUK", "BELGESEL & COCUK", "BELGESEL":
			return "5", "🦁 Belgesel & Çocuk"
		case "MÜZİK & EĞLENCE", "MUZIK & EGLENCE", "MÜZİK", "MUZIK":
			return "6", "🎵 Müzik & Eğlence"
		default:
			return "7", "✨ Diğer Kanallar"
		}
	}

	// 2. Categories
	if action == "get_live_categories" {
		cats := []XtreamCategory{
			{CategoryID: "1", CategoryName: "⚽️ Spor Kanalları", ParentID: 0},
			{CategoryID: "2", CategoryName: "📺 Ulusal Kanallar", ParentID: 0},
			{CategoryID: "3", CategoryName: "📰 Haber Kanalları", ParentID: 0},
			{CategoryID: "4", CategoryName: "🎬 Sinema Kanalları", ParentID: 0},
			{CategoryID: "5", CategoryName: "🦁 Belgesel & Çocuk", ParentID: 0},
			{CategoryID: "6", CategoryName: "🎵 Müzik & Eğlence", ParentID: 0},
			{CategoryID: "7", CategoryName: "✨ Diğer Kanallar", ParentID: 0},
		}
		json.NewEncoder(w).Encode(cats)
		return
	}

	// 3. Live Streams Catalog
	if action == "get_live_streams" {
		categoryFilter := r.URL.Query().Get("category_id")
		streams := make([]XtreamLiveStream, 0, len(engine.ChannelList))

		for idx, ch := range engine.ChannelList {
			catID, _ := getCategoryDetails(ch.Category)
			if categoryFilter != "" && categoryFilter != catID {
				continue
			}

			streams = append(streams, XtreamLiveStream{
				Num:               idx + 1,
				Name:              ch.Name,
				StreamType:        "live",
				StreamID:          ch.StreamID,
				StreamIcon:        ch.LogoURL,
				EPGChannelID:      fmt.Sprintf("%s.tr", ch.Key),
				Added:             strconv.FormatInt(now.Unix(), 10),
				CategoryID:        catID,
				CustomSID:         "",
				TVArchive:         0,
				DirectSource:      "",
				TVArchiveDuration: 0,
			})
		}
		json.NewEncoder(w).Encode(streams)
		return
	}

	// VOD & Series empty fallbacks for IPTV apps
	if action == "get_vod_categories" || action == "get_vod_streams" || action == "get_series_categories" || action == "get_series" {
		w.Write([]byte("[]"))
		return
	}

	http.Error(w, `{"error":"unsupported_action"}`, http.StatusBadRequest)
}
