package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultPort       = 3555
	DefaultEPGURL     = "https://epg.share247.net/epg.xml.gz"
	DefaultServerName = "Anthology Live"
)

// PlatformURLs holds current base URLs for streaming providers.
type PlatformURLs struct {
	TaraftariumBase    string
	TaraftariumAPIBase string
	TaraftariumBaseNum int
	ArdaBase           string
	AlemdarBase        string
	MahsunBase         string
	PapazBase          string
}

// Config represents runtime application configuration.
type Config struct {
	mu           sync.RWMutex
	Host         string
	Port         int
	BaseURL      string
	DBPath       string
	EPGURL       string
	SecretKey    string
	SegmentTTL   time.Duration
	ResolverTTL  time.Duration
	HealthTicker time.Duration
	Platforms    PlatformURLs
}

type rawConfigJSON struct {
	Port int `json:"port"`
	URLs struct {
		Live struct {
			Taraftarium struct {
				Base       string `json:"base"`
				APIBase    string `json:"api_base"`
				BaseNumber int    `json:"base_number"`
			} `json:"taraftarium"`
			ArdaSports struct {
				Base string `json:"base"`
			} `json:"ardasports"`
			AlemdarSport struct {
				Base string `json:"base"`
			} `json:"alemdarsport"`
			MahsunSports struct {
				Base string `json:"base"`
			} `json:"mahsunsports"`
			PapazSports struct {
				Base string `json:"base"`
			} `json:"papazsports"`
		} `json:"live"`
	} `json:"urls"`
}

var (
	globalCfg *Config
	cfgOnce   sync.Once
)

// GetLocalIPv4 returns the first non-loopback local IPv4 address.
func GetLocalIPv4() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

// resolveDBPath locates the sqlite database file.
func resolveDBPath() string {
	if env := os.Getenv("DB_PATH"); env != "" {
		return env
	}

	// 1. Check adjacent anthology-dav/data/anthology.db
	if wd, err := os.Getwd(); err == nil {
		adj := filepath.Join(filepath.Dir(wd), "anthology-dav", "data", "anthology.db")
		if _, err := os.Stat(adj); err == nil {
			return adj
		}
		// Also check adjacent anthology-dav root
		adjRoot := filepath.Join(filepath.Dir(wd), "anthology-dav", "anthology.db")
		if _, err := os.Stat(adjRoot); err == nil {
			return adjRoot
		}
		// 2. Check current working dir data/anthology.db
		cur := filepath.Join(wd, "data", "anthology.db")
		if _, err := os.Stat(cur); err == nil {
			return cur
		}
	}

	// Fallback path
	return filepath.Join("data", "anthology.db")
}

// Get returns the singleton configuration instance.
func Get() *Config {
	cfgOnce.Do(func() {
		port := DefaultPort
		if envPort := os.Getenv("PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
				port = p
			}
		}

		localIP := GetLocalIPv4()
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://%s:%d", localIP, port)
		}

		epgURL := os.Getenv("EPG_URL")
		if epgURL == "" {
			epgURL = DefaultEPGURL
		}

		defaultPlatforms := PlatformURLs{
			TaraftariumBase:    "https://taraftarium24-246.top/",
			TaraftariumAPIBase: "https://teletv15.top",
			TaraftariumBaseNum: 246,
			ArdaBase:           "https://selcuksporkral1.cc",
			AlemdarBase:        "https://alemdarsport.cfd",
			MahsunBase:         "https://mahsunsports58.xyz",
			PapazBase:          "https://www.papazsports1033.pro",
		}

		// Try loading config.json
		for _, candPath := range []string{"config.json", filepath.Join("..", "config.json")} {
			if data, err := os.ReadFile(candPath); err == nil {
				var r rawConfigJSON
				if err := json.Unmarshal(data, &r); err == nil {
					if r.Port > 0 && os.Getenv("PORT") == "" {
						port = r.Port
					}
					if r.URLs.Live.Taraftarium.Base != "" {
						defaultPlatforms.TaraftariumBase = r.URLs.Live.Taraftarium.Base
					}
					if r.URLs.Live.Taraftarium.APIBase != "" {
						defaultPlatforms.TaraftariumAPIBase = r.URLs.Live.Taraftarium.APIBase
					}
					if r.URLs.Live.Taraftarium.BaseNumber > 0 {
						defaultPlatforms.TaraftariumBaseNum = r.URLs.Live.Taraftarium.BaseNumber
					}
					if r.URLs.Live.ArdaSports.Base != "" {
						defaultPlatforms.ArdaBase = r.URLs.Live.ArdaSports.Base
					}
					if r.URLs.Live.AlemdarSport.Base != "" {
						defaultPlatforms.AlemdarBase = r.URLs.Live.AlemdarSport.Base
					}
					if r.URLs.Live.MahsunSports.Base != "" {
						defaultPlatforms.MahsunBase = r.URLs.Live.MahsunSports.Base
					}
					if r.URLs.Live.PapazSports.Base != "" {
						defaultPlatforms.PapazBase = r.URLs.Live.PapazSports.Base
					}
					break
				}
			}
		}

		globalCfg = &Config{
			Host:         "0.0.0.0",
			Port:         port,
			BaseURL:      baseURL,
			DBPath:       resolveDBPath(),
			EPGURL:       epgURL,
			SecretKey:    os.Getenv("SECRET_KEY"),
			SegmentTTL:   25 * time.Second,
			ResolverTTL:  60 * time.Second,
			HealthTicker: 15 * time.Second,
			Platforms:    defaultPlatforms,
		}
	})
	return globalCfg
}

// GetPlatforms thread-safely returns a copy of PlatformURLs.
func (c *Config) GetPlatforms() PlatformURLs {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Platforms
}

// UpdatePlatformBase updates a platform's base URL thread-safely.
func (c *Config) UpdatePlatformBase(platform string, newURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch platform {
	case "taraftarium":
		c.Platforms.TaraftariumBase = newURL
	case "ardasports":
		c.Platforms.ArdaBase = newURL
	case "alemdarsport":
		c.Platforms.AlemdarBase = newURL
	case "mahsunsports":
		c.Platforms.MahsunBase = newURL
	case "papazsports":
		c.Platforms.PapazBase = newURL
	}
}

// SetTaraftariumBaseNum updates the base number for Taraftarium.
func (c *Config) SetTaraftariumBaseNum(num int, siteURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Platforms.TaraftariumBaseNum = num
	c.Platforms.TaraftariumBase = siteURL
}
