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

func TestFlightManager_IniChangeDoesNotBroadcast(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	err := os.WriteFile(iniPath, []byte("[Pilot]\nCallsign=Viper\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write ini file: %v", err)
	}

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath, 30*time.Second)

	// Listen for broadcast messages on hub.broadcast in a non-blocking way
	fm.HandleFileChange(iniPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast after ini change, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no message sent
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to be nil before briefing is saved")
	}
}

func TestFlightManager_BriefingWithinThresholdCorrelatesAndBroadcasts(t *testing.T) {
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
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath, 5*time.Second)

	// Step 1: ini changed
	fm.HandleFileChange(iniPath)

	// Step 2: briefing changed shortly after
	time.Sleep(10 * time.Millisecond)
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

func TestFlightManager_BriefingExceedingThresholdDoesNotCorrelate(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	_ = os.WriteFile(iniPath, []byte("ini data"), 0644)
	_ = os.WriteFile(briefingPath, []byte("briefing data"), 0644)

	hub := NewWebsocketHub()
	// Set threshold very short: 10ms
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath, 10*time.Millisecond)

	fm.HandleFileChange(iniPath)

	// Wait longer than threshold
	time.Sleep(30 * time.Millisecond)

	fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast when threshold is exceeded, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no message sent
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to remain nil when threshold exceeded")
	}
}

func TestFlightManager_BriefingWithoutIniDoesNotBroadcast(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	_ = os.WriteFile(briefingPath, []byte("briefing data"), 0644)

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath, 30*time.Second)

	fm.HandleFileChange(briefingPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast when no ini save recorded, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no message sent
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to remain nil")
	}
}

func TestFlightManager_IniAfterBriefingWithinThresholdCorrelatesAndBroadcasts(t *testing.T) {
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
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath, 5*time.Second)

	// Step 1: briefing changed first
	fm.HandleFileChange(briefingPath)

	// Step 2: ini changed shortly after
	time.Sleep(10 * time.Millisecond)
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

func TestFlightManager_IniAfterBriefingExceedingThresholdDoesNotCorrelate(t *testing.T) {
	tempDir := t.TempDir()
	iniPath := filepath.Join(tempDir, "Viper.ini")
	briefingPath := filepath.Join(tempDir, "briefing.txt")

	_ = os.WriteFile(iniPath, []byte("ini data"), 0644)
	_ = os.WriteFile(briefingPath, []byte("briefing data"), 0644)

	hub := NewWebsocketHub()
	// Set threshold very short: 10ms
	fm := NewFlightManager(hub, "Viper", iniPath, briefingPath, 10*time.Millisecond)

	fm.HandleFileChange(briefingPath)

	// Wait longer than threshold
	time.Sleep(30 * time.Millisecond)

	fm.HandleFileChange(iniPath)

	select {
	case msg := <-hub.broadcast:
		t.Fatalf("expected no broadcast when threshold is exceeded, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected: no message sent
	}

	if fm.GetCurrentFlight() != nil {
		t.Fatalf("expected current flight to remain nil when threshold exceeded")
	}
}

func TestWebserver_FlightEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ws := NewWebserver(":0")

	hub := NewWebsocketHub()
	fm := NewFlightManager(hub, "Viper", "", "", 30*time.Second)
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
