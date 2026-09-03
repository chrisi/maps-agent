package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestFlightManager_IniFirstThenBriefing_CorrelatesAndBroadcasts(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	iniContent := "[Pilot]\nCallsign=Viper\n"
	briefingContent := "MISSION BRIEFING:\nTarget: Airbase\n"

	err := os.WriteFile(iniPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("failed to write ini file: %v", err)
	}
	err = os.WriteFile(briefingPath, []byte(briefingContent), 0644)
	if err != nil {
		t.Fatalf("failed to write briefing file: %v", err)
	}

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath)

	// Step 1: INI saved first when nothing was saved before -> no event, waiting for briefing
	fm.HandleFileChange(iniPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast after initial ini change, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no event sent
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to be nil before briefing is saved")
	}

	// Step 2: Briefing saved when INI is already present -> event sent, flight data populated
	go fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		var agentMsg AgentMessage
		if err := json.Unmarshal(msg, &agentMsg); err != nil {
			t.Fatalf("failed to unmarshal broadcast message: %v", err)
		}
		if agentMsg.Type != "update" {
			t.Errorf("expected msg type 'update', got '%s'", agentMsg.Type)
		}
		payloadBytes, err := json.Marshal(agentMsg.Payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		var readyPayload MissionReadyPayload
		if err := json.Unmarshal(payloadBytes, &readyPayload); err != nil {
			t.Fatalf("failed to unmarshal ready payload: %v", err)
		}
		if readyPayload.Status != "ready" || readyPayload.Endpoint != "/flight" {
			t.Errorf("unexpected payload: %+v", readyPayload)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for broadcast message")
	}

	flight := fm.GetCurrentFlight()
	if flight == nil {
		t.Fatalf("expected current flight data to be set")
	}
	if flight.Callsign != "Viper" {
		t.Errorf("expected callsign 'Viper', got '%s'", flight.Callsign)
	}
	if flight.Ini != iniContent {
		t.Errorf("expected ini content '%s', got '%s'", iniContent, flight.Ini)
	}
	if flight.Briefing != briefingContent {
		t.Errorf("expected briefing content '%s', got '%s'", briefingContent, flight.Briefing)
	}
}

func TestFlightManager_BriefingFirstThenIni_CorrelatesAndBroadcasts(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	iniContent := "[Pilot]\nCallsign=Viper\n"
	briefingContent := "MISSION BRIEFING:\nTarget: Airbase\n"

	err := os.WriteFile(iniPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("failed to write ini file: %v", err)
	}
	err = os.WriteFile(briefingPath, []byte(briefingContent), 0644)
	if err != nil {
		t.Fatalf("failed to write briefing file: %v", err)
	}

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath)

	// Step 1: Briefing saved when no INI is present -> no event, waiting for INI
	fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast after initial briefing change without INI, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no event sent
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to be nil before INI is saved")
	}

	// Step 2: INI saved when Briefing is present -> event sent, flight data populated
	go fm.HandleFileChange(iniPath)

	select {
	case msg := <-hub.broadcast:
		var agentMsg AgentMessage
		if err := json.Unmarshal(msg, &agentMsg); err != nil {
			t.Fatalf("failed to unmarshal broadcast message: %v", err)
		}
		if agentMsg.Type != "update" {
			t.Errorf("expected msg type 'update', got '%s'", agentMsg.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for broadcast message")
	}

	flight := fm.GetCurrentFlight()
	if flight == nil {
		t.Fatalf("expected current flight data to be set")
	}
	if flight.Ini != iniContent || flight.Briefing != briefingContent {
		t.Errorf("unexpected flight content: %+v", flight)
	}
}

func TestFlightManager_IniRepeatedWithBriefingPresent_IsIdempotentAndBroadcasts(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	iniContent1 := "[Pilot]\nCallsign=Viper\nLoadout=1\n"
	iniContent2 := "[Pilot]\nCallsign=Viper\nLoadout=2\n"
	briefingContent := "MISSION BRIEFING:\nTarget: Airbase\n"

	_ = os.WriteFile(iniPath, []byte(iniContent1), 0644)
	_ = os.WriteFile(briefingPath, []byte(briefingContent), 0644)

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath)

	// Step 1: Briefing saved first
	fm.HandleFileChange(briefingPath)

	// Step 2: INI 1 saved -> event 1
	go fm.HandleFileChange(iniPath)
	select {
	case <-hub.broadcast:
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for first broadcast")
	}
	if fm.GetCurrentFlight().Ini != iniContent1 {
		t.Errorf("expected ini content 1, got %s", fm.GetCurrentFlight().Ini)
	}

	// Step 3: INI 2 saved (repeated) -> event 2
	_ = os.WriteFile(iniPath, []byte(iniContent2), 0644)
	go fm.HandleFileChange(iniPath)
	select {
	case <-hub.broadcast:
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for second broadcast")
	}
	if fm.GetCurrentFlight().Ini != iniContent2 {
		t.Errorf("expected ini content 2, got %s", fm.GetCurrentFlight().Ini)
	}
}

