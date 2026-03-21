package main

import (
	"context"
	"encoding/json"
	"maps-agent/util"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// The CollabClient is used to share position updates with other flight members when connected to IMCS.
// This is in contrast to the WebsocketHub which sends position updates directly to the Maps-App
// usually hosted locally alongside the BMS Client when connected to this agent.
type CollabClient struct {
	logger    *util.Logger
	serverURL string
	callsign  string
	session   string
	smr       *SharedMemReader
}

func NewCollabClient(serverURL string, callsign string, session string, smr *SharedMemReader) *CollabClient {
	return &CollabClient{
		logger:    util.NewLogger("CollabClient", os.Stdout, util.Info, true),
		smr:       smr,
		serverURL: serverURL,
		callsign:  callsign,
		session:   session,
	}
}

func (c *CollabClient) Run(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		c.logger.Errorf("failed to connect to IMCS: %v\n", err)
		return err
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	c.logger.Infof("connected to IMCS using URL: %s\n ", c.serverURL)

	//TODO: send auth with callsign and session

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("disconnecting from IMCS")
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
			)
			return nil

		case <-ticker.C:
			if c.smr.IsReady() {
				ship := c.smr.GetOwnShip()
				//TODO: change AgentMessage to correct IMCS data structure
				msg := AgentMessage{
					Type:    "pos",
					Payload: &ship,
				}
				payload, err := json.Marshal(msg)
				if err != nil {
					c.logger.Errorf("error serializing position: %s\n", err)
					continue
				}
				if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					return err
				}
			}
		}
	}
}
