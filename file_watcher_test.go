package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestFileWatcher_NotifiesCallbacks(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "briefing.txt")

	var (
		mu           sync.Mutex
		receivedFile string
		notified     = make(chan struct{}, 1)
	)

	fw, err := NewFileWatcher(50, func(path string) {
		mu.Lock()
		receivedFile = path
		mu.Unlock()
		select {
		case notified <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("failed to create file watcher: %v", err)
	}
	defer fw.Stop()

	fw.Add(tempDir)
	fw.Start()

	// Write file
	time.Sleep(50 * time.Millisecond)
	err = os.WriteFile(testFile, []byte("briefing data"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	select {
	case <-notified:
		mu.Lock()
		defer mu.Unlock()
		if filepath.Base(receivedFile) != "briefing.txt" {
			t.Errorf("expected briefing.txt, got %s", receivedFile)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for file watcher notification")
	}
}
