package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		//w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/cache/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[len("/cache/"):]

		if key == "" {
			if r.Method == "GET" {
				handleList(w, r)
				return
			}
		}

		switch r.Method {
		case "PUT":
			handlePut(w, r, key)
		case "GET":
			handleGet(w, r, key)
		case "HEAD":
			handleHead(w, r, key)
		case "DELETE":
			handleDelete(w, r, key)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePut(w http.ResponseWriter, r *http.Request, key string) {
	cacheDir := "./cache-data"
	_ = os.MkdirAll(cacheDir, 0755)

	filePath := filepath.Join(cacheDir, key)
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		log.Println("ERROR: failed to create file:", err)
	}
	defer file.Close()

	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		log.Println("ERROR: failed to write file:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("File stored successfully"))
	log.Printf("PUT /cache/%s - stored successfully\n", key)
}

func handleGet(w http.ResponseWriter, r *http.Request, key string) {
	filepath := filepath.Join("./cache-data", key)

	file, err := os.Open(filepath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		log.Printf("GET /cache/%s - not found\n", key)
		return
	}
	defer file.Close()

	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, file)
	if err != nil {
		log.Println("ERROR: error reading file:", err)
	}
	log.Printf("GET /cache/%s - retrieve successfully\n", key)
}

func handleHead(w http.ResponseWriter, r *http.Request, key string) {
	filepath := filepath.Join("./cache-data", key)
	_, err := os.Stat(filepath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("HEAD /cache/%s - not found\n", key)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Printf("HEAD /cache/%s - exists\n", key)
}

func handleDelete(w http.ResponseWriter, r *http.Request, key string) {
	filepath := filepath.Join("./cache-data", key)
	err := os.Remove(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			log.Printf("DELETE /cache/%s - not found\n", key)
			return
		}
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		log.Println("ERROR: Failed to delete:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("File deleted successfully"))
	log.Printf("DELETE /cache/%s - deleted successfully\n", key)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	cacheDir := "./cache-data"
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		http.Error(w, "Failed to read cache directory", http.StatusInternalServerError)
		log.Println("ERROR: Failed to read directory:", err)
		return
	}

	keys := []string{}
	for _, file := range files {
		if !file.IsDir() {
			keys = append(keys, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"keys":keys,
		"count":len(keys),
	})
	log.Printf("LIST /cache/ - returned %d keys\n", len(keys))
}
