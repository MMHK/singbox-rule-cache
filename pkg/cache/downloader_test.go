package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloader_SuccessfulDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "test.srs")

	downloader := NewDownloader(3, 30)
	ctx := context.Background()

	url := "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/private.srs"

	meta, err := downloader.Download(ctx, url, localPath)
	require.NoError(t, err, "download should succeed")
	require.NotNil(t, meta)

	assert.Equal(t, url, meta.SourceURL)
	assert.Equal(t, localPath, meta.LocalFile)
	assert.Equal(t, "valid", meta.ValidationStatus)
	assert.Greater(t, meta.FileSize, int64(0))
	assert.NotZero(t, meta.DownloadedAt)

	fileInfo, err := os.Stat(localPath)
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0))

	metaExists, err := os.Stat(localPath + ".meta.json")
	require.NoError(t, err)
	assert.Greater(t, metaExists.Size(), int64(0))
}

func TestDownloader_ConditionalRequestNotModified(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "test-conditional.srs")

	downloader := NewDownloader(3, 30)
	ctx := context.Background()

	url := "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/private.srs"

	firstMeta, err := downloader.Download(ctx, url, localPath)
	require.NoError(t, err, "first download should succeed")

	time.Sleep(1 * time.Second)

	secondMeta, err := downloader.Download(ctx, url, localPath)
	require.NoError(t, err, "second download should succeed or return cached")

	assert.Equal(t, firstMeta.ETag, secondMeta.ETag)
	assert.Equal(t, firstMeta.FileSize, secondMeta.FileSize)
}

func TestDownloader_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "invalid.srs")

	downloader := NewDownloader(1, 5)
	ctx := context.Background()

	url := "http://invalid.domain.that.does.not.exist/test.srs"

	_, err := downloader.Download(ctx, url, localPath)
	assert.Error(t, err, "download should fail for invalid URL")
	assert.Contains(t, err.Error(), "download failed")
}

func TestDownloader_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "notfound.srs")

	downloader := NewDownloader(1, 10)
	ctx := context.Background()

	url := "https://httpbin.org/status/404"

	_, err := downloader.Download(ctx, url, localPath)
	assert.Error(t, err, "download should fail for 404")
	assert.Contains(t, err.Error(), "client error 404")
}

func TestDownloader_SaveAndLoadMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "metadata-test.srs")

	meta := &DownloadMetadata{
		SourceURL:        "https://example.com/test.srs",
		LocalFile:        localPath,
		DownloadedAt:     time.Now(),
		FileSize:         12345,
		ETag:             "\"abc123\"",
		LastModified:     time.Now().Add(-1 * time.Hour),
		ValidationStatus: "valid",
	}

	err := saveMetadata(meta)
	require.NoError(t, err)

	loadedMeta, err := loadMetadata(localPath)
	require.NoError(t, err)
	require.NotNil(t, loadedMeta)

	assert.Equal(t, meta.SourceURL, loadedMeta.SourceURL)
	assert.Equal(t, meta.LocalFile, loadedMeta.LocalFile)
	assert.Equal(t, meta.FileSize, loadedMeta.FileSize)
	assert.Equal(t, meta.ETag, loadedMeta.ETag)
	assert.Equal(t, meta.ValidationStatus, loadedMeta.ValidationStatus)
}
