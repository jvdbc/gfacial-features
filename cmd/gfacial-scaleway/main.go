package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jvdbc/gfacial-features/internal/api"
	"github.com/jvdbc/gfacial-features/internal/handler"
)

func main() {
	apiKey := os.Getenv("SCW_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("Error: env var SCW_SECRET_KEY is not set")
	}

	client, err := api.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("front")))
	mux.HandleFunc("/upload-face", handler.NewUploadFaceHandler(client))

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, handler.SecurityHeaders(mux)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
