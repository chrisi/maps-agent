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
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity, match existing CORS policy
	},
}

type AgentMessage struct {
	Type    string      `json:"type"`
	Version int         `json:"version,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func main() {
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
					debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
						log.Printf("broadcasting update for: %s", fullPath)
						msg := AgentMessage{
							Type: "update",
						}
						if err != nil {
							log.Printf("Error reading shared memory: %v", err)
							return
						}
						data, err := json.Marshal(msg)
						if err != nil {
							log.Printf("Error marshaling ship data: %v", err)
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
		log.Printf("Warning: Failed to add watcher for %s: %v", filepath.Dir(fullPath), err)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ship, err := smr.read()
				if err != nil {
					log.Printf("Error reading shared memory: %v", err)
					continue
				}

				msg := AgentMessage{
					Type:    "pos",
					Payload: &ship,
				}

				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Error marshaling ship data: %v", err)
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
		read, err2 := smr.read()

		if err2 != nil {
			c.String(http.StatusInternalServerError, err2.Error())
			return
		}
		c.JSON(http.StatusOK, read)
	})

	r.GET("/ini", func(c *gin.Context) {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, fullPath+" not found")
			return
		}

		log.Printf("Serving file: %s", fullPath)
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File(fullPath)
	})

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
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

	log.Printf("Starting server on %s", *addr)

	if err := r.Run(*addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
