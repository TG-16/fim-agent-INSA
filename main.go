package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	OSUrl           string   `json:"opensearch_url"`
	OSUser          string   `json:"opensearch_user"`
	OSPass          string   `json:"opensearch_pass"`
	WatchPaths      []string `json:"watch_paths"`
	IndexName       string   `json:"index_name"`
	ShipIntervalSec int      `json:"ship_interval_sec"`
	LogLevel        int      `json:"log_level"`
	Exclusions      []string `json:"exclusions"`
}

type FimLog struct {
	Timestamp   string `json:"@timestamp"`
	FilePath    string `json:"file_path"`
	Event       string `json:"event_type"`
	Hash        string `json:"hash_sha256,omitempty"`
	Size        int64  `json:"file_size_bytes"`
	Permissions string `json:"permissions"`
	Owner       string `json:"owner"`
}

var (
	appConfig  Config
	batchMutex sync.Mutex
	logBuffer  []FimLog
	metaCache  sync.Map
)

func loadConfig() {
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
        log.Fatalf("CRITICAL ERROR: 'config.json' not found! Please create it in the same folder as the agent.")
    }

	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Cannot open config.json: %v", err)
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&appConfig)
}

func isExcluded(path string) bool {
	for _, exc := range appConfig.Exclusions {
		if strings.HasPrefix(path, exc) {
			return true
		}
	}
	return false
}

func main() {
	loadConfig()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 1. Loop through all paths in your config for initial scan
	for _, path := range appConfig.WatchPaths {
		log.Printf("Initializing scan on %s...", path)
		watchRecursive(path, watcher)
	}

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

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if isExcluded(event.Name) {
				continue
			}

			// Handle Rename specifically
			if event.Has(fsnotify.Rename) {
				processFile(event.Name, "RENAMED_FROM")
				watcher.Remove(event.Name)
				// Re-scan based on the root of the event to catch the new name
				// We search which root path contains this event name
				for _, root := range appConfig.WatchPaths {
					if strings.HasPrefix(event.Name, root) {
						watchRecursive(root, watcher)
						break
					}
				}
				continue
			}

			eventType := "UNKNOWN"
			if event.Has(fsnotify.Write) {
				eventType = "MODIFIED"
			}
			if event.Has(fsnotify.Create) {
				eventType = "CREATED"
			}
			if event.Has(fsnotify.Remove) {
				eventType = "DELETED"
			}

			timerMutex.Lock()
			if t, ok := timers[event.Name]; ok {
				t.Stop()
			}
			timers[event.Name] = time.AfterFunc(500*time.Millisecond, func() {
				processFile(event.Name, eventType)
			})
			timerMutex.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func watchRecursive(root string, watcher *fsnotify.Watcher) error {
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if isExcluded(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			if count > 500 {
				if appConfig.LogLevel >= 1 {
					log.Printf("[WARNING] Watch limit reached. Skipping deeper folders in %s", path)
				}
				return filepath.SkipDir
			}

			err := watcher.Add(path)
			if err != nil {
				return nil
			}
			count++
		} else {
			owner := getOwnerInfo(info, path)
			metaCache.Store(path, owner)
		}
		return nil
	})
	return err
}

func processFile(path string, eventType string) {
	info, err := os.Stat(path)
	if err != nil {
		lastOwner := "Unknown"
		if val, ok := metaCache.Load(path); ok {
			lastOwner = val.(string)
		}
		if os.IsNotExist(err) {
			addToBuffer(FimLog{
				Timestamp: time.Now().Format(time.RFC3339),
				FilePath:  path,
				Event:     eventType,
				Owner:     lastOwner,
			})
			if eventType == "DELETED" || eventType == "RENAMED_FROM" {
				metaCache.Delete(path)
			}
		}
		return
	}
	if info.IsDir() {
		return
	}

	currentOwner := getOwnerInfo(info, path)
	metaCache.Store(path, currentOwner)

	fileHash := "SKIPPED_LARGE_FILE"
	if info.Size() < 100*1024*1024 {
		fileHash = getHash(path)
	}

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

func getHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
}

func addToBuffer(entry FimLog) {
	batchMutex.Lock()
	logBuffer = append(logBuffer, entry)
	batchMutex.Unlock()
}

func ensureIndexExists() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	// 1. Check if index already exists
	checkReq, _ := http.NewRequest("HEAD", strings.TrimSuffix(appConfig.OSUrl, "/_bulk")+"/"+appConfig.IndexName, nil)
	checkReq.SetBasicAuth(appConfig.OSUser, appConfig.OSPass)
	resp, err := client.Do(checkReq)
	
	if err == nil && resp.StatusCode == 200 {
		return // Index already exists, nothing to do
	}

	// 2. If it doesn't exist, create it with mappings
	mapping := `{
		"mappings": {
			"properties": {
				"@timestamp":  { "type": "date" },
				"file_path":    { "type": "keyword" },
				"event_type":   { "type": "keyword" },
				"hash_sha256":  { "type": "keyword" },
				"owner":        { "type": "keyword" }
			}
		}
	}`

	createUrl := strings.TrimSuffix(appConfig.OSUrl, "/_bulk") + "/" + appConfig.IndexName
	req, _ := http.NewRequest("PUT", createUrl, bytes.NewBuffer([]byte(mapping)))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(appConfig.OSUser, appConfig.OSPass)

	client.Do(req)
	log.Println("[INFO] OpenSearch Index initialized automatically.")
}

func startBulkShipper() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: 15 * time.Second}

	for {
		time.Sleep(time.Duration(appConfig.ShipIntervalSec) * time.Second)

		batchMutex.Lock()
		count := len(logBuffer)
		if count == 0 {
			batchMutex.Unlock()
			continue
		}

		var bulkBody bytes.Buffer
		for _, logEntry := range logBuffer {
			bulkBody.WriteString(fmt.Sprintf(`{"index":{"_index":"%s"}}`, appConfig.IndexName) + "\n")
			data, _ := json.Marshal(logEntry)
			bulkBody.Write(data)
			bulkBody.WriteString("\n")
		}
		batchMutex.Unlock()

		req, _ := http.NewRequest("POST", appConfig.OSUrl, &bulkBody)
		req.Header.Set("Content-Type", "application/x-ndjson")
		req.SetBasicAuth(appConfig.OSUser, appConfig.OSPass)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[ERROR] Network error: %v", err)
			continue
		}

		if resp.StatusCode >= 400 {
			log.Printf("[ERROR] OpenSearch rejected logs: %s", resp.Status)
			resp.Body.Close()
			continue
		}

		if appConfig.LogLevel >= 2 {
			log.Printf("[INFO] Successfully shipped %d logs to OpenSearch (Status: %s)", count, resp.Status)
		}

		resp.Body.Close()

		batchMutex.Lock()
		logBuffer = nil
		batchMutex.Unlock()
	}
}