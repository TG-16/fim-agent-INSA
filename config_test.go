package main

import (
	"os"
	"testing"
)

// 1. Test the logic for excluding files
func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		exclusions []string
		want       bool
	}{
		{"Match folder", "C:\\Windows\\Temp\\file.log", []string{"C:\\Windows\\Temp"}, true},
		{"No match", "C:\\Users\\Desktop\\file.txt", []string{"C:\\Windows"}, false},
		{"Linux match", "/etc/passwd", []string{"/etc"}, true},
		{"Empty list", "/home/user", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExcluded(tt.path, tt.exclusions); got != tt.want {
				t.Errorf("isExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 2. Test loading the config from a file
func TestLoadConfig(t *testing.T) {
	// Setup: Create a temporary config.json
	tmpFile := "config.json"
	content := []byte(`{"opensearch_url": "http://localhost:9200", "ship_interval_sec": 10}`)
	
	// Backup real config if it exists
	_, err := os.Stat(tmpFile)
	realExists := err == nil
	if realExists {
		os.Rename(tmpFile, "config.json.bak")
	}

	err = os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatal("Could not write temp config file")
	}
	
	// Cleanup: restore real config after test
	defer func() {
		os.Remove(tmpFile)
		if realExists {
			os.Rename("config.json.bak", tmpFile)
		}
	}()

	cfg := loadConfig()

	if cfg.OSUrl != "http://localhost:9200" {
		t.Errorf("Expected URL http://localhost:9200, got %s", cfg.OSUrl)
	}
}

// 3. Test the log buffer (NEW)
func TestAddToBuffer(t *testing.T) {
	// Reset the buffer before testing
	batchMutex.Lock()
	logBuffer = []FimLog{}
	batchMutex.Unlock()

	testEntry := FimLog{
		FilePath: "test.txt",
		Event:    "CREATED",
	}

	// Call the function
	addToBuffer(testEntry)

	// Check if the buffer has exactly 1 item
	batchMutex.Lock()
	defer batchMutex.Unlock()
	
	if len(logBuffer) != 1 {
		t.Errorf("Expected buffer size 1, got %d", len(logBuffer))
	}

	if logBuffer[0].FilePath != "test.txt" {
		t.Errorf("Buffer data mismatch! Got %s", logBuffer[0].FilePath)
	}
}