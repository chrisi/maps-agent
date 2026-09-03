package main

import (
	"fmt"
	"log"
	"maps-agent/camtac"
	"maps-agent/util"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Webserver struct {
	logger *util.Logger
	r      *gin.Engine
	addr   string
}

func NewWebserver(addr string) *Webserver {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type"},
	}))

	return &Webserver{
		logger: util.NewLogger("Webserver", os.Stdout, util.Info, true),
		r:      r,
		addr:   addr,
	}
}

func (w *Webserver) RegisterBriefingEndpoint(briefingPath string) {
	w.logger.Infof("Registering briefing endpoint: %s", briefingPath)
	w.r.GET("/briefing", func(c *gin.Context) {
		log.Printf("briefing.txt requested")
		if _, err := os.Stat(briefingPath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, briefingPath+" not found")
			return
		}
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File(briefingPath)
	})
}

func (w *Webserver) RegisterIniEndpoint(iniPath string) {
	w.logger.Infof("Registering callsign.ini endpoint: %s", iniPath)
	w.r.GET("/ini", func(c *gin.Context) {
		log.Printf("callsign.ini requested")
		if _, err := os.Stat(iniPath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, iniPath+" not found")
			return
		}
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File(iniPath)
	})
}

func (w *Webserver) RegisterFlightEndpoint(fm *FlightManager) {
	w.logger.Infof("Registering flight endpoint: /flight")
	w.r.GET("/flight", func(c *gin.Context) {
		flight := fm.GetCurrentFlight()
		if flight == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no flight data available"})
			return
		}
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.JSON(http.StatusOK, flight)
	})
}

func getTheater(theater string) (camtac.Theater, error) {
	s := strings.ToLower(theater)
	switch s {
	case "korea":
		return camtac.Korea, nil
	case "balkans":
		return camtac.Balkans, nil
	case "israel":
		return camtac.Israel, nil
	case "hellas":
		return camtac.Hellas, nil
	default:
		return "", fmt.Errorf("unknown theater %s", theater)
	}
}

func (w *Webserver) RegisterMissionResourceEndpoints(manager *camtac.MissionManager) {
	w.logger.Infof("Registering mission resource endpoint")
	w.r.GET("/mission/stations", func(c *gin.Context) {
		read := minimizeObjectives(manager.GetObjectivesByTypes([]int{camtac.TypeAirbase, camtac.TypeAirstrip, camtac.TypeNavBeacon}))
		c.JSON(http.StatusOK, read)
	})

	w.r.GET("/mission/factories", func(c *gin.Context) {
		read := minimizeObjectives(manager.GetObjectivesByTypes([]int{camtac.TypeFactory}))
		c.JSON(http.StatusOK, read)
	})

	w.r.GET("/missionfiles/:theater/:type", func(c *gin.Context) {
		theater, err := getTheater(c.Param("theater"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tp := camtac.MissionType(strings.ToLower(c.Param("type")))
		if tp.Valid() == false {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid mission type %q", tp)})
			return
		}
		read := manager.GetMissionFiles(theater, tp)
		c.JSON(http.StatusOK, read)
	})
}

func (w *Webserver) RegisterSharedMemoryEndpoint(smr *SharedMemReader) {
	w.logger.Infof("Registering aircraft position endpoint")
	if smr != nil {
		w.r.GET("/pos", func(c *gin.Context) {
			read := smr.GetOwnShip()
			c.JSON(http.StatusOK, read)
		})
	}
}

func (w *Webserver) RegisterWebsocketEndpoint(hub *WebsocketHub) {
	w.r.GET("/ws", func(c *gin.Context) {
		log.Printf("client connected from %s", c.ClientIP())
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("upgrade error: %v", err)
			return
		}
		hub.register <- conn
		defer func() {
			hub.unregister <- conn
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})
}

func (w *Webserver) Start() {
	w.logger.Infof("starting server on %s", w.addr)

	if err := w.r.Run(w.addr); err != nil {
		w.logger.Errorf("failed to start server: %v", err)
	}
}
