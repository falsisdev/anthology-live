package engine

import (
	_ "embed"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

//go:embed canli.m3u
var EmbeddedCanliM3U []byte

// ChannelMapping defines a channel from anthology's canli.m3u and its multi-platform provider links.
type ChannelMapping struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	StreamID     int    `json:"stream_id"`
	Category     string `json:"category"`
	LogoURL      string `json:"logo_url"`
	DirectURL    string `json:"direct_url,omitempty"`
	BackupURL    string `json:"backup_url,omitempty"`
	Taraftarium  string `json:"taraftarium,omitempty"`
	ArdaSports   string `json:"ardasports,omitempty"`
	AlemdarSport string `json:"alemdarsport,omitempty"`
	MahsunSports string `json:"mahsunsports,omitempty"`
	PapazSports  string `json:"papazsports,omitempty"`
}

type SportMapping struct {
	Taraftarium  string
	ArdaSports   string
	AlemdarSport string
	MahsunSports string
	PapazSports  string
}

// SportPlatformMap provides platform identifiers for sports channels.
var SportPlatformMap = map[string]SportMapping{
	"beinsports1":     {Taraftarium: "bein-sports-1", ArdaSports: "bein-sports-1", AlemdarSport: "bein1", MahsunSports: "androstreamlivebs1", PapazSports: "100001"},
	"beinsports2":     {Taraftarium: "bein-sports-2", ArdaSports: "bein-sports-2", AlemdarSport: "bein2", MahsunSports: "androstreamlivebs2", PapazSports: "100002"},
	"beinsports3":     {Taraftarium: "bein-sports-3", ArdaSports: "bein-sports-3", AlemdarSport: "bein3", MahsunSports: "androstreamlivebs3", PapazSports: "100003"},
	"beinsports4":     {Taraftarium: "bein-sports-4", ArdaSports: "bein-sports-4", AlemdarSport: "bein4", MahsunSports: "androstreamlivebs4", PapazSports: "100004"},
	"beinsports5":     {Taraftarium: "bein-sports-5", ArdaSports: "bein-sports-5", AlemdarSport: "bein5", MahsunSports: "androstreamlivebs5", PapazSports: "100005"},
	"beinsportsmax1":  {Taraftarium: "bein-sports-max-1", ArdaSports: "bein-sports-max-1", AlemdarSport: "beinmax1", MahsunSports: "androstreamlivebsm1", PapazSports: "100006"},
	"beinsportsmax2":  {Taraftarium: "bein-sports-max-2", ArdaSports: "bein-sports-max-2", AlemdarSport: "beinmax2", MahsunSports: "androstreamlivebsm2", PapazSports: "100007"},
	"ssport":          {Taraftarium: "s-sport", ArdaSports: "ssport", AlemdarSport: "ssport1", MahsunSports: "androstreamlivess1", PapazSports: "100010"},
	"ssport1":         {Taraftarium: "s-sport", ArdaSports: "ssport", AlemdarSport: "ssport1", MahsunSports: "androstreamlivess1", PapazSports: "100010"},
	"ssport2":         {Taraftarium: "s-sport-2", ArdaSports: "ssport2", AlemdarSport: "ssport2", MahsunSports: "androstreamlivess2", PapazSports: "100011"},
	"ssportplus":      {Taraftarium: "s-sport-plus", ArdaSports: "ssportplus", AlemdarSport: "ssport1", MahsunSports: "androstreamlivessplus1"},
	"tivibuspor":      {AlemdarSport: "tivibu1", MahsunSports: "androstreamlivets", PapazSports: "100021"},
	"tivibuspor1":     {AlemdarSport: "tivibu1", MahsunSports: "androstreamlivets1", PapazSports: "100021"},
	"tivibuspor2":     {AlemdarSport: "tivibu2", MahsunSports: "androstreamlivets2", PapazSports: "100022"},
	"tivibuspor3":     {AlemdarSport: "tivibu3", MahsunSports: "androstreamlivets3", PapazSports: "100023"},
	"tivibuspor4":     {AlemdarSport: "tivibu4", MahsunSports: "androstreamlivets4", PapazSports: "100024"},
	"smartspor1":      {AlemdarSport: "sporsmart", MahsunSports: "androstreamlivesm1", PapazSports: "100030"},
	"sporsmart":       {AlemdarSport: "sporsmart", MahsunSports: "androstreamlivesm1", PapazSports: "100030"},
	"smartspor2":      {AlemdarSport: "sporsmart2", MahsunSports: "androstreamlivesm2", PapazSports: "100031"},
	"sporsmart2":      {AlemdarSport: "sporsmart2", MahsunSports: "androstreamlivesm2", PapazSports: "100031"},
	"trtspor":         {Taraftarium: "trt-spor", ArdaSports: "trtspor", AlemdarSport: "trtspor"},
	"trtsporyildiz":   {AlemdarSport: "trtspor2"},
	"aspor":           {Taraftarium: "a-spor", ArdaSports: "aspor", AlemdarSport: "aspor", PapazSports: "100052"},
	"tv85":            {AlemdarSport: "tv85", PapazSports: "100051"},
	"nbatv":           {PapazSports: "100053", MahsunSports: "androstreamlivenbatv"},
	"eurosport1":      {MahsunSports: "androstreamlivees1", PapazSports: "100056"},
	"eurosport2":      {MahsunSports: "androstreamlivees2", PapazSports: "100057"},
	"cbcsport":        {MahsunSports: "androstreamlivecbcs"},
	"idmantv":         {MahsunSports: "androstreamliveidm"},
	"exxentv":         {MahsunSports: "androstreamliveexn"},
	"exxensports1":    {MahsunSports: "androstreamliveexn1"},
	"exxensports2":    {MahsunSports: "androstreamliveexn2"},
	"exxensports3":    {MahsunSports: "androstreamliveexn3"},
	"exxensports4":    {MahsunSports: "androstreamliveexn4"},
	"exxensports5":    {MahsunSports: "androstreamliveexn5"},
	"exxensports6":    {MahsunSports: "androstreamliveexn6"},
	"exxensports7":    {MahsunSports: "androstreamliveexn7"},
	"exxensports8":    {MahsunSports: "androstreamliveexn8"},
	"tabiispor":       {MahsunSports: "androstreamlivetb"},
	"tabiispor1":      {MahsunSports: "androstreamlivetb1"},
	"tabiispor2":      {MahsunSports: "androstreamlivetb2"},
	"tabiispor3":      {MahsunSports: "androstreamlivetb3"},
	"tabiispor4":      {MahsunSports: "androstreamlivetb4"},
	"tabiispor5":      {MahsunSports: "androstreamlivetb5"},
	"tabiispor6":      {MahsunSports: "androstreamlivetb6"},
	"tabiispor7":      {MahsunSports: "androstreamlivetb7"},
	"tabiispor8":      {MahsunSports: "androstreamlivetb8"},
}

