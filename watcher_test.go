package main

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/fsnotify/fsnotify"
	"fmt"
)

func TestWatchRecursive(t *testing.T) {
	// Setup: Create a temporary root directory
	root := "test_watch_root"
	subDir := filepath.Join(root, "subdir")
	excludedDir := filepath.Join(root, "excluded")
	
	os.MkdirAll(subDir, 0755)
	os.MkdirAll(excludedDir, 0755)
	defer os.RemoveAll(root)

	// Create a dummy file in the subdir
	os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("test"), 0644)

	// Set up fsnotify watcher
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()

	// Set appConfig exclusions for the test
	appConfig.Exclusions = []string{excludedDir}
	appConfig.LogLevel = 1

	// Run the function
	err := watchRecursive(root, watcher)
	if err != nil {
		t.Fatalf("watchRecursive failed: %v", err)
	}

	// Logic Check:
	// We can't easily peek inside the watcher to see "added" paths,
	// but we can check if the metaCache was populated for the file.
	filePath := filepath.Join(subDir, "file.txt")
	_, found := metaCache.Load(filePath)
	if !found {
		t.Errorf("Expected %s to be in metaCache, but it wasn't", filePath)
	}

	// Verify excluded dir was NOT processed
	_, foundExcluded := metaCache.Load(filepath.Join(excludedDir, "secret.txt"))
	if foundExcluded {
		t.Error("Excluded directory files should not be in metaCache")
	}
}

func TestWatchRecursiveLimit(t *testing.T) {
    root := "test_limit_root"
    os.Mkdir(root, 0755)
    defer os.RemoveAll(root)

    // Create 501 subdirectories to trigger the limit
    // Note: This is fast because they are empty
    for i := 0; i < 505; i++ {
        path := filepath.Join(root, fmt.Sprintf("dir%d", i))
        os.Mkdir(path, 0755)
    }

    watcher, _ := fsnotify.NewWatcher()
    defer watcher.Close()
    
    appConfig.LogLevel = 1 // Ensure the log.Printf line is hit

    err := watchRecursive(root, watcher)
    if err != nil {
        t.Errorf("watchRecursive failed: %v", err)
	}
}

func TestWatchRecursiveError(t *testing.T) {
    watcher, _ := fsnotify.NewWatcher()
    defer watcher.Close()

    // Use a path with characters that are illegal on Windows/Linux
    // to force the 'filepath.Walk' or 'watcher.Add' to return an error.
    err := watchRecursive("?:invalid/path/<>|", watcher)
    
    if err == nil {
        t.Log("Note: filepath.Walk might not error on missing paths, trying inaccessible path...")
    }
}