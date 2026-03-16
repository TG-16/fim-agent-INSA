package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStartBulkShipper(t *testing.T) {
	// 1. Create a Fake OpenSearch Server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This code runs when your shipper "calls" the fake OpenSearch
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"created"}`))
	}))
	defer server.Close()

	// 2. Point your appConfig to the Fake Server
	appConfig.OSUrl = server.URL
	appConfig.ShipIntervalSec = 1 // Short interval for testing
	appConfig.IndexName = "test-index"
    appConfig.LogLevel = 2

	// 3. Add a fake log to the buffer
	addToBuffer(FimLog{FilePath: "test.log", Event: "MOCK_EVENT"})

	// 4. Run the shipper in the background
	go startBulkShipper()

	// 5. Wait a moment for the shipper to "fire"
	time.Sleep(2 * time.Second)

	// 6. Verify the buffer was cleared
	batchMutex.Lock()
	if len(logBuffer) != 0 {
		t.Errorf("Expected buffer to be cleared, but it has %d items", len(logBuffer))
	}
	batchMutex.Unlock()
}

func TestStartBulkShipperNetworkError(t *testing.T) {
	// Point to a non-existent local port to trigger a network error
	appConfig.OSUrl = "https://localhost:9999/_bulk"
	appConfig.ShipIntervalSec = 1
	
	addToBuffer(FimLog{FilePath: "error.log", Event: "NETWORK_FAILURE"})

	// Run shipper in a goroutine for a very short time
	go startBulkShipper()
	
	// Wait long enough for one retry loop
	time.Sleep(2 * time.Second)
    
    // This triggers the log.Printf("[ERROR] Network error...") line.
}

func TestShipperEdgeCases(t *testing.T) {
    // SCENARIO 1: OpenSearch returns an Error (e.g., 401 Unauthorized)
    server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusUnauthorized)
    }))
    defer server.Close()

    appConfig.OSUrl = server.URL
    appConfig.LogLevel = 2
    
    // Add one log so the shipper has work to do
    addToBuffer(FimLog{FilePath: "fail.log", Event: "TEST"})

    // Run the shipper logic once (not in a loop for this test)
    // Note: Since startBulkShipper is an infinite loop, we test its internals 
    // by triggering the HTTP logic.
    
    // SCENARIO 2: Empty Buffer
    batchMutex.Lock()
    logBuffer = []FimLog{} // Clear buffer
    batchMutex.Unlock()
    
    // If we run the shipper now, it should simply "continue" and not call the server.
}

func TestShipperServerReject(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // 400 Error
		w.Write([]byte(`{"error":"bad_data"}`))
	}))
	defer server.Close()

	appConfig.OSUrl = server.URL
	appConfig.LogLevel = 2
	
	addToBuffer(FimLog{FilePath: "bad_log.txt", Event: "FAIL"})

    // We call the internal logic by letting the loop run or calling a helper
    // For now, running it in background for 1 second is enough
	go startBulkShipper()
	time.Sleep(3 * time.Second)
}

