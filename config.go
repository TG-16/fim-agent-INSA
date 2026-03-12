package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
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

func loadConfig() Config {
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		log.Fatalf("CRITICAL ERROR: 'config.json' not found!")
	}
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Cannot open config.json: %v", err)
	}
	defer file.Close()
	var cfg Config
	json.NewDecoder(file).Decode(&cfg)
	return cfg
}

func isExcluded(path string, exclusions []string) bool {
	for _, exc := range exclusions {
		if strings.HasPrefix(path, exc) {
			return true
		}
	}
	return false
}

func addToBuffer(entry FimLog) {
    batchMutex.Lock()
    logBuffer = append(logBuffer, entry)
    batchMutex.Unlock()
}