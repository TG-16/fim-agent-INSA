package main

import (
	"os"
	"time"
)

func processFile(path string, eventType string) {
	info, err := os.Stat(path)
	if err != nil {
		lastOwner := "Unknown"
		if val, ok := metaCache.Load(path); ok { lastOwner = val.(string) }
		if os.IsNotExist(err) {
			addToBuffer(FimLog{
				Timestamp: time.Now().Format(time.RFC3339),
				FilePath:  path,
				Event:     eventType,
				Owner:     lastOwner,
			})
			if eventType == "DELETED" || eventType == "RENAMED_FROM" { metaCache.Delete(path) }
		}
		return
	}
	if info.IsDir() { return }

	currentOwner := getOwnerInfo(info, path)
	metaCache.Store(path, currentOwner)

	fileHash := "SKIPPED_LARGE_FILE"
	if info.Size() < 100*1024*1024 { fileHash = getHash(path) }

	addToBuffer(FimLog{
		Timestamp:   time.Now().Format(time.RFC3339),
		FilePath:    path,
		Event:       eventType,
		Hash:        fileHash,
		Size:        info.Size(),
		Permissions: info.Mode().String(),
		Owner:       currentOwner,
	})
}