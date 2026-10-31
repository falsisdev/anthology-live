package engine

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"anthology-live/config"
)

var (
	scraperClient = &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	huntMu sync.Mutex
)

// HuntTaraftariumDomain tests sequential numbers for Taraftarium24.
func HuntTaraftariumDomain(ctx context.Context) (string, int, bool) {
	huntMu.Lock()
	defer huntMu.Unlock()

	cfg := config.Get()
	platforms := cfg.GetPlatforms()
	startNum := platforms.TaraftariumBaseNum

	log.Printf("🔍 [Domain Hunter] Taraftarium için aktif domain aranıyor (Başlangıç: %d)...", startNum)

	// Check current first
	currentURL := fmt.Sprintf("https://taraftarium24-%d.top/", startNum)
	if testURLAlive(ctx, currentURL) {
		return currentURL, startNum, true
	}

	// Scan up to +20 ahead
	for i := 1; i <= 20; i++ {
		candidateNum := startNum + i
		candidateURL := fmt.Sprintf("https://taraftarium24-%d.top/", candidateNum)
		if testURLAlive(ctx, candidateURL) {
			log.Printf("🎯 [Domain Hunter] Yeni Taraftarium24 bulundu: %s", candidateURL)
			cfg.SetTaraftariumBaseNum(candidateNum, candidateURL)
			return candidateURL, candidateNum, true
		}
	}

	// Try with .xyz or .cfd if .top doesn't respond
	alts := []string{
		fmt.Sprintf("https://taraftarium24-%d.xyz/", startNum),
		fmt.Sprintf("https://taraftarium24-%d.pro/", startNum),
	}
	for _, alt := range alts {
		if testURLAlive(ctx, alt) {
			log.Printf("🎯 [Domain Hunter] Alternatif Taraftarium bulundu: %s", alt)
			cfg.UpdatePlatformBase("taraftarium", alt)
			return alt, startNum, true
		}
	}

	return "", 0, false
}

func testURLAlive(ctx context.Context, target string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", target, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := scraperClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
