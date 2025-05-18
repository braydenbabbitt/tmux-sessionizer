package utils

import (
	"os"
	"time"
)

// GetLastModifiedTime returns the last modified time for a directory
// It finds the most recently modified file in the directory (non-recursive)
func GetLastModifiedTime(dirPath string) time.Time {
	var latestTime time.Time

	// Get directory information first
	dirInfo, err := os.Stat(dirPath)
	if err == nil {
		latestTime = dirInfo.ModTime()
	}

	// Check files in the directory to find the latest modified time
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return latestTime
	}

	for _, file := range files {
		// Skip hidden files and directories
		if file.Name()[0] == '.' {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
		}
	}

	return latestTime
}