// ChannelList contains all parsed channels from anthology's canli.m3u.
var ChannelList []ChannelMapping
var channelMapByKey = make(map[string]ChannelMapping)
var channelMapByID = make(map[int]ChannelMapping)
var modelsOnce sync.Once

func cleanKey(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// InitChannels loads all 137+ channels from anthology's canli.m3u.
func InitChannels() {
	modelsOnce.Do(func() {
		content := EmbeddedCanliM3U

		// Check if anthology canli.m3u is available on disk
		if wd, err := os.Getwd(); err == nil {
			candDisk := filepath.Join(filepath.Dir(wd), "anthology", "providers", "M3U", "Liste", "canli.m3u")
			if diskBytes, err := os.ReadFile(candDisk); err == nil && len(diskBytes) > 0 {
				content = diskBytes
			}
		}

		lines := strings.Split(string(content), "\n")
		var currentINF string
		streamCounter := 100

		nameRe := regexp.MustCompile(`,([^,]+)$`)
		groupRe := regexp.MustCompile(`group-title="([^"]+)"`)
		logoRe := regexp.MustCompile(`tvg-logo="([^"]+)"`)
		idRe := regexp.MustCompile(`tvg-id="([^"]+)"`)
		backupRe := regexp.MustCompile(`tvg-backup="([^"]+)"`)

		for _, rawLine := range lines {
			line := strings.TrimSpace(rawLine)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "#EXTINF") {
				currentINF = line
			} else if !strings.HasPrefix(line, "#") && currentINF != "" {
				streamCounter++
				directURL := line

				// Parse attributes
				name := "Kanal"
				if m := nameRe.FindStringSubmatch(currentINF); len(m) > 1 {
					name = strings.TrimSpace(m[1])
				}

				category := "Genel"
				if m := groupRe.FindStringSubmatch(currentINF); len(m) > 1 {
					category = strings.TrimSpace(m[1])
					category = strings.TrimPrefix(category, "🇹🇷 ")
				}

				logo := ""
				if m := logoRe.FindStringSubmatch(currentINF); len(m) > 1 {
					logo = strings.TrimSpace(m[1])
				}

				tvgID := ""
				if m := idRe.FindStringSubmatch(currentINF); len(m) > 1 {
					tvgID = strings.TrimSpace(m[1])
				}

				backupURL := ""
				if m := backupRe.FindStringSubmatch(currentINF); len(m) > 1 {
					backupURL = strings.TrimSpace(m[1])
				}

				// Build normalized key
				rawKey := strings.TrimSuffix(tvgID, ".tr")
				if rawKey == "" {
					rawKey = name
				}
				key := cleanKey(rawKey)

				ch := ChannelMapping{
					Key:       key,
					Name:      name,
					StreamID:  streamCounter,
					Category:  category,
					LogoURL:   logo,
					DirectURL: directURL,
					BackupURL: backupURL,
				}

				// Check sport mapping
				if sm, ok := findSportPlatform(key); ok {
					ch.Taraftarium = sm.Taraftarium
					ch.ArdaSports = sm.ArdaSports
					ch.AlemdarSport = sm.AlemdarSport
					ch.MahsunSports = sm.MahsunSports
					ch.PapazSports = sm.PapazSports
				}

				ChannelList = append(ChannelList, ch)
				channelMapByKey[ch.Key] = ch
				channelMapByID[ch.StreamID] = ch

				currentINF = ""
			}
		}
	})
}

