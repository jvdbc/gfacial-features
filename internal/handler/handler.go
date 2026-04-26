package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jvdbc/gfacial-features/internal/api"
)

const (
	uploadDir     = "/tmp"
	maxUploadSize = 5 * 1024 * 1024
	uploadTimeout = 30 * time.Second
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

func NewUploadFaceHandler(client *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File not found in form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if handler.Header.Get("Content-Type") != "image/jpeg" {
			http.Error(w, "Only JPEG files are allowed", http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create upload directory: %v", err), http.StatusInternalServerError)
			return
		}

		filename := filepath.Join(uploadDir, handler.Filename)
		dst, err := os.Create(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
		defer cancel()

		start := time.Now()
		result, err := client.AnalyzeFace(ctx, filename)
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("Analysis error: %v", err)
			http.Error(w, fmt.Sprintf("Failed to analyze image: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"filename":"%s","analysis_time":"%.3fs","result":%s}`, handler.Filename, elapsed.Seconds(), result)
	}
}
