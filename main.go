package main

import (
	"context"
	"encoding/json"
	"maps-agent/camtac"
	"maps-agent/util"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := util.NewLogger("main", os.Stdout, util.Info, true)

	gin.SetMode(gin.ReleaseMode)

	cfg := parseConfig()

	theater := camtac.GetActiveTheater("Falcon BMS 4.38")
	logger.Infof("Active Theater: %s", theater)

	campaignPath := filepath.Join(cfg.FalconBase, "Data", string(theater), "Campaign")

	ws := NewWebserver(cfg.ServerAddr)

	hub := NewWebsocketHub()
	go hub.run()
	ws.RegisterWebsocketEndpoint(hub)

	iniPath := filepath.Join(cfg.FalconBase, `User\Config`, cfg.Callsign+".ini")
	briefingPath := filepath.Join(cfg.FalconBase, `User\Briefings`, "briefing.txt")

	flightManager := NewFlightManager(hub, cfg.Callsign, iniPath, briefingPath)
	ws.RegisterFlightEndpoint(flightManager)

	watcher, err := NewFileWatcher(cfg.FsCheckFreq, flightManager.HandleFileChange)
	if err == nil {
		watcher.Add(iniPath)
		watcher.Add(briefingPath)
		for _, theater := range []camtac.Theater{camtac.Korea, camtac.Balkans, camtac.Israel, camtac.Hellas} {
			probePath := filepath.Join(cfg.FalconBase, "Data", string(theater), "Campaign")
			_, err := os.Stat(probePath)
			if err != nil {
				logger.Infof("%s not found", theater)
				continue
			}
			if theater == camtac.Korea {
				logger.Infof("Adding watcher for default theater")
			} else {
				logger.Infof("Adding watcher for theater: %s", theater)
			}
			watcher.Add(probePath)
		}
		watcher.Start()
	}

	manager := camtac.NewMissionManager(cfg.FalconBase)

	camFiles := manager.GetMissionFiles(theater, camtac.MissionTypeCAM)
	logger.Infof("Found %d campaign files", len(camFiles))
	if len(camFiles) > 0 {
		for _, file := range camFiles {
			logger.Infof("- %s, %s", file.Filename, file.SaveDate)
		}
	}

	tacFiles := manager.GetMissionFiles(theater, camtac.MissionTypeTAC)
	logger.Infof("Found %d TE files", len(tacFiles))
	if len(tacFiles) > 0 {
		for _, file := range tacFiles {
			logger.Infof("- %s, %s", file.Filename, file.SaveDate)
		}
	}

	latestMissionFile, err := camtac.LatestMissionFile(campaignPath)
	if err != nil {
		logger.Errorf("Failed to get latest mission file: %v", err)
		return
	}
	logger.Infof("Latest mission file: %s", latestMissionFile)

	if cfg.MissionFilename == "" {
		cfg.MissionFilename = latestMissionFile
	}

	if cfg.MissionFilename != "" {
		manager.ReadMission(theater, cfg.MissionFilename)
		if cfg.ExportJsonDir != "" {
			fileNoExt := strings.TrimSuffix(cfg.MissionFilename, filepath.Ext(cfg.MissionFilename))
			manager.OutputJson(cfg.ExportJsonDir, fileNoExt)
		}
	}
	ws.RegisterMissionResourceEndpoints(manager)

	if cfg.ReadSharedMem {
		logger.Infof("Enabling Falcon BMS shared memory reader")
		smr := NewSharedMemReader()
		_ = smr.Open()
		defer smr.Close()
		ws.RegisterSharedMemoryEndpoint(smr)

		if cfg.Imcs {
			client := NewCollabClient(cfg.ImcsServer, cfg.Callsign, cfg.ImcsSession, smr)
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			_ = client.Run(ctx)
		}

		go func() {
			ticker := time.NewTicker(time.Duration(cfg.PosUpdateFreq) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if smr.IsReady() {
						ship := smr.GetOwnShip()
						msg := AgentMessage{
							Type:    "pos",
							Payload: &ship,
						}
						payload, err := json.Marshal(msg)
						if err != nil {
							logger.Errorf("error serializing position: %s", err)
							continue
						}
						hub.broadcast <- payload
					}
				}
			}
		}()
	} else {
		if cfg.Imcs {
			logger.Warnf("IMCS enabled, but Falcon BMS shared memory reader is disabled")
		}
	}

	ws.Start()
	if watcher != nil {
		watcher.Stop()
	}
}

func minimizeObjectives(objectives []*camtac.Objective) []*Objective {
	result := make([]*Objective, 0, len(objectives))
	for _, obj := range objectives {
		result = append(result, minimizeObjective(obj))
	}
	return result
}

func minimizeObjective(objective *camtac.Objective) *Objective {
	return &Objective{
		Name:   objective.CampName,
		OcdIdx: objective.ClassType.EntityIdx,
		Type:   objective.ClassType.Type,
		Owner:  int(objective.CampaignBase.Owner),
		Pos: Point{
			// swapped intentionally because of falconq
			objective.PosY,
			objective.PosX,
		},
	}
}
