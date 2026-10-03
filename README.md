# singbox-rule-cache

[![Build Docker Image](https://github.com/MMHK/singbox-rule-cache/actions/workflows/docker-build.yml/badge.svg)](https://github.com/MMHK/singbox-rule-cache/actions/workflows/docker-build.yml)
[![Docker Hub](https://img.shields.io/docker/pulls/mmhk/singbox-rule-cache?logo=docker&label=pulls)](https://hub.docker.com/r/mmhk/singbox-rule-cache)
[![Docker Image Version](https://img.shields.io/docker/v/mmhk/singbox-rule-cache?logo=docker&label=docker)](https://hub.docker.com/r/mmhk/singbox-rule-cache)
[![Go Version](https://img.shields.io/github/go-mod/go-version/MMHK/singbox-rule-cache?logo=go)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/v/release/MMHK/singbox-rule-cache?logo=github)](https://github.com/MMHK/singbox-rule-cache/releases)
[![License](https://img.shields.io/github/license/MMHK/singbox-rule-cache)](LICENSE)

A tool to cache [sing-box](https://sing-box.sagernet.org/) remote rule-set `.srs` files to local storage.

## Features

- 🚀 **Concurrent Download** - Download multiple SRS files simultaneously
- ✅ **SRS Validation** - Verify file integrity (magic bytes, version, zlib compression)
- 🔄 **Auto Sync** - Periodic synchronization with configurable intervals
- 💾 **Atomic Replacement** - Smooth file updates without corruption
- 🎯 **Conditional Requests** - ETag/If-Modified-Since support to save bandwidth
- ⚙️ **Fully Configurable** - YAML config + environment variable overrides
- 🖥️ **CLI Interface** - Simple commands: `start`, `sync`, `validate`, `list`

## Quick Start

### Docker (Recommended)

```bash
# Pull image
docker pull mmhk/singbox-rule-cache:latest

# Run with docker compose
docker compose up -d

# Or run directly
docker run -d \
  --name singbox-rule-cache \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v cache-data:/app/cache \
  -e SYNC_INTERVAL=60 \
  mmhk/singbox-rule-cache:latest start
```

### Build from Source

```bash
go build -o singbox-rule-cache.exe .
```

### Configure

Create `config.yaml`:

```yaml
cache:
  dir: "./cache"
  max_retries: 3
  timeout_seconds: 30

sync:
  interval_minutes: 60
  auto_start: false

rules:
  - name: "geoip-cn"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/cn.srs"
    local_file: "ip-cn.srs"
    enabled: true
```

See `config.example.yaml` for a complete example.

### Run

```bash
# Start periodic sync
./singbox-rule-cache start

# Manual sync
./singbox-rule-cache sync

# Validate cache
./singbox-rule-cache validate

# List cache status
./singbox-rule-cache list
```

## Documentation

- [CLI Usage Guide](doc/cli-usage.md) - Complete CLI documentation
- [Implementation Plan](doc/plans/srs-cache-sync-plan.md) - Detailed task plan and design

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run E2E tests
go test -run TestE2E
```

All 24 tests pass ✅

## License

MIT