func TestFlightManager_BriefingIdentical_DoesNothing(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	iniContent := "[Pilot]\nCallsign=Viper\n"
	briefingContent := "MISSION BRIEFING:\nGenerated: 2026-09-03 12:00:00\nTarget: Airbase\n"

	_ = os.WriteFile(iniPath, []byte(iniContent), 0644)
	_ = os.WriteFile(briefingPath, []byte(briefingContent), 0644)

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath)

	// Initial correlation: Briefing + INI
	fm.HandleFileChange(briefingPath)
	go fm.HandleFileChange(iniPath)
	<-hub.broadcast

	// Now save the exact same briefing content again
	fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast when identical briefing is saved, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: nothing happens
	}

	// Flight data should still be intact
	flight := fm.GetCurrentFlight()
	if flight == nil || flight.Ini != iniContent {
		t.Fatalf("expected flight data to remain unchanged")
	}
}

func TestFlightManager_BriefingTimestampLineDifferenceOnly_DoesNothing(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	iniContent := "[Pilot]\nCallsign=Viper\n"
	briefingContent1 := "MISSION BRIEFING:\nGenerated: 2026-09-03 12:00:00\nTarget: Airbase\n"
	briefingContent2 := "MISSION BRIEFING:\nGenerated: 2026-09-03 12:05:30\nTarget: Airbase\n"

	_ = os.WriteFile(iniPath, []byte(iniContent), 0644)
	_ = os.WriteFile(briefingPath, []byte(briefingContent1), 0644)

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath)

	// Initial correlation: Briefing 1 + INI
	fm.HandleFileChange(briefingPath)
	go fm.HandleFileChange(iniPath)
	<-hub.broadcast

	// Now save briefing 2 which only differs in line 2 (timestamp)
	_ = os.WriteFile(briefingPath, []byte(briefingContent2), 0644)
	fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast when briefing only differs by timestamp on line 2, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: nothing happens (considered unchanged)
	}

	// Flight data should still be intact and not reset
	flight := fm.GetCurrentFlight()
	if flight == nil || flight.Ini != iniContent {
		t.Fatalf("expected flight data to remain intact when only timestamp changes")
	}
}

func TestIsBriefingEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical strings",
			a:        "Header\nTime: 1234\nLine 3\nLine 4",
			b:        "Header\nTime: 1234\nLine 3\nLine 4",
			expected: true,
		},
		{
			name:     "different timestamp on line 2",
			a:        "Header\nGenerated: 2026-01-01 10:00\nLine 3\nLine 4",
			b:        "Header\nGenerated: 2026-09-03 15:30\nLine 3\nLine 4",
			expected: true,
		},
		{
			name:     "different timestamp on line 2 with CRLF vs LF",
			a:        "Header\r\nGenerated: 2026-01-01 10:00\r\nLine 3\r\nLine 4",
			b:        "Header\nGenerated: 2026-09-03 15:30\nLine 3\nLine 4",
			expected: true,
		},
		{
			name:     "different line 1 (header)",
			a:        "Header A\nTime: 1234\nLine 3",
			b:        "Header B\nTime: 1234\nLine 3",
			expected: false,
		},
		{
			name:     "different line 3 (content)",
			a:        "Header\nTime: 1234\nLine 3 A",
			b:        "Header\nTime: 1234\nLine 3 B",
			expected: false,
		},
		{
			name:     "different line counts",
			a:        "Header\nTime: 1234\nLine 3",
			b:        "Header\nTime: 1234\nLine 3\nLine 4",
			expected: false,
		},
		{
			name:     "single line identical",
			a:        "Header only",
			b:        "Header only",
			expected: true,
		},
		{
			name:     "single line different",
			a:        "Header 1",
			b:        "Header 2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isBriefingEqual(tt.a, tt.b)
			if actual != tt.expected {
				t.Errorf("isBriefingEqual() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestFlightManager_BriefingChanged_ClearsIniAndDoesNotBroadcast(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	iniContent1 := "[Pilot]\nCallsign=Viper\nMission=1\n"
	briefingContent1 := "MISSION BRIEFING 1:\nTarget: Airbase\n"

	iniContent2 := "[Pilot]\nCallsign=Viper\nMission=2\n"
	briefingContent2 := "MISSION BRIEFING 2:\nTarget: Bridge\n"

	_ = os.WriteFile(iniPath, []byte(iniContent1), 0644)
	_ = os.WriteFile(briefingPath, []byte(briefingContent1), 0644)

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath)

	// Step 1: Correlate initial mission (Briefing 1 + INI 1)
	fm.HandleFileChange(briefingPath)
	go fm.HandleFileChange(iniPath)
	<-hub.broadcast
	if fm.GetCurrentFlight() == nil || fm.GetCurrentFlight().Briefing != briefingContent1 {
		t.Fatalf("expected flight with briefing 1")
	}

	// Step 2: Save Briefing 2 (different content) -> INI data cleared, current flight nil, no event
	_ = os.WriteFile(briefingPath, []byte(briefingContent2), 0644)
	fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast when briefing content changes, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no event
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to be nil after briefing changed and INI was cleared")
	}

	// Step 3: Save INI 2 -> now correlates with Briefing 2 and sends event
	_ = os.WriteFile(iniPath, []byte(iniContent2), 0644)
	go fm.HandleFileChange(iniPath)

	select {
	case <-hub.broadcast:
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for broadcast after new INI saved")
	}

	flight := fm.GetCurrentFlight()
	if flight == nil {
		t.Fatalf("expected flight data to be set")
	}
	if flight.Briefing != briefingContent2 || flight.Ini != iniContent2 {
		t.Errorf("expected briefing 2 and ini 2, got: %+v", flight)
	}
}

func TestWebserver_FlightEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ws := NewWebserver(":0")

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", "", "")
	ws.RegisterFlightEndpoint(fm)

	// 1. When no flight data is present -> 404
	req404, _ := http.NewRequest("GET", "/flight", nil)
	w404 := httptest.NewRecorder()
	ws.r.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w404.Code)
	}

	// 2. Set flight data -> 200 OK with JSON
	fm.SetCurrentFlight(&FlightData{
		Callsign:  "Viper",
		Ini:       "[Pilot]\nCallsign=Viper\n",
		Briefing:  "Target: Objective Alpha",
		UpdatedAt: time.Now(),
	})

	req200, _ := http.NewRequest("GET", "/flight", nil)
	w200 := httptest.NewRecorder()
	ws.r.ServeHTTP(w200, req200)
	if w200.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w200.Code)
	}

	var resp FlightData
	if err := json.Unmarshal(w200.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Callsign != "Viper" || resp.Briefing != "Target: Objective Alpha" {
		t.Errorf("unexpected response content: %+v", resp)
	}
}
