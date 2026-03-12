package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func startBulkShipper() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: 15 * time.Second}

	for {
		time.Sleep(time.Duration(appConfig.ShipIntervalSec) * time.Second)

		batchMutex.Lock()
		if len(logBuffer) == 0 {
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
		count := len(logBuffer)
		batchMutex.Unlock()

		req, _ := http.NewRequest("POST", appConfig.OSUrl, &bulkBody)
		req.Header.Set("Content-Type", "application/x-ndjson")
		req.SetBasicAuth(appConfig.OSUser, appConfig.OSPass)

		resp, err := client.Do(req)
		// ... after resp, err := client.Do(req)
        if err != nil {
            log.Printf("[ERROR] Network error: %v", err)
            continue
        }

        if resp.StatusCode >= 400 {
            log.Printf("[ERROR] OpenSearch rejected logs: %s", resp.Status)
            resp.Body.Close()
            continue
        }

        // USE THE COUNT VARIABLE HERE:
        if appConfig.LogLevel >= 2 {
            log.Printf("[INFO] Successfully shipped %d logs to OpenSearch (Status: %s)", count, resp.Status)
        }

        resp.Body.Close()

		batchMutex.Lock()
		logBuffer = nil
		batchMutex.Unlock()
	}
}