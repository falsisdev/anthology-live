package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"anthology-live/api"
	"anthology-live/auth"
	"anthology-live/config"
	"anthology-live/engine"
	"anthology-live/web"
)

func main() {
	cfg := config.Get()

	log.Println("============================================================")
	log.Println("🚀 Anthology Live (Go v2) Başlatılıyor...")
	log.Println("============================================================")

	// 1. Initialize SQLite Auth Database (matching anthology-dav)
	_, err := auth.InitDB(cfg.DBPath)
	if err != nil {
		log.Printf("⚠️ [Auth] SQLite açılamadı (%s): %v", cfg.DBPath, err)
		log.Println("ℹ️  Kimlik doğrulama misafir (guest) modunda devam edecek.")
	} else {
		log.Printf("✅ [Auth] Veritabanı bağlandı: %s", cfg.DBPath)
	}

	// 2. Start Self-Healing Background Health Monitor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.StartHealthMonitor(ctx)

	// 3. Setup HTTP Router
	mux := http.NewServeMux()

	// Web TV Frontend (Embedded HTML/JS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			data, err := web.TemplateFS.ReadFile("template/index.html")
			if err != nil {
				http.Error(w, "Web portal yüklenemedi", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}
		http.NotFound(w, r)
	})

	// Xtream Codes API (Smart TV apps like TiviMate, Smarters Pro, Televizo)
	mux.Handle("/player_api.php", auth.AuthMiddleware(http.HandlerFunc(api.HandleXtreamAPI)))

	// M3U Playlist Downloads (Xtream get.php & Standard /iptv.m3u)
	m3uHandler := auth.AuthMiddleware(http.HandlerFunc(api.HandleM3UPlaylist))
	mux.Handle("/get.php", m3uHandler)
	mux.Handle("/iptv.m3u", m3uHandler)
	mux.Handle("/playlist.m3u", m3uHandler)

	// HLS Stream & Segment Proxies
	mux.HandleFunc("/live/", func(w http.ResponseWriter, r *http.Request) {
		api.HandleLiveStream(w, r)
	})
	mux.HandleFunc("/sub_manifest", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSubManifest(w, r)
	})
	mux.HandleFunc("/segment", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSegment(w, r)
	})
	mux.HandleFunc("/api/sources", func(w http.ResponseWriter, r *http.Request) {
		api.HandleChannelSources(w, r)
	})

	// Wrap with Global CORS & Logging Middleware
	handler := corsAndLogMiddleware(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// Print Startup Details
	localIP := config.GetLocalIPv4()
	log.Println("============================================================")
	log.Printf("🌐 Web TV Portalı:          http://%s:%d/", localIP, cfg.Port)
	log.Printf("📱 IPTV M3U Linki:          http://%s:%d/iptv.m3u", localIP, cfg.Port)
	log.Printf("📺 Xtream Codes Sunucu URL: http://%s:%d", localIP, cfg.Port)
	log.Printf("💻 Yerel Erişim (Localhost): http://127.0.0.1:%d/", cfg.Port)
	log.Println("============================================================")

	// Graceful shutdown handling
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Sunucu hatası: %v", err)
		}
	}()

	<-stopChan
	log.Println("🛑 Kapatma sinyali alındı, sunucu durduruluyor...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)

	log.Println("👋 Anthology Live başarıyla kapatıldı.")
}

func corsAndLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Skip high-frequency segment logs
		if !strings.HasPrefix(r.URL.Path, "/segment") {
			log.Printf("📡 [%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		}

		next.ServeHTTP(w, r)
	})
}
