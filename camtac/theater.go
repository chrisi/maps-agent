package camtac

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// GetActiveTheater returns the active theater based on the provided version.
func GetActiveTheater(version string) Theater {
	keyPath := `SOFTWARE\WOW6432Node\Benchmark Sims\` + version

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
	if err != nil {
		log.Printf("Failed to open registry key: %v", err)
		return Korea
	}
	defer k.Close()

	theater, _, err := k.GetStringValue("curTheater")
	if err != nil {
		log.Printf("Error reading curTheater value from registry: %v", err)
		return Korea
	}

	switch theater {
	case "":
		return Korea
	case "Balkans":
		return Balkans
	case "Israel":
		return Israel
	case "Hellas":
		return Hellas
	}
	log.Println("Could not determine theater, falling back to Korea")
	return Korea
}

// LatestMissionFile returns the name of the most recently modified .tac or .cam file in directory.
func LatestMissionFile(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}

	var latestName string
	var latestModTime int64
	for _, entry := range entries {
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".tac" && extension != ".cam" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return "", fmt.Errorf("read file information for %q: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() {
			continue
		}

		modTime := info.ModTime().UnixNano()
		if latestName == "" || modTime > latestModTime {
			latestName = entry.Name()
			latestModTime = modTime
		}
	}

	return latestName, nil
}
