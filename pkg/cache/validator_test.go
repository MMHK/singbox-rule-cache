package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSRSFile(t *testing.T) {
	t.Run("valid SRS file", func(t *testing.T) {
		// Use the real SRS file downloaded from OverseasAI.list
		filePath := filepath.Join("testdata", "valid-overseas.srs")

		// Skip if test data doesn't exist
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Skip("test data file not found, run: curl -L -o pkg/cache/testdata/valid-overseas.srs https://raw.githubusercontent.com/mm-sam/OverseasAI.list/refs/heads/main/rule/Singbox/OverseasAI/OverseasAI.srs")
		}

		err := ValidateSRSFile(filePath)
		assert.NoError(t, err)
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty.srs")

		err := os.WriteFile(filePath, []byte{}, 0644)
		require.NoError(t, err)

		err = ValidateSRSFile(filePath)
		assert.ErrorIs(t, err, ErrInvalidMagicBytes)
	})

	t.Run("wrong magic bytes", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "wrong-magic.srs")

		// Create file with wrong magic bytes
		data := []byte{0x00, 0x01, 0x02, 0x01}
		err := os.WriteFile(filePath, data, 0644)
		require.NoError(t, err)

		err = ValidateSRSFile(filePath)
		assert.ErrorIs(t, err, ErrInvalidMagicBytes)
	})

	t.Run("unsupported version", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bad-version.srs")

		// Valid magic bytes but unsupported version (99)
		data := []byte{0x53, 0x52, 0x53, 99}
		err := os.WriteFile(filePath, data, 0644)
		require.NoError(t, err)

		err = ValidateSRSFile(filePath)
		assert.ErrorIs(t, err, ErrUnsupportedVersion)
	})

	t.Run("corrupted zlib data", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "corrupted-zlib.srs")

		// Valid header but corrupted compressed data
		data := []byte{0x53, 0x52, 0x53, 0x01, 0xFF, 0xFF, 0xFF, 0xFF}
		err := os.WriteFile(filePath, data, 0644)
		require.NoError(t, err)

		err = ValidateSRSFile(filePath)
		assert.ErrorIs(t, err, ErrInvalidZlibData)
	})
}
