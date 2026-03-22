package main

import (
	"encoding/json"
	"maps-agent/util"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	logger      *util.Logger
	watcher     *fsnotify.Watcher
	hub         *WebsocketHub
	fsCheckFreq int
}

func NewFileWatcher(hub *WebsocketHub, fsCheckFreq int) *FileWatcher {
	return &FileWatcher{
		logger:      util.NewLogger("FileWatcher", os.Stdout, util.Debug, true),
		hub:         hub,
		fsCheckFreq: fsCheckFreq,
	}
}

func (fw *FileWatcher) Stop() {
	err := fw.watcher.Close()
	if err != nil {
		fw.logger.Errorf("failed to close watcher: %v", err)
	}
}

func (fw *FileWatcher) Start(path string) {
	if w, err := fsnotify.NewWatcher(); err != nil {
		fw.logger.Errorf("failed to create watcher: %v", err)
	} else {
		fw.watcher = w
	}
	err := fw.watcher.Add(path)
	if err != nil {
		fw.logger.Errorf("failed to watch %s: %v", path, err)
	}

	go func() {
		var (
			debounceTimer *time.Timer
			mu            sync.Mutex
		)
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) && filepath.Base(event.Name) == filepath.Base(path) {
					mu.Lock()
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					fw.logger.Debugf("file changed: %s", event.Name)
					debounceTimer = time.AfterFunc(time.Duration(fw.fsCheckFreq)*time.Millisecond, func() {
						fw.logger.Debugf("broadcasting callsign.ini update")
						msg := AgentMessage{
							Type: "update",
						}
						data, err := json.Marshal(msg)
						if err != nil {
							fw.logger.Errorf("error serializing update: %v", err)
							return
						}
						fw.hub.broadcast <- data
					})
					mu.Unlock()
				}
			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				fw.logger.Errorf("watcher error: %v", err)
			}
		}
	}()
}
