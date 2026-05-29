package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jvdbc/gfacial-features/internal/openapi"
	"github.com/jvdbc/gfacial-features/internal/s3"
)

const (
	maxUploadSize = 5 * 1024 * 1024
	uploadTimeout = 30 * time.Second
)

type UploadHandler struct {
	apiClient *openapi.Client
	s3Client  *s3.Client
}

func NewUploadFaceHandler(apiClient *openapi.Client, s3Client *s3.Client) http.HandlerFunc {
	h := &UploadHandler{
		apiClient: apiClient,
		s3Client:  s3Client,
	}
	return h.handle
}

func (h *UploadHandler) handle(w http.ResponseWriter, r *http.Request) {
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

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusBadRequest)
		return
	}

	s3URL, err := h.s3Client.Upload(r.Context(), data, handler.Filename)
	if err != nil {
		log.Printf("S3 upload error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to upload to S3: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := h.s3Client.Delete(s3URL); err != nil {
			log.Printf("Failed to delete S3 object: %v", err)
		}
	}()

	tmpPath, err := h.s3Client.DownloadToTemp(r.Context(), s3URL)
	if err != nil {
		log.Printf("S3 download error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to download from S3: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpPath)

	ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
	defer cancel()

	start := time.Now()
	result, err := h.apiClient.AnalyzeFace(ctx, tmpPath)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("Analysis error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to analyze image: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"filename":"%s","analysis_time":"%.3fs","result":%s}`, handler.Filename, elapsed.Seconds(), result)
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}
