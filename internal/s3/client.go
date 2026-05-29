package s3

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	BucketName = "gfacial-images"
	Region     = "fr-par"
)

type Client struct {
	httpClient     *http.Client
	bucketURL      string
	accessKey      string
	secretKey      string
	debugSignature bool
}

func NewClient(accessKey, secretKey string, debug bool) *Client {
	return &Client{
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		bucketURL:      fmt.Sprintf("https://%s.s3.%s.scw.cloud", BucketName, Region),
		accessKey:      accessKey,
		secretKey:      secretKey,
		debugSignature: debug,
	}
}

func (c *Client) Upload(ctx context.Context, data []byte, filename string) (string, error) {
	objectKey := sanitizeObjectKey(filename)
	url := c.bucketURL + "/" + objectKey

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "image/jpeg")
	payloadSum := sha256.Sum256(data)
	payloadHash := hex.EncodeToString(payloadSum[:])

	now := time.Now().UTC()
	if err := c.signRequest(req, payloadSum[:], now); err != nil {
		return "", fmt.Errorf("failed to sign upload: %w", err)
	}

	canonicalRequest, stringToSign := c.buildDebugSignatureData(req, payloadHash, now)
	c.logSignatureDebug("upload", req, payloadHash, canonicalRequest, stringToSign)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.logErrorDebug("upload", req, resp, body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return url, nil
}

func (c *Client) DownloadToTemp(ctx context.Context, s3URL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s3URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	tmpDir := "/tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tmp dir: %w", err)
	}

	filename := filepath.Join(tmpDir, filepath.Base(s3URL))
	tmpFile, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return filename, nil
}

func (c *Client) buildDebugSignatureData(req *http.Request, payloadHash string, now time.Time) (string, string) {
	date := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", host, payloadHash, date)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s", req.Method, canonicalURI, canonicalHeaders, signedHeaders, payloadHash)
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	canonicalRequestHashHex := hex.EncodeToString(canonicalRequestHash[:])

	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, Region)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", algorithm, date, credentialScope, canonicalRequestHashHex)

	return canonicalRequest, stringToSign
}

func sanitizeObjectKey(filename string) string {
	base := strings.ToLower(filepath.Base(filename))
	ext := strings.ToLower(filepath.Ext(base))
	name := strings.TrimSuffix(base, ext)

	sanitize := func(s string) string {
		re := regexp.MustCompile(`[^a-z0-9._-]+`)
		s = re.ReplaceAllString(strings.ToLower(s), "-")
		s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
		return strings.Trim(s, "-._")
	}

	name = sanitize(name)
	if name == "" {
		name = "upload"
	}

	ext = sanitize(strings.TrimPrefix(ext, "."))
	if ext == "" {
		ext = "jpg"
	}

	return name + "." + ext
}

func (c *Client) Delete(s3URL string) error {
	req, err := http.NewRequest(http.MethodDelete, s3URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()
	payloadSum := sha256.Sum256(nil)
	payloadHash := hex.EncodeToString(payloadSum[:])
	if err := c.signRequest(req, payloadSum[:], now); err != nil {
		return fmt.Errorf("failed to sign delete: %w", err)
	}

	canonicalRequest, stringToSign := c.buildDebugSignatureData(req, payloadHash, now)
	c.logSignatureDebug("delete", req, payloadHash, canonicalRequest, stringToSign)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.logErrorDebug("delete", req, resp, body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) signRequest(req *http.Request, payloadHash []byte, now time.Time) error {
	req.Host = req.URL.Host
	payloadHashHex := hex.EncodeToString(payloadHash)
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHashHex)

	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", host, payloadHashHex, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s", req.Method, canonicalURI, canonicalHeaders, signedHeaders, payloadHashHex)
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, Region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, credentialScope, hex.EncodeToString(canonicalRequestHash[:]))
	signingKey := getSignatureKey(c.secretKey, dateStamp, Region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	authorization := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.accessKey,
		credentialScope,
		signedHeaders,
		signature,
	)

	req.Header.Set("Authorization", authorization)
	return nil
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(data))
	return h.Sum(nil)
}

func getSignatureKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func (c *Client) logSignatureDebug(operation string, req *http.Request, payloadHash, canonicalRequest, stringToSign string) {
	if !c.debugSignature {
		return
	}

	log.Printf("[s3:%s] url=%s", operation, req.URL.String())
	log.Printf("[s3:%s] path=%s escapedPath=%s payloadSHA256=%s", operation, req.URL.Path, req.URL.EscapedPath(), payloadHash)
	log.Printf("[s3:%s] canonicalRequest:\n%s", operation, canonicalRequest)
	log.Printf("[s3:%s] stringToSign:\n%s", operation, stringToSign)
}

func (c *Client) logErrorDebug(operation string, req *http.Request, resp *http.Response, body []byte) {
	if !c.debugSignature {
		return
	}

	log.Printf("[s3:%s] status=%d requestID=%s hostID=%s", operation, resp.StatusCode, resp.Header.Get("x-amz-request-id"), resp.Header.Get("x-amz-id-2"))
	log.Printf("[s3:%s] responseBody=%s", operation, string(body))
	log.Printf("[s3:%s] authorization=%s", operation, req.Header.Get("Authorization"))
	log.Printf("[s3:%s] x-amz-date=%s", operation, req.Header.Get("x-amz-date"))
	log.Printf("[s3:%s] x-amz-content-sha256=%s", operation, req.Header.Get("x-amz-content-sha256"))
}
