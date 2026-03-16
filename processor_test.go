package main

import (
	"os"
	"testing"
)

func TestProcessFile(t *testing.T) {
	// 1. Setup: Reset the global buffer so we start fresh
	batchMutex.Lock()
	logBuffer = []FimLog{}
	batchMutex.Unlock()

	testFile := "test_event.txt"
	content := []byte("unit test content")

	// --- SCENARIO A: File Created (The Happy Path) ---
	err := os.WriteFile(testFile, content, 0644)
	if err != nil { t.Fatal(err) }
	defer os.Remove(testFile) // Cleanup

	processFile(testFile, "CREATED")

	batchMutex.Lock()
	if len(logBuffer) == 0 {
		t.Error("Expected a log in the buffer, but it is empty")
	} else if logBuffer[0].Event != "CREATED" {
		t.Errorf("Expected CREATED event, got %s", logBuffer[0].Event)
	}
	batchMutex.Unlock()

	// --- SCENARIO B: Directory (Should be ignored) ---
	testDir := "test_folder"
	os.Mkdir(testDir, 0755)
	defer os.Remove(testDir)
	
	currentLogCount := len(logBuffer)
	processFile(testDir, "CREATED")
	
	if len(logBuffer) > currentLogCount {
		t.Error("Processor should ignore directories, but a log was added")
	}

	// --- SCENARIO C: File Deleted ---
	os.Remove(testFile)
	// We call it with DELETED. It should find that the file doesn't exist 
	// and still add a log based on the logic in your processor.go
	processFile(testFile, "DELETED")

	if len(logBuffer) <= currentLogCount {
		t.Error("Expected a log for the DELETED event")
	}
}