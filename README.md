# singbox-rule-cache

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

### Build

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
