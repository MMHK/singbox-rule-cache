package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadMetadata holds information about a downloaded file
type DownloadMetadata struct {
	SourceURL        string    `json:"source_url"`
	LocalFile        string    `json:"local_file"`
	DownloadedAt     time.Time `json:"downloaded_at"`
	FileSize         int64     `json:"file_size"`
	ETag             string    `json:"etag,omitempty"`
	LastModified     time.Time `json:"last_modified,omitempty"`
	ValidationStatus string    `json:"validation_status"`
}

// Downloader handles HTTP downloads with retry and validation
type Downloader struct {
	client     *http.Client
	maxRetries int
	timeout    time.Duration
}

// NewDownloader creates a new Downloader instance
func NewDownloader(maxRetries int, timeoutSeconds int) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		maxRetries: maxRetries,
		timeout:    time.Duration(timeoutSeconds) * time.Second,
	}
}

// Download downloads a file from URL to local path with validation
func (d *Downloader) Download(ctx context.Context, url string, localPath string) (*DownloadMetadata, error) {
	slog.Info("starting download", "url", url, "local_path", localPath)

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	meta, err := d.downloadWithRetry(ctx, url, localPath)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	slog.Info("download completed successfully",
		"url", url,
		"file_size", meta.FileSize,
		"validation_status", meta.ValidationStatus)

	return meta, nil
}

// downloadWithRetry attempts download with exponential backoff
func (d *Downloader) downloadWithRetry(ctx context.Context, url string, localPath string) (*DownloadMetadata, error) {
	var lastErr error

	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			slog.Warn("retrying download",
				"url", url,
				"attempt", attempt+1,
				"max_retries", d.maxRetries+1,
				"backoff", backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		meta, err := d.doDownload(ctx, url, localPath)
		if err == nil {
			return meta, nil
		}

		lastErr = err

		if !isTransientError(err) {
			slog.Error("non-transient error, not retrying", "error", err)
			break
		}

		slog.Debug("transient error occurred", "error", err, "attempt", attempt+1)
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", d.maxRetries+1, lastErr)
}

// doDownload performs a single download attempt
func (d *Downloader) doDownload(ctx context.Context, url string, localPath string) (*DownloadMetadata, error) {
	req, err := d.createRequest(ctx, url, localPath)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	slog.Debug("received response", "status", resp.StatusCode, "url", url)

	if resp.StatusCode == http.StatusNotModified {
		return d.handleNotModified(localPath)
	}

	if err := d.checkStatusCode(resp); err != nil {
		return nil, err
	}

	tmpPath, written, err := d.downloadToTemp(resp.Body, localPath)
	if err != nil {
		return nil, err
	}

	if err := d.validateAndReplace(tmpPath, localPath); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}

	return d.buildMetadata(url, localPath, written, resp), nil
}

// createRequest creates an HTTP request with conditional headers
func (d *Downloader) createRequest(ctx context.Context, url string, localPath string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	existingMeta, _ := loadMetadata(localPath)
	if existingMeta != nil {
		d.addConditionalHeaders(req, existingMeta)
	}

	return req, nil
}

// addConditionalHeaders adds ETag and Last-Modified headers for conditional requests
func (d *Downloader) addConditionalHeaders(req *http.Request, meta *DownloadMetadata) {
	if meta.ETag != "" {
		req.Header.Set("If-None-Match", meta.ETag)
		slog.Debug("adding If-None-Match header", "etag", meta.ETag)
	}
	if !meta.LastModified.IsZero() {
		req.Header.Set("If-Modified-Since", meta.LastModified.Format(http.TimeFormat))
		slog.Debug("adding If-Modified-Since header", "time", meta.LastModified)
	}
}

// handleNotModified handles 304 Not Modified response
func (d *Downloader) handleNotModified(localPath string) (*DownloadMetadata, error) {
	existingMeta, _ := loadMetadata(localPath)
	if existingMeta == nil {
		return nil, fmt.Errorf("cached metadata not found")
	}
	slog.Info("resource not modified, using cached version", "local_path", localPath)
	existingMeta.ValidationStatus = "valid"
	return existingMeta, nil
}

// checkStatusCode validates HTTP response status code
func (d *Downloader) checkStatusCode(resp *http.Response) error {
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("client error %d: %s", resp.StatusCode, resp.Status)
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error %d: %s", resp.StatusCode, resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// downloadToTemp downloads response body to a temporary file
func (d *Downloader) downloadToTemp(body io.Reader, localPath string) (string, int64, error) {
	tmpFile, err := os.CreateTemp(filepath.Dir(localPath), "*.tmp")
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}

	written, err := io.Copy(tmpFile, body)
	if err != nil {
		cleanup()
		return "", 0, fmt.Errorf("write to temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", 0, fmt.Errorf("close temp file: %w", err)
	}

	return tmpPath, written, nil
}

// validateAndReplace validates SRS file and atomically replaces target
func (d *Downloader) validateAndReplace(tmpPath, localPath string) error {
	if err := ValidateSRSFile(tmpPath); err != nil {
		slog.Warn("SRS validation failed", "path", tmpPath, "error", err)
		return fmt.Errorf("validate SRS file: %w", err)
	}

	if err := os.Rename(tmpPath, localPath); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// buildMetadata constructs DownloadMetadata from response data
func (d *Downloader) buildMetadata(url, localPath string, fileSize int64, resp *http.Response) *DownloadMetadata {
	etag := resp.Header.Get("ETag")
	lastModifiedStr := resp.Header.Get("Last-Modified")
	var lastModified time.Time
	if lastModifiedStr != "" {
		lastModified, _ = time.Parse(http.TimeFormat, lastModifiedStr)
	}

	meta := &DownloadMetadata{
		SourceURL:        url,
		LocalFile:        localPath,
		DownloadedAt:     time.Now(),
		FileSize:         fileSize,
		ETag:             etag,
		LastModified:     lastModified,
		ValidationStatus: "valid",
	}

	if err := saveMetadata(meta); err != nil {
		slog.Warn("failed to save metadata", "error", err)
	}

	return meta
}

// saveMetadata saves download metadata to a JSON sidecar file
func saveMetadata(meta *DownloadMetadata) error {
	metaPath := meta.LocalFile + ".meta.json"
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("write metadata file: %w", err)
	}

	slog.Debug("metadata saved", "path", metaPath)
	return nil
}

// loadMetadata loads download metadata from a JSON sidecar file
func loadMetadata(localPath string) (*DownloadMetadata, error) {
	metaPath := localPath + ".meta.json"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read metadata file: %w", err)
	}

	var meta DownloadMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	slog.Debug("metadata loaded", "path", metaPath)
	return &meta, nil
}

// isTransientError checks if an error is transient and should be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	transientKeywords := []string{
		"connection refused",
		"no such host",
		"temporary failure",
		"i/o timeout",
		"connection reset",
	}

	for _, keyword := range transientKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
