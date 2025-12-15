package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

type CacheMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	Key string `json:"key"`
	Size int64 `json:"size"`
}

func saveMetadata(key string, size int64) error {
	metadata := CacheMetadata {
		CreatedAt: time.Now(),
		Key: key,
		Size: size,
	}

	metaPath := filepath.Join("./cache-data", key+".meta")
	file, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(metadata)
}

func readMetadata(key string) (*CacheMetadata, error) {
	metaPath := filepath.Join("./cache-data", key+".meta")
	file, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}

	var metadata CacheMetadata
	json.NewDecoder(file).Decode(&metadata)
	return &metadata, err
}

func cleanUpExpiredFiles(retentionDays int) {
	cacheDir := "./cache-data"
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Println("ERROR: failed to read cache directory:", err)
		return
	}

	cutoffTime := time.Now().Add(-time.Duration(retentionDays)* 24 *time.Hour)
	deletedCount := 0

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) == ".meta" {
			continue
		}

		key := file.Name()
		if filepath.Ext(key) == ".gz" {
			key = key[:len(key)-3]
		}

		metadata, err := readMetadata(key)
		if err != nil {
			log.Printf("WARNING: No metadata for %s, skipping\n", err)
			continue
		}

		if metadata.CreatedAt.Before(cutoffTime) {
			datapath := filepath.Join(cacheDir +key+".gz")
			metapath := filepath.Join(cacheDir +key+".meta")

			if err := os.Remove(datapath); err != nil {
				log.Printf("ERROR: failed to delete %s\n", datapath)
			}

			if err := os.Remove(metapath); err != nil {
				log.Printf("ERROR: failed to delete %s\n", metapath)
			}

			deletedCount++
			log.Printf("JANITOR: Deleted expiry entry: %s (age: %v)\n", key, time.Since(metadata.CreatedAt))
		}
	}

	if deletedCount > 0 {
		log.Printf("JANITOR: Cleanup complete - delete %d expired entries\n", deletedCount)
	}
}

func startJanitor(retentionDays int, intervalHours int) {
	go func() {
		log.Printf("JANITOR: Started (retention: %d days, interval: %d hours)\n",retentionDays, intervalHours)

		cleanUpExpiredFiles(retentionDays)

		ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
		for range ticker.C {
			cleanUpExpiredFiles(retentionDays)
		}
	}()
}