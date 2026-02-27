package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	defaultPath := `C:\Progam Files\Falcon BMS 4.38`
	defaultCallsign := `Viper`
	configPath := flag.String("path", defaultPath, "Falcon BMS directory")
	callsign := flag.String("callsign", defaultCallsign, "Callsign")
	addr := flag.String("addr", ":8080", "HTTP service address")
	flag.Parse()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type"},
	}))

	r.GET("/ini", func(c *gin.Context) {
		fullPath := filepath.Join(*configPath, `User\Config`, *callsign+".ini")

		fmt.Println(fullPath)

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

	log.Printf("Starting server on %s", *addr)

	if err := r.Run(*addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
