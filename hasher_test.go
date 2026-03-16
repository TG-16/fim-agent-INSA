package main

import (
	"os"
	"testing"
)

func TestGetHash(t *testing.T) {
	// --- SETUP: Create a temporary file to hash ---
	tempFile := "test_file.txt"
	content := []byte("hello world")
	err := os.WriteFile(tempFile, content, 0644)
	if err != nil {
		t.Fatal("Could not create temp file for testing")
	}

	// Make sure the file is deleted after the test finishes
	defer os.Remove(tempFile)

	// --- TEST SCENARIO 1: Valid File ---
	// This is the expected SHA256 for "hello world"
	expectedHash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	
	actualHash := getHash(tempFile)

	if actualHash != expectedHash {
		// If they don't match, tell 't' to fail the test
		t.Errorf("Hash mismatch! Expected %s, got %s", expectedHash, actualHash)
	}

	// --- TEST SCENARIO 2: Non-existent File ---
	// If a file doesn't exist, your function returns ""
	missingFileHash := getHash("does_not_exist.txt")
	if missingFileHash != "" {
		t.Errorf("Expected empty string for missing file, but got %s", missingFileHash)
	}
}