package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	apiKey := os.Getenv("SCW_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("Error: env var SCW_SECRET_KEY is not set")
	}

	client, err := NewOpenAIClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("front")))
	mux.HandleFunc("/upload-face", handleUploadFace(client))

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, securityHeaders(mux)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
