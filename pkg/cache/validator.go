package cache

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// SRS validation errors
var (
	ErrInvalidMagicBytes  = errors.New("invalid SRS magic bytes")
	ErrUnsupportedVersion = errors.New("unsupported SRS version")
	ErrInvalidZlibData    = errors.New("invalid zlib compressed data")
	ErrInvalidRuleCount   = errors.New("invalid rule count")
)

const (
	srsMagicByte1 = 0x53 // 'S'
	srsMagicByte2 = 0x52 // 'R'
	srsMagicByte3 = 0x53 // 'S'
	minHeaderSize = 4    // 3 bytes magic + 1 byte version
	maxVersion    = 5
	minVersion    = 1
)

// ValidateSRSFile validates the integrity and validity of an SRS file
func ValidateSRSFile(filePath string) error {
	slog.Debug("starting SRS file validation", "path", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if len(data) < minHeaderSize {
		slog.Warn("file too small to be valid SRS", "size", len(data))
		return ErrInvalidMagicBytes
	}

	if err := validateMagicBytes(data); err != nil {
		return err
	}

	version := data[3]
	if err := validateVersion(version); err != nil {
		return err
	}

	compressed := data[minHeaderSize:]
	if len(compressed) == 0 {
		slog.Warn("no compressed data found")
		return ErrInvalidZlibData
	}

	ruleCount, err := decompressAndValidate(compressed)
	if err != nil {
		return err
	}

	slog.Info("SRS file validation successful",
		"path", filePath,
		"version", version,
		"rule_count", ruleCount,
		"data_size", len(data))

	return nil
}

// validateMagicBytes checks the first 3 bytes match SRS magic
func validateMagicBytes(data []byte) error {
	if data[0] != srsMagicByte1 || data[1] != srsMagicByte2 || data[2] != srsMagicByte3 {
		slog.Warn("invalid magic bytes detected",
			"got", []byte{data[0], data[1], data[2]},
			"expected", []byte{srsMagicByte1, srsMagicByte2, srsMagicByte3})
		return ErrInvalidMagicBytes
	}
	slog.Debug("magic bytes validated successfully")
	return nil
}

// validateVersion checks if the version is supported
func validateVersion(version byte) error {
	if version < minVersion || version > maxVersion {
		slog.Warn("unsupported SRS version", "version", version)
		return fmt.Errorf("%w: got %d, supported %d-%d",
			ErrUnsupportedVersion, version, minVersion, maxVersion)
	}
	slog.Debug("version validated", "version", version)
	return nil
}

// decompressAndValidate decompresses zlib data and reads rule count
func decompressAndValidate(compressed []byte) (uint64, error) {
	zlibReader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		slog.Warn("failed to create zlib reader", "error", err)
		return 0, fmt.Errorf("%w: %v", ErrInvalidZlibData, err)
	}
	defer zlibReader.Close()

	decompressed, err := io.ReadAll(zlibReader)
	if err != nil {
		slog.Warn("failed to decompress data", "error", err)
		return 0, fmt.Errorf("%w: %v", ErrInvalidZlibData, err)
	}

	slog.Debug("decompression successful", "compressed_size", len(compressed), "decompressed_size", len(decompressed))

	if len(decompressed) == 0 {
		return 0, ErrInvalidRuleCount
	}

	ruleCount, n := binary.Uvarint(decompressed)
	if n <= 0 {
		slog.Warn("failed to read rule count", "bytes_read", n)
		return 0, ErrInvalidRuleCount
	}

	slog.Debug("rule count extracted", "count", ruleCount, "bytes_consumed", n)
	return ruleCount, nil
}
