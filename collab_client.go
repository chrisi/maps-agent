package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type CollabClient struct {
	reader    *SharedMemReader
	serverURL string
	callsign  string
	session   string
}

func newCollabClient(serverURL string, reader *SharedMemReader, callsign string, session string) *CollabClient {
	return &CollabClient{
		reader:    reader,
		serverURL: serverURL,
		callsign:  callsign,
		session:   session,
	}
}

func (c *CollabClient) Run(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return err
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	log.Printf("connected to IMCS: %s", c.serverURL)

	//TODO: send auth with callsign and session

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("disconnecting from IMCS")
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
			)
			return nil

		case <-ticker.C:
			pos := c.reader.getOwnShip()
			if err != nil {
				log.Printf("error reading position: %v", err)
				continue
			}

			payload, err := json.Marshal(pos)
			if err != nil {
				log.Printf("error serializing position: %v", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return err
			}
		}
	}
}
