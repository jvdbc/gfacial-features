package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jvdbc/gfacial-features/internal/handler"
	"github.com/jvdbc/gfacial-features/internal/openapi"
	"github.com/jvdbc/gfacial-features/internal/s3"
)

func main() {
	accessKey := os.Getenv("SCW_ACCESS_KEY")
	if accessKey == "" {
		log.Fatal("Error: env var SCW_ACCESS_KEY is not set")
	}

	secretKey := os.Getenv("SCW_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("Error: env var SCW_SECRET_KEY is not set")
	}

	oapiCli := openapi.NewClient(secretKey)

	debug := false
	debugMode := os.Getenv("GFACIAL_SCALEWAY_DEBUG")
	if strings.ToLower(strings.TrimSpace(debugMode)) == "1" || strings.ToLower(strings.TrimSpace(debugMode)) == "true" {
		log.Println("Debug mode enabled for Scaleway client")
		debug = true
	}

	s3Cli := s3.NewClient(accessKey, secretKey, debug)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("front")))
	mux.HandleFunc("/upload-face", handler.NewUploadFaceHandler(oapiCli, s3Cli))

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, handler.SecurityHeaders(mux)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
