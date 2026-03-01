package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	defaultPath := `C:\Progam Files\Falcon BMS 4.38`
	defaultCallsign := `Viper`
	configPath := flag.String("path", defaultPath, "Falcon BMS directory")
	callsign := flag.String("callsign", defaultCallsign, "Callsign")
	addr := flag.String("addr", ":8080", "HTTP service address")
	flag.Parse()

	fullPath := filepath.Join(*configPath, `User\Config`, *callsign+".ini")

	hub := newHub()
	go hub.run()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	smr := SharedMemReader{}
	err = smr.open()
	if err != nil {
		log.Fatal(err)
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
					debounceTimer = time.AfterFunc(1*time.Second, func() {
						log.Printf("broadcasting callsign.ini update")
						msg := AgentMessage{
							Type: "update",
						}
						if err != nil {
							log.Printf("error reading shared memory: %v", err)
							return
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
		ticker := time.NewTicker(1 * time.Second)
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
