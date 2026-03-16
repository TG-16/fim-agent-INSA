package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnsureIndexExists(t *testing.T) {
	// Step 1: Mock Server that handles HEAD and PUT
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			// Simulate index NOT existing the first time
			w.WriteHeader(http.StatusNotFound)
		} else if r.Method == "PUT" {
			// Simulate successful creation
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Step 2: Configure app to point to Mock
	appConfig.OSUrl = server.URL + "/_bulk"
	appConfig.IndexName = "test-index"
	appConfig.OSUser = "admin"
	appConfig.OSPass = "admin"

	// Step 3: Run function
	ensureIndexExists()
    
    // If the function reaches here without crashing, it covered both HEAD and PUT logic.
}