func findSportPlatform(key string) (SportMapping, bool) {
	if sm, ok := SportPlatformMap[key]; ok {
		return sm, true
	}
	// Try substring match for sports
	for k, sm := range SportPlatformMap {
		if strings.Contains(key, k) || strings.Contains(k, key) {
			return sm, true
		}
	}
	return SportMapping{}, false
}

func init() {
	InitChannels()
}

// FindChannelByKey finds a channel mapping by its key.
func FindChannelByKey(key string) (ChannelMapping, bool) {
	InitChannels()
	key = cleanKey(key)
	ch, ok := channelMapByKey[key]
	if ok {
		return ch, true
	}
	// Fallback fuzzy key matching
	for k, c := range channelMapByKey {
		if strings.Contains(k, key) || strings.Contains(key, k) {
			return c, true
		}
	}
	return ChannelMapping{}, false
}

// Provider offsets for virtual StreamIDs in Xtream Codes API
const (
	OffsetAlemdar    = 10000
	OffsetArda       = 20000
	OffsetTaraftarium = 30000
	OffsetMahsun     = 40000
	OffsetPapaz      = 50000
	OffsetDirect     = 60000
)

// ProviderFromVirtualID returns the base channel ID and the specific provider name.
func ProviderFromVirtualID(virtualID int) (baseID int, provider string) {
	if virtualID >= OffsetDirect {
		return virtualID - OffsetDirect, "direct"
	}
	if virtualID >= OffsetPapaz {
		return virtualID - OffsetPapaz, "papazsports"
	}
	if virtualID >= OffsetMahsun {
		return virtualID - OffsetMahsun, "mahsunsports"
	}
	if virtualID >= OffsetTaraftarium {
		return virtualID - OffsetTaraftarium, "taraftarium"
	}
	if virtualID >= OffsetArda {
		return virtualID - OffsetArda, "ardasports"
	}
	if virtualID >= OffsetAlemdar {
		return virtualID - OffsetAlemdar, "alemdarsport"
	}
	return virtualID, ""
}

// VirtualIDForProvider returns the virtual StreamID for a channel and provider.
func VirtualIDForProvider(baseID int, provider string) int {
	switch strings.ToLower(provider) {
	case "alemdarsport", "alemdar":
		return OffsetAlemdar + baseID
	case "ardasports", "arda", "selcuk":
		return OffsetArda + baseID
	case "taraftarium", "taraftarium24":
		return OffsetTaraftarium + baseID
	case "mahsunsports", "mahsun":
		return OffsetMahsun + baseID
	case "papazsports", "papaz":
		return OffsetPapaz + baseID
	case "direct":
		return OffsetDirect + baseID
	default:
		return baseID
	}
}

// FindChannelByID finds a channel mapping by its Xtream Codes StreamID (including virtual IDs).
func FindChannelByID(id int) (ChannelMapping, bool) {
	InitChannels()
	baseID, _ := ProviderFromVirtualID(id)
	ch, ok := channelMapByID[baseID]
	return ch, ok
}
