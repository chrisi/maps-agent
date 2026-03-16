package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"maps-agent/camtac"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	defaultFalconBase := "C:/Progam Files/Falcon BMS 4.38"
	defaultCallsign := "Viper"
	readCamtac := flag.Bool("camtac", false, "Read TAM/TAC files")
	falconBase := flag.String("path", defaultFalconBase, "Falcon BMS directory")
	imcs := flag.Bool("imcs", false, "Establish IMCS connection")
	imcsServer := flag.String("imcs-server", "wss://collab.falcon-bms.com:443", "IMCS Server the client should connect to")
	imcsSession := flag.String("imcs-session", "47DF", "IMCS Session the client should connect to")
	callsign := flag.String("callsign", defaultCallsign, "Callsign")
	fsCheckFreqStr := flag.String("fs-check-freq", "1000", "Milliseconds between checking file system for changes")
	posUpdateFreqStr := flag.String("pos-update-freq", "250", "Milliseconds between sending position updates")
	addr := flag.String("addr", ":8080", "HTTP service address")
	flag.Parse()

	if *readCamtac {
		filename := "mc-test-campaing.cam"
		//filename := "te_1_flight.tac"
		//filename := "bata_1.tac"

		camtac.ReadMission(*falconBase, filename, "c:/projects/Skunkworks/cam-tac-files")

		os.Exit(0)
	}

	fsCheckFreq, _ := strconv.Atoi(*fsCheckFreqStr)
	posUpdateFreq, _ := strconv.Atoi(*posUpdateFreqStr)

	fullPath := filepath.Join(*falconBase, `User\Config`, *callsign+".ini")

	hub := newHub()
	go hub.run()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	smr := SharedMemReader{}
	err = smr.open()
	if err != nil {
		log.Fatalf("error opening Falcon BMS shared memory: %v", err)
	}
	log.Println("successfully opened Falcon BMS shared memory")
	log.Printf("FlightData-version: %d\n", smr.getVersion())

	if *imcs {
		client := newCollabClient(*imcsServer, &smr, *callsign, *imcsSession)
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		if err := client.Run(ctx); err != nil {
			log.Fatalf("error connecting to the IMCS server: %v", err)
		}
		log.Println("successfully connected to the IMCS server")
	}

	defer smr.close()

	go func() {
		var (
			debounceTimer *time.Timer
			mu            sync.Mutex
		)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) && filepath.Base(event.Name) == filepath.Base(fullPath) {
					mu.Lock()
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(time.Duration(fsCheckFreq)*time.Millisecond, func() {
						log.Printf("broadcasting callsign.ini update")
						msg := AgentMessage{
							Type: "update",
						}
						data, err := json.Marshal(msg)
						if err != nil {
							log.Printf("error marshaling ship data: %v", err)
							return
						}
						hub.broadcast <- data
					})
					mu.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()

	err = watcher.Add(filepath.Dir(fullPath))
	if err != nil {
		log.Printf("warning: failed to add watcher for %s: %v", filepath.Dir(fullPath), err)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(posUpdateFreq) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ship := smr.getOwnShip()
				msg := AgentMessage{
					Type:    "pos",
					Payload: &ship,
				}
				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("error marshaling ship data: %v", err)
					continue
				}
				hub.broadcast <- data
			}
		}
	}()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type"},
	}))

	r.GET("/pos", func(c *gin.Context) {
		read := smr.getOwnShip()
		c.JSON(http.StatusOK, read)
	})

	r.GET("/ini", func(c *gin.Context) {
		log.Printf("callsign.ini requested")
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, fullPath+" not found")
			return
		}
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File(fullPath)
	})

	r.GET("/ws", func(c *gin.Context) {
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

	log.Printf("starting server on %s", *addr)
	log.Printf("callsign.ini file: %s", fullPath)
	if err := r.Run(*addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
