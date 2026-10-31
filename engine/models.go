package engine

// ChannelMapping defines how a channel is identified across different providers.
type ChannelMapping struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	StreamID     int    `json:"stream_id"`
	Category     string `json:"category"`
	LogoURL      string `json:"logo_url"`
	Taraftarium  string `json:"taraftarium,omitempty"`
	ArdaSports   string `json:"ardasports,omitempty"`
	AlemdarSport string `json:"alemdarsport,omitempty"`
	MahsunSports string `json:"mahsunsports,omitempty"`
	PapazSports  string `json:"papazsports,omitempty"`
}

// StreamResult is the resolved upstream HLS URL and required HTTP headers.
type StreamResult struct {
	ChannelKey  string            `json:"channel_key"`
	ChannelName string            `json:"channel_name"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Provider    string            `json:"provider"`
}

const (
	AssetBaseURL = "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/"
)

// ChannelList is the ordered catalog of supported live channels.
var ChannelList = []ChannelMapping{
	{
		Key: "beinsports1", Name: "BEIN SPORTS 1 HD", StreamID: 101, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-1",
		ArdaSports:   "bein-sports-1",
		AlemdarSport: "bein1",
		MahsunSports: "androstreamlivebs1",
		PapazSports:  "100001",
	},
	{
		Key: "beinsports2", Name: "BEIN SPORTS 2 HD", StreamID: 102, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-2",
		ArdaSports:   "bein-sports-2",
		AlemdarSport: "bein2",
		MahsunSports: "androstreamlivebs2",
		PapazSports:  "100002",
	},
	{
		Key: "beinsports3", Name: "BEIN SPORTS 3 HD", StreamID: 103, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-3",
		ArdaSports:   "bein-sports-3",
		AlemdarSport: "bein3",
		MahsunSports: "androstreamlivebs3",
		PapazSports:  "100003",
	},
	{
		Key: "beinsports4", Name: "BEIN SPORTS 4 HD", StreamID: 104, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-4",
		ArdaSports:   "bein-sports-4",
		AlemdarSport: "bein4",
		MahsunSports: "androstreamlivebs4",
		PapazSports:  "100004",
	},
	{
		Key: "beinsports5", Name: "BEIN SPORTS 5 HD", StreamID: 105, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-5",
		ArdaSports:   "bein-sports-5",
		AlemdarSport: "bein5",
		MahsunSports: "androstreamlivebs5",
		PapazSports:  "100005",
	},
	{
		Key: "beinsportsmax1", Name: "BEIN SPORTS MAX 1 HD", StreamID: 106, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-max-1",
		ArdaSports:   "bein-sports-max-1",
		AlemdarSport: "beinmax1",
		MahsunSports: "androstreamlivebsm1",
		PapazSports:  "100006",
	},
	{
		Key: "beinsportsmax2", Name: "BEIN SPORTS MAX 2 HD", StreamID: 107, Category: "Spor",
		LogoURL:      AssetBaseURL + "beinsports.png",
		Taraftarium:  "bein-sports-max-2",
		ArdaSports:   "bein-sports-max-2",
		AlemdarSport: "beinmax2",
		MahsunSports: "androstreamlivebsm2",
		PapazSports:  "100007",
	},
	{
		Key: "ssport1", Name: "S SPORT 1 HD", StreamID: 108, Category: "Spor",
		LogoURL:      AssetBaseURL + "ssport.png",
		Taraftarium:  "s-sport",
		ArdaSports:   "ssport",
		AlemdarSport: "ssport1",
		MahsunSports: "androstreamlivessport1",
		PapazSports:  "100008",
	},
	{
		Key: "ssport2", Name: "S SPORT 2 HD", StreamID: 109, Category: "Spor",
		LogoURL:      AssetBaseURL + "ssport.png",
		Taraftarium:  "s-sport-2",
		ArdaSports:   "ssport2",
		AlemdarSport: "ssport2",
		MahsunSports: "androstreamlivessport2",
		PapazSports:  "100009",
	},
	{
		Key: "ssportplus", Name: "S SPORT PLUS", StreamID: 110, Category: "Spor",
		LogoURL:      AssetBaseURL + "ssport.png",
		Taraftarium:  "s-sport-plus",
		ArdaSports:   "ssportplus",
		MahsunSports: "androstreamlivessportplus",
	},
	{
		Key: "tivibuspor1", Name: "TIVIBU SPOR 1 HD", StreamID: 111, Category: "Spor",
		LogoURL:      AssetBaseURL + "tivibuspor.png",
		Taraftarium:  "tivibu-spor-1",
		ArdaSports:   "tivibu-spor-1",
		AlemdarSport: "tivibu1",
		MahsunSports: "androstreamlivetivibuspor1",
		PapazSports:  "100010",
	},
	{
		Key: "tivibuspor2", Name: "TIVIBU SPOR 2 HD", StreamID: 112, Category: "Spor",
		LogoURL:      AssetBaseURL + "tivibuspor.png",
		Taraftarium:  "tivibu-spor-2",
		ArdaSports:   "tivibu-spor-2",
		AlemdarSport: "tivibu2",
		MahsunSports: "androstreamlivetivibuspor2",
		PapazSports:  "100011",
	},
	{
		Key: "tivibuspor3", Name: "TIVIBU SPOR 3 HD", StreamID: 113, Category: "Spor",
		LogoURL:      AssetBaseURL + "tivibuspor.png",
		Taraftarium:  "tivibu-spor-3",
		ArdaSports:   "tivibu-spor-3",
		AlemdarSport: "tivibu3",
		MahsunSports: "androstreamlivetivibuspor3",
		PapazSports:  "100012",
	},
	{
		Key: "tivibuspor4", Name: "TIVIBU SPOR 4 HD", StreamID: 114, Category: "Spor",
		LogoURL:      AssetBaseURL + "tivibuspor.png",
		Taraftarium:  "tivibu-spor-4",
		ArdaSports:   "tivibu-spor-4",
		MahsunSports: "androstreamlivetivibuspor4",
	},
	{
		Key: "smartspor1", Name: "SMART SPOR 1 HD", StreamID: 115, Category: "Spor",
		LogoURL:      AssetBaseURL + "smartspor.png",
		Taraftarium:  "smart-spor",
		ArdaSports:   "smart-spor",
		AlemdarSport: "smart1",
		MahsunSports: "androstreamlivesmartspor1",
		PapazSports:  "100013",
	},
	{
		Key: "smartspor2", Name: "SMART SPOR 2 HD", StreamID: 116, Category: "Spor",
		LogoURL:      AssetBaseURL + "smartspor.png",
		Taraftarium:  "smart-spor-2",
		ArdaSports:   "smart-spor-2",
		AlemdarSport: "smart2",
		MahsunSports: "androstreamlivesmartspor2",
		PapazSports:  "100014",
	},
	{
		Key: "trtspor", Name: "TRT SPOR HD", StreamID: 117, Category: "Spor",
		LogoURL:      AssetBaseURL + "trtspor.png",
		Taraftarium:  "trt-spor",
		ArdaSports:   "trtspor",
		AlemdarSport: "trtspor",
		MahsunSports: "androstreamlivetrtspor",
		PapazSports:  "100015",
	},
	{
		Key: "trtsporyildiz", Name: "TRT SPOR YILDIZ", StreamID: 118, Category: "Spor",
		LogoURL:      AssetBaseURL + "trtsporyildiz.png",
		Taraftarium:  "trt-spor-yildiz",
		ArdaSports:   "trtsporyildiz",
		AlemdarSport: "trtyildiz",
		MahsunSports: "androstreamlivetrtsporyildiz",
		PapazSports:  "100016",
	},
	{
		Key: "trt1", Name: "TRT 1 HD", StreamID: 119, Category: "Ulusal",
		LogoURL:      AssetBaseURL + "trt1.png",
		Taraftarium:  "trt-1",
		ArdaSports:   "trt1",
		AlemdarSport: "trt1",
		MahsunSports: "androstreamlivetrt1",
		PapazSports:  "100017",
	},
	{
		Key: "aspor", Name: "A SPOR HD", StreamID: 120, Category: "Spor",
		LogoURL:      AssetBaseURL + "aspor.png",
		Taraftarium:  "a-spor",
		ArdaSports:   "aspor",
		AlemdarSport: "aspor",
		MahsunSports: "androstreamliveaspor",
		PapazSports:  "100018",
	},
	{
		Key: "tv85", Name: "TV 8.5 HD", StreamID: 121, Category: "Spor",
		LogoURL:      AssetBaseURL + "tv85.png",
		Taraftarium:  "tv-8-5",
		ArdaSports:   "tv85",
		AlemdarSport: "tv85",
		MahsunSports: "androstreamlivetv85",
		PapazSports:  "100019",
	},
	{
		Key: "nbatv", Name: "NBA TV HD", StreamID: 122, Category: "Spor",
		LogoURL:      AssetBaseURL + "nbatv.png",
		Taraftarium:  "nba-tv",
		ArdaSports:   "nbatv",
		MahsunSports: "androstreamlivenbatv",
	},
	{
		Key: "eurosport1", Name: "EUROSPORT 1 HD", StreamID: 123, Category: "Spor",
		LogoURL:      AssetBaseURL + "eurosport.png",
		Taraftarium:  "eurosport-1",
		ArdaSports:   "eurosport1",
		AlemdarSport: "eurosport1",
		MahsunSports: "androstreamliveeurosport1",
		PapazSports:  "100020",
	},
	{
		Key: "eurosport2", Name: "EUROSPORT 2 HD", StreamID: 124, Category: "Spor",
		LogoURL:      AssetBaseURL + "eurosport.png",
		Taraftarium:  "eurosport-2",
		ArdaSports:   "eurosport2",
		AlemdarSport: "eurosport2",
		MahsunSports: "androstreamliveeurosport2",
		PapazSports:  "100021",
	},
	{
		Key: "cbcsport", Name: "CBC SPORT HD", StreamID: 125, Category: "Spor",
		LogoURL:      AssetBaseURL + "cbcsport.png",
		Taraftarium:  "cbc-sport",
		ArdaSports:   "cbcsport",
		AlemdarSport: "cbcsport",
		MahsunSports: "androstreamlivecbcsport",
	},
	{
		Key: "idmantv", Name: "İDMAN TV HD", StreamID: 126, Category: "Spor",
		LogoURL:      AssetBaseURL + "idmantv.png",
		Taraftarium:  "idman-tv",
		ArdaSports:   "idmantv",
		AlemdarSport: "idmantv",
		MahsunSports: "androstreamliveidmantv",
	},
	{
		Key: "exxentv", Name: "EXXEN SPOR 1 HD", StreamID: 127, Category: "Spor",
		LogoURL:      AssetBaseURL + "exxen.png",
		Taraftarium:  "exxen-spor-1",
		ArdaSports:   "exxenspor1",
		AlemdarSport: "exxen1",
		MahsunSports: "androstreamliveexxen1",
		PapazSports:  "100022",
	},
}

var channelMapByKey = make(map[string]ChannelMapping)
var channelMapByID = make(map[int]ChannelMapping)

func init() {
	for _, ch := range ChannelList {
		channelMapByKey[ch.Key] = ch
		channelMapByID[ch.StreamID] = ch
	}
}

// FindChannelByKey finds a channel mapping by its key or aliases.
func FindChannelByKey(key string) (ChannelMapping, bool) {
	ch, ok := channelMapByKey[key]
	return ch, ok
}

// FindChannelByID finds a channel mapping by its Xtream Codes StreamID.
func FindChannelByID(id int) (ChannelMapping, bool) {
	ch, ok := channelMapByID[id]
	return ch, ok
}
