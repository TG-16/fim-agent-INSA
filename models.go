package main

import (
	"sync"
)

type FimLog struct {
	Timestamp   string `json:"@timestamp"`
	FilePath    string `json:"file_path"`
	Event       string `json:"event_type"`
	Hash        string `json:"hash_sha256,omitempty"`
	Size        int64  `json:"file_size_bytes"`
	Permissions string `json:"permissions"`
	Owner       string `json:"owner"`
}

var (
	appConfig  Config
	batchMutex sync.Mutex
	logBuffer  []FimLog
	metaCache  sync.Map
)
