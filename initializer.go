package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"net/http"
	"strings"
	"time"
)

func ensureIndexExists() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	checkUrl := strings.TrimSuffix(appConfig.OSUrl, "/_bulk") + "/" + appConfig.IndexName
	checkReq, _ := http.NewRequest("HEAD", checkUrl, nil)
	checkReq.SetBasicAuth(appConfig.OSUser, appConfig.OSPass)
	resp, err := client.Do(checkReq)
	
	if err == nil && resp.StatusCode == 200 { return }

	mapping := `{"mappings":{"properties":{"@timestamp":{"type":"date"},"file_path":{"type":"keyword"},"event_type":{"type":"keyword"},"hash_sha256":{"type":"keyword"},"owner":{"type":"keyword"}}}}`
	req, _ := http.NewRequest("PUT", checkUrl, bytes.NewBuffer([]byte(mapping)))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(appConfig.OSUser, appConfig.OSPass)
	client.Do(req)
	log.Println("[INFO] OpenSearch Index initialized automatically.")
}