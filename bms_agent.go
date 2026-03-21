package main

import "log/slog"

type FalconBmsAgent struct {
	logger       *slog.Logger
	hub          *WebsocketHub
	smr          *SharedMemReader
	collabClient *CollabClient
}

func NewFalconBmsAgent(logger *slog.Logger) *FalconBmsAgent {
	return &FalconBmsAgent{
		logger:       logger.With(slog.String("component", "FalconBmsAgent")),
		hub:          NewWebsocketHub(),
		smr:          &SharedMemReader{},
		collabClient: NewCollabClient("", "", "", nil),
	}
}
