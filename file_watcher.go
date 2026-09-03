package main

import (
	"maps-agent/util"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	logger      *util.Logger
	watcher     *fsnotify.Watcher
	fsCheckFreq int
	callbacks   []func(string)
	mu          sync.Mutex
}

func NewFileWatcher(fsCheckFreq int, callbacks ...func(string)) (*FileWatcher, error) {
	logger := util.NewLogger("FileWatcher", os.Stdout, util.Debug, true)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("failed to create watcher: %v", err)
		return nil, err
	}
	return &FileWatcher{
		logger:      logger,
		watcher:     watcher,
		fsCheckFreq: fsCheckFreq,
		callbacks:   callbacks,
	}, nil
}

func (fw *FileWatcher) OnChange(cb func(string)) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.callbacks = append(fw.callbacks, cb)
}

func (fw *FileWatcher) Stop() {
	err := fw.watcher.Close()
	if err != nil {
		fw.logger.Errorf("failed to close watcher: %v", err)
	}
}

func (fw *FileWatcher) Add(path string) {
	err := fw.watcher.Add(path)
	if err != nil {
		fw.logger.Errorf("failed to watch %s: %v", path, err)
	}
}

func (fw *FileWatcher) Start() {
	go func() {
		var (
			debounceTimers = make(map[string]*time.Timer)
			mu             sync.Mutex
		)
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					fw.logger.Errorf("event not ok: %s", event.Name)
					return
				}
				if event.Has(fsnotify.Write) &&
					(strings.HasSuffix(event.Name, "briefing.txt") ||
						strings.HasSuffix(event.Name, "ini") ||
						strings.HasSuffix(event.Name, "cam") ||
						strings.HasSuffix(event.Name, "tac")) {
					fileName := event.Name
					mu.Lock()
					if timer, exists := debounceTimers[fileName]; exists && timer != nil {
						timer.Stop()
					}
					var timer *time.Timer
					timer = time.AfterFunc(time.Duration(fw.fsCheckFreq)*time.Millisecond, func() {
						mu.Lock()
						if debounceTimers[fileName] == timer {
							delete(debounceTimers, fileName)
						}
						mu.Unlock()

						fw.logger.Infof("file changed: %s", fileName)
						fw.notifyCallbacks(fileName)
					})
					debounceTimers[fileName] = timer
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

func (fw *FileWatcher) notifyCallbacks(fileName string) {
	fw.mu.Lock()
	callbacks := make([]func(string), len(fw.callbacks))
	copy(callbacks, fw.callbacks)
	fw.mu.Unlock()

	for _, cb := range callbacks {
		cb(fileName)
	}
}
