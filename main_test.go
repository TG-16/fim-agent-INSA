package main

import (
	"testing"
)

func TestRunAppInTestMode(t *testing.T) {
	// We call the function directly. 
	// Because testMode is true, it will initialize everything and then RETURN
	// instead of entering the infinite loop.
	
	// Note: This requires a config.json to exist in the folder.
	// We already have logic for this in TestLoadConfig, so it should work!
	RunApp(true)
}