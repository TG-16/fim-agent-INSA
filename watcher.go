package main

import (
	"log"
	"os"
	"path/filepath"
	"github.com/fsnotify/fsnotify"
)

func watchRecursive(root string, watcher *fsnotify.Watcher) error {
	count := 0
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if isExcluded(path, appConfig.Exclusions) {
			if info.IsDir() { return filepath.SkipDir }
			return nil
		}
		if info.IsDir() {
			if count > 500 {
				if appConfig.LogLevel >= 1 { log.Printf("[WARNING] Watch limit reached in %s", path) }
				return filepath.SkipDir
			}
			watcher.Add(path)
			count++
		} else {
			metaCache.Store(path, getOwnerInfo(info, path))
		}
		return nil
	})
}