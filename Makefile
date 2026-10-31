BINARY_NAME := anthology-live
LINUX_BINARY := anthology-live-linux

.PHONY: help build build-linux run test clean install-service uninstall-service status logs

help:
	@echo "Anthology Live (Go v2) — Canlı TV & Spor Akış Servisi"
	@echo ""
	@echo "Kullanılabilir Komutlar:"
	@echo "  make run            : Sunucuyu doğrudan başlatır (go run main.go)"
	@echo "  make build          : macOS için ikili dosya (binary) derler"
	@echo "  make build-linux    : VPS için Linux AMD64 binary derler"
	@echo "  make test           : Birim testleri çalıştırır (go test ./...)"
	@echo "  make clean          : Derlenen dosyaları temizler"

build:
	@echo "🔨 macOS için derleniyor..."
	@go build -o $(BINARY_NAME) main.go
	@echo "✅ Derleme tamamlandı: ./$(BINARY_NAME)"

build-linux:
	@echo "🐧 Linux (AMD64) VPS için derleniyor..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(LINUX_BINARY) main.go
	@echo "✅ Linux derlemesi tamamlandı: ./$(LINUX_BINARY)"

run:
	@go run main.go

test:
	@go test -v ./...

clean:
	@rm -f $(BINARY_NAME) $(LINUX_BINARY)
	@echo "🧹 Temizlendi."

PLIST_NAME := com.falsisdev.anthology-live.plist
LAUNCH_AGENTS_DIR := $(HOME)/Library/LaunchAgents
PLIST_SRC := scripts/$(PLIST_NAME)
PLIST_DEST := $(LAUNCH_AGENTS_DIR)/$(PLIST_NAME)

install-service: build
	@mkdir -p $(LAUNCH_AGENTS_DIR)
	@cp $(PLIST_SRC) $(PLIST_DEST)
	@launchctl unload $(PLIST_DEST) 2>/dev/null || true
	@launchctl load $(PLIST_DEST)
	@echo "✅ Anthology Live (Go) servisi kuruldu ve arka planda başlatıldı!"
	@echo "📱 IPTV M3U: http://localhost:3555/iptv.m3u"
	@echo "🌐 Web TV Portalı: http://localhost:3555/"

uninstall-service:
	@launchctl unload $(PLIST_DEST) 2>/dev/null || true
	@rm -f $(PLIST_DEST)
	@echo "🛑 Anthology Live servisi durduruldu ve başlangıçtan kaldırıldı."

status:
	@launchctl list | grep anthology-live || echo "Servis çalışmıyor."

logs:
	@tail -f /tmp/anthology-live.log
