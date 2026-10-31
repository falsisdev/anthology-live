package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestXtreamAPI(t *testing.T) {
	// 1. Test Auth info
	req := httptest.NewRequest("GET", "/player_api.php?username=test&password=test", nil)
	w := httptest.NewRecorder()
	HandleXtreamAPI(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var authRes XtreamAuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &authRes); err != nil {
		t.Fatalf("Failed to parse auth json: %v", err)
	}
	if authRes.UserInfo.Status != "Active" {
		t.Errorf("Expected Active status, got %s", authRes.UserInfo.Status)
	}

	// 2. Test Live Streams catalog
	reqStreams := httptest.NewRequest("GET", "/player_api.php?username=test&password=test&action=get_live_streams", nil)
	wStreams := httptest.NewRecorder()
	HandleXtreamAPI(wStreams, reqStreams)

	if wStreams.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", wStreams.Code)
	}

	var streams []XtreamLiveStream
	if err := json.Unmarshal(wStreams.Body.Bytes(), &streams); err != nil {
		t.Fatalf("Failed to parse streams json: %v", err)
	}
	if len(streams) == 0 {
		t.Errorf("Expected stream items, got 0")
	}

	// 3. Test M3U Playlist
	reqM3U := httptest.NewRequest("GET", "/iptv.m3u?username=test&password=test", nil)
	wM3U := httptest.NewRecorder()
	HandleM3UPlaylist(wM3U, reqM3U)

	if wM3U.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", wM3U.Code)
	}

	m3uContent := wM3U.Body.String()
	if !strings.HasPrefix(m3uContent, "#EXTM3U") {
		t.Errorf("Expected #EXTM3U header, got: %s", m3uContent[:30])
	}
	if !strings.Contains(m3uContent, "BEIN SPORTS 1 HD") {
		t.Errorf("Expected BEIN SPORTS 1 HD in playlist")
	}
}
