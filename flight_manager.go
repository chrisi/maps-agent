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

// FlightManager manages flight data collection from ini and briefing files,
// correlates their updates based on a time threshold, and broadcasts WebSocket update events.
type FlightManager struct {
	logger              *util.Logger
	hub                 *WebsocketHub
	callsign            string
	iniPath             string
	briefingPath        string
	threshold           time.Duration
	lastIniSave         time.Time
	lastIniContent      string
	lastBriefingSave    time.Time
	lastBriefingContent string
	currentFlight       *FlightData
	mu                  sync.RWMutex
}

// NewFlightManager creates a new FlightManager instance.
func NewFlightManager(hub *WebsocketHub, callsign, iniPath, briefingPath string, threshold time.Duration) *FlightManager {
	if threshold <= 0 {
		threshold = 30 * time.Second
	}
	return &FlightManager{
		logger:       util.NewLogger("FlightManager", os.Stdout, util.Info, true),
		hub:          hub,
		callsign:     callsign,
		iniPath:      iniPath,
		briefingPath: briefingPath,
		threshold:    threshold,
	}
}

// SetThreshold updates the time threshold for correlating ini and briefing file changes.
func (fm *FlightManager) SetThreshold(threshold time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.threshold = threshold
}

// GetThreshold returns the current correlation threshold.
func (fm *FlightManager) GetThreshold() time.Duration {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.threshold
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

// HandleIniChange records the modification of the callsign.ini file.
// If briefing.txt was saved within the threshold, it correlates both files, updates current flight data, and broadcasts a WebSocket update event.
func (fm *FlightManager) HandleIniChange(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fm.logger.Errorf("failed to read ini file %s: %v", filePath, err)
		return
	}

	iniContent := string(content)

	fm.mu.Lock()
	fm.lastIniSave = time.Now()
	fm.lastIniContent = iniContent
	iniSaveTime := fm.lastIniSave

	lastBriefingSave := fm.lastBriefingSave
	lastBriefingContent := fm.lastBriefingContent
	threshold := fm.threshold
	fm.mu.Unlock()

	if lastBriefingSave.IsZero() {
		fm.logger.Infof("Recorded %s change at %v (waiting for briefing.txt within %v)", filepath.Base(filePath), iniSaveTime, threshold)
		return
	}

	elapsed := time.Since(lastBriefingSave)
	if elapsed > threshold {
		fm.logger.Infof("Recorded %s change at %v (briefing.txt was saved %v ago, threshold: %v; waiting for new briefing.txt)",
			filepath.Base(filePath), iniSaveTime, elapsed.Round(time.Millisecond), threshold)
		return
	}

	if lastBriefingContent == "" && fm.briefingPath != "" {
		if briefingBytes, err := os.ReadFile(fm.briefingPath); err == nil {
			lastBriefingContent = string(briefingBytes)
		}
	}

	flight := &FlightData{
		Callsign:  fm.callsign,
		Ini:       iniContent,
		Briefing:  lastBriefingContent,
		UpdatedAt: time.Now(),
	}

	fm.mu.Lock()
	fm.currentFlight = flight
	fm.mu.Unlock()

	fm.logger.Infof("Correlated flight files (briefing.txt was saved %v ago). Broadcasting mission ready update.",
		elapsed.Round(time.Millisecond))

	fm.BroadcastMissionReady()
}

// HandleBriefingChange records the modification of the briefing.txt file.
// If callsign.ini was saved within the threshold, it correlates both files, updates current flight data, and broadcasts a WebSocket update event.
func (fm *FlightManager) HandleBriefingChange(filePath string) {
	briefingBytes, err := os.ReadFile(filePath)
	if err != nil {
		fm.logger.Errorf("failed to read briefing file %s: %v", filePath, err)
		return
	}

	briefingContent := string(briefingBytes)

	fm.mu.Lock()
	fm.lastBriefingSave = time.Now()
	fm.lastBriefingContent = briefingContent
	briefingSaveTime := fm.lastBriefingSave

	lastIniSave := fm.lastIniSave
	lastIniContent := fm.lastIniContent
	threshold := fm.threshold
	fm.mu.Unlock()

	if lastIniSave.IsZero() {
		fm.logger.Infof("Recorded %s change at %v (waiting for %s.ini within %v)", filepath.Base(filePath), briefingSaveTime, fm.callsign, threshold)
		return
	}

	elapsed := time.Since(lastIniSave)
	if elapsed > threshold {
		fm.logger.Infof("Recorded %s change at %v (%s.ini was saved %v ago, threshold: %v; waiting for new %s.ini)",
			filepath.Base(filePath), briefingSaveTime, fm.callsign, elapsed.Round(time.Millisecond), threshold, fm.callsign)
		return
	}

	if lastIniContent == "" && fm.iniPath != "" {
		if iniBytes, err := os.ReadFile(fm.iniPath); err == nil {
			lastIniContent = string(iniBytes)
		}
	}

	flight := &FlightData{
		Callsign:  fm.callsign,
		Ini:       lastIniContent,
		Briefing:  briefingContent,
		UpdatedAt: time.Now(),
	}

	fm.mu.Lock()
	fm.currentFlight = flight
	fm.mu.Unlock()

	fm.logger.Infof("Correlated flight files (%s.ini was saved %v ago). Broadcasting mission ready update.",
		fm.callsign, elapsed.Round(time.Millisecond))

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
