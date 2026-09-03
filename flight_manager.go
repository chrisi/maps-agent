package main

import (
	"encoding/json"
	"maps-agent/util"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FlightData holds the correlated flight configuration and briefing data.
type FlightData struct {
	Callsign  string    `json:"callsign,omitempty"`
	Ini       string    `json:"ini"`
	Briefing  string    `json:"briefing"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MissionReadyPayload represents the information payload sent via WebSocket
// notifying the client that current mission data is ready for download.
type MissionReadyPayload struct {
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
	Message  string `json:"message"`
}

// FlightManager manages flight data collection from ini and briefing files
// and broadcasts WebSocket update events according to the mission state machine.
type FlightManager struct {
	logger          *util.Logger
	hub             *WebsocketHub
	callsign        string
	iniPath         string
	briefingPath    string
	hasIni          bool
	iniContent      string
	hasBriefing     bool
	briefingContent string
	currentFlight   *FlightData
	mu              sync.RWMutex
}

// NewFlightManager creates a new FlightManager instance.
func NewFlightManager(hub *WebsocketHub, callsign, iniPath, briefingPath string) *FlightManager {
	return &FlightManager{
		logger:       util.NewLogger("FlightManager", os.Stdout, util.Info, true),
		hub:          hub,
		callsign:     callsign,
		iniPath:      iniPath,
		briefingPath: briefingPath,
	}
}

// GetCurrentFlight returns the currently stored flight data.
func (fm *FlightManager) GetCurrentFlight() *FlightData {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.currentFlight
}

// SetCurrentFlight explicitly sets the current flight data.
func (fm *FlightManager) SetCurrentFlight(data *FlightData) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.currentFlight = data
}

// HandleFileChange processes a changed file path and routes it to ini or briefing handlers.
func (fm *FlightManager) HandleFileChange(filePath string) {
	base := filepath.Base(filePath)
	lowerBase := strings.ToLower(base)

	callsignIni := strings.ToLower(fm.callsign + ".ini")
	if lowerBase == callsignIni || (fm.iniPath != "" && filepath.Clean(filePath) == filepath.Clean(fm.iniPath)) || (fm.callsign == "" && strings.HasSuffix(lowerBase, ".ini")) {
		fm.HandleIniChange(filePath)
		return
	}

	if lowerBase == "briefing.txt" || (fm.briefingPath != "" && filepath.Clean(filePath) == filepath.Clean(fm.briefingPath)) {
		fm.HandleBriefingChange(filePath)
		return
	}
}

// HandleIniChange processes saving of the callsign.ini file.
// If a briefing is already present, the INI is adopted, current flight updated, and an event is sent.
// If no briefing is present yet, the INI is stored and we wait for the briefing.
func (fm *FlightManager) HandleIniChange(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fm.logger.Errorf("failed to read ini file %s: %v", filePath, err)
		return
	}

	iniContent := string(content)

	fm.mu.Lock()
	fm.hasIni = true
	fm.iniContent = iniContent

	if !fm.hasBriefing {
		fm.mu.Unlock()
		fm.logger.Infof("Recorded %s change (waiting for briefing.txt)", filepath.Base(filePath))
		return
	}

	flight := &FlightData{
		Callsign:  fm.callsign,
		Ini:       fm.iniContent,
		Briefing:  fm.briefingContent,
		UpdatedAt: time.Now(),
	}
	fm.currentFlight = flight
	fm.mu.Unlock()

	fm.logger.Infof("Correlated flight files (%s.ini and briefing.txt). Broadcasting mission ready update.", fm.callsign)
	fm.BroadcastMissionReady()
}

// HandleBriefingChange processes saving of the briefing.txt file.
// If a briefing was already saved, the content is compared:
//   - If identical, nothing happens.
//   - If different, the new briefing is adopted and previously saved INI data is cleared.
//
// If no briefing was saved previously:
//   - If an INI is present, the briefing is adopted and an event is sent.
//   - If no INI is present, the briefing is adopted and we wait for the INI.
func (fm *FlightManager) HandleBriefingChange(filePath string) {
	briefingBytes, err := os.ReadFile(filePath)
	if err != nil {
		fm.logger.Errorf("failed to read briefing file %s: %v", filePath, err)
		return
	}

	briefingContent := string(briefingBytes)

	fm.mu.Lock()
	if fm.hasBriefing {
		if fm.briefingContent == briefingContent {
			fm.mu.Unlock()
			fm.logger.Infof("Briefing content unchanged; ignoring event")
			return
		}

		// Content differs: adopt new briefing and clear saved INI data
		fm.briefingContent = briefingContent
		fm.hasIni = false
		fm.iniContent = ""
		fm.currentFlight = nil
		fm.mu.Unlock()

		fm.logger.Infof("Briefing content changed; updated briefing and cleared INI data (waiting for %s.ini)", fm.callsign)
		return
	}

	// First time briefing is saved
	fm.hasBriefing = true
	fm.briefingContent = briefingContent

	if !fm.hasIni {
		fm.mu.Unlock()
		fm.logger.Infof("Recorded %s change (waiting for %s.ini)", filepath.Base(filePath), fm.callsign)
		return
	}

	flight := &FlightData{
		Callsign:  fm.callsign,
		Ini:       fm.iniContent,
		Briefing:  fm.briefingContent,
		UpdatedAt: time.Now(),
	}
	fm.currentFlight = flight
	fm.mu.Unlock()

	fm.logger.Infof("Correlated flight files (%s.ini and briefing.txt). Broadcasting mission ready update.", fm.callsign)
	fm.BroadcastMissionReady()
}

// BroadcastMissionReady sends a WebSocket update event notifying clients that the mission can be downloaded.
func (fm *FlightManager) BroadcastMissionReady() {
	if fm.hub == nil {
		return
	}

	msg := AgentMessage{
		Type: "update",
		Payload: MissionReadyPayload{
			Status:   "ready",
			Endpoint: "/flight",
			Message:  "Current mission is ready for download",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		fm.logger.Errorf("failed to serialize mission ready message: %v", err)
		return
	}

	fm.logger.Infof("Broadcasting mission ready event via websocket")
	fm.hub.broadcast <- data
}
