package engine

import (
	"context"
	"log"
	"time"

	"anthology-live/config"
)

// StartHealthMonitor runs background stream verification and proactive warming.
func StartHealthMonitor(ctx context.Context) {
	go func() {
		cfg := config.Get()
		ticker := time.NewTicker(cfg.HealthTicker)
		defer ticker.Stop()

		log.Printf("🩺 [Health Monitor] Yayın nöbetçisi aktif (Aralık: %v)", cfg.HealthTicker)

		// Warm up key sports channels on boot
		warmupKeyChannels(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				checkAndHealChannels(ctx)
			}
		}
	}()
}

func warmupKeyChannels(ctx context.Context) {
	keyChannels := []string{"beinsports1", "beinsports2", "ssport1", "tivibuspor1"}
	restreamer := GetRestreamer()

	for _, k := range keyChannels {
		go func(channelKey string) {
			stream, err := restreamer.GetOrResolveStream(ctx, channelKey)
			if err == nil && stream != nil {
				log.Printf("⚡️ [Warmup] %s hazır: [%s] %s", channelKey, stream.Provider, stream.URL)
			}
		}(k)
	}
}

func checkAndHealChannels(ctx context.Context) {
	restreamer := GetRestreamer()
	testChannels := []string{"beinsports1", "ssport1"}

	for _, k := range testChannels {
		stream, err := restreamer.GetOrResolveStream(ctx, k)
		if err != nil || stream == nil {
			log.Printf("⚠️ [Health Check] %s yanıt vermiyor, self-healing başlatılıyor...", k)
			restreamer.InvalidateStream(k)

			// Try hunting new domain if taraftarium failed
			HuntTaraftariumDomain(ctx)

			// Retry resolving
			newStream, err := restreamer.GetOrResolveStream(ctx, k)
			if err == nil && newStream != nil {
				log.Printf("✅ [Self-Healing] %s başarıyla kurtarıldı -> [%s]", k, newStream.Provider)
			}
		}
	}
}
