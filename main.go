package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	RunApp(false)
}

func RunApp(testMode bool) {
	// 1. Setup
	appConfig = loadConfig()
	ensureIndexExists()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 2. Initial Scan (Calls watchRecursive from watcher.go)
	for _, path := range appConfig.WatchPaths {
		log.Printf("Initializing scan on %s...", path)
		watchRecursive(path, watcher)
	}

	// 3. Start Background Tasks (Calls startBulkShipper from shipper.go)
	go startBulkShipper()

	// Heartbeat goroutine
	go func() {
		for {
			addToBuffer(FimLog{
				Timestamp: time.Now().Format(time.RFC3339),
				FilePath:  "SYSTEM",
				Event:     "HEARTBEAT",
				Owner:     "AGENT",
			})
			time.Sleep(1 * time.Hour)
		}
	}()

	timers := make(map[string]*time.Timer)
	var timerMutex sync.Mutex

	fmt.Printf("FIM Agent Active. Monitoring %d paths.\n", len(appConfig.WatchPaths))

	if testMode {
        return // Exit early so the test doesn't hang
    }

	// 4. Main Event Loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok { return }
			if isExcluded(event.Name, appConfig.Exclusions) { continue }

			if event.Has(fsnotify.Rename) {
				processFile(event.Name, "RENAMED_FROM")
				watcher.Remove(event.Name)
				for _, root := range appConfig.WatchPaths {
					if strings.HasPrefix(event.Name, root) {
						watchRecursive(root, watcher)
						break
					}
				}
				continue
			}

			eventType := "UNKNOWN"
			if event.Has(fsnotify.Write)  { eventType = "MODIFIED" }
			if event.Has(fsnotify.Create) { eventType = "CREATED" }
			if event.Has(fsnotify.Remove) { eventType = "DELETED" }

			timerMutex.Lock()
			if t, ok := timers[event.Name]; ok { t.Stop() }
			timers[event.Name] = time.AfterFunc(500*time.Millisecond, func() {
				// Calls processFile from processor.go
				processFile(event.Name, eventType)
			})
			timerMutex.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok { return }
			log.Println("Watcher error:", err)
		}
	}
}