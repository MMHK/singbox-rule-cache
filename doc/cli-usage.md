# CLI 使用指南

## 安裝

```bash
go build -o singbox-rule-cache.exe .
```

## 基本用法

### 查看所有命令

```bash
./singbox-rule-cache --help
```

### 全局標誌

- `--config string` - 配置文件路徑（默認：`./config.yaml`）
- `--cache-dir string` - 緩存目錄（覆蓋配置）
- `--verbose` - 啟用詳細日誌（Debug 級別）

## 子命令

### start - 啟動定時同步

啟動周期性同步器，持續運行直到按下 Ctrl+C。

```bash
# 使用配置文件中的間隔
./singbox-rule-cache start

# 覆蓋同步間隔為 30 分鐘
./singbox-rule-cache start --interval 30

# 使用自定義配置文件
./singbox-rule-cache --config my-config.yaml start
```

**功能特點**：
- 自動按照配置的間隔定期同步所有啟用的規則
- 支持條件請求（ETag/If-Modified-Since），避免不必要下載
- 併發下載多個規則，提高效率
- 下載後自動驗證 SRS 文件完整性
- 原子文件替換，確保外部訪問不會看到不完整文件
- 按 Ctrl+C 優雅關閉

### sync - 手動觸發同步

立即執行一次同步操作，完成後退出。

```bash
# 同步所有啟用的規則
./singbox-rule-cache sync

# 只同步指定規則
./singbox-rule-cache sync geoip-cn

# 使用自定義緩存目錄
./singbox-rule-cache --cache-dir /tmp/cache sync
```

**功能特點**：
- 一次性同步，適合手動更新或測試
- 可以選擇性同步單個規則
- 同樣支持條件請求和自動驗證

### validate - 驗證緩存文件

檢查緩存的 SRS 文件完整性。

```bash
# 驗證所有緩存文件
./singbox-rule-cache validate

# 驗證指定文件
./singbox-rule-cache validate ./cache/geoip-cn.srs
```

**驗證內容**：
- Magic Bytes（必須是 "SRS"）
- 版本號（支持 1-5）
- zlib 壓縮數據完整性
- 規則數量解析

### list - 列出緩存狀態

顯示所有規則的緩存狀態。

```bash
./singbox-rule-cache list
```

輸出示例：
```
NAME                     URL                                               EXISTS  SIZE      LAST UPDATED           STATUS
----                     ---                                               ------  ----      ------------           ------
overseas-ai              https://raw.githubusercontent.com/...             Yes     4.9 KiB   2026-10-03T14:00:00Z   valid
ip-private               https://testingcf.jsdelivr.net/gh/...             Yes     144 B     2026-10-03T14:00:00Z   valid
site-private             https://testingcf.jsdelivr.net/gh/...             Yes     696 B     2026-10-03T14:00:00Z   valid
category-ads-all         https://testingcf.jsdelivr.net/gh/...             Yes     7.3 KiB   2026-10-03T14:00:00Z   valid
accelerated-domains-china https://raw.githubusercontent.com/...            Yes     432.8 KiB 2026-10-03T14:00:00Z   valid
ip-cn                    https://testingcf.jsdelivr.net/gh/...             Yes     35.6 KiB  2026-10-03T14:00:00Z   valid
site-cn                  https://testingcf.jsdelivr.net/gh/...             Yes     435.5 KiB 2026-10-03T14:00:00Z   valid
geolocation-not-cn       https://testingcf.jsdelivr.net/gh/...             Yes     162.3 KiB 2026-10-03T14:00:00Z   valid
```

## 配置文件示例

### 基礎配置

創建 `config.yaml`：

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
    url: "https://example.com/geoip-cn.srs"
    local_file: "geoip-cn.srs"
    enabled: true
    
  - name: "geosite-private"
    url: "https://example.com/geosite-private.srs"
    local_file: "geosite-private.srs"
    enabled: true
```

### E2E 測試配置示例

使用真實的 SRS URL（來自 `config.e2e.yaml`）：

```yaml
cache:
  dir: "./cache"
  max_retries: 3
  timeout_seconds: 30

sync:
  interval_minutes: 60
  auto_start: false

rules:
  - name: "overseas-ai"
    url: "https://raw.githubusercontent.com/mm-sam/OverseasAI.list/refs/heads/main/rule/Singbox/OverseasAI/OverseasAI.srs"
    local_file: "OverseasAI.srs"
    enabled: true
    
  - name: "ip-private"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/private.srs"
    local_file: "ip-private.srs"
    enabled: true
    
  - name: "site-private"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/private.srs"
    local_file: "site-private.srs"
    enabled: true
    
  - name: "category-ads-all"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/category-ads-all.srs"
    local_file: "category-ads-all.srs"
    enabled: true
    
  - name: "accelerated-domains-china"
    url: "https://raw.githubusercontent.com/Dreista/sing-box-rule-set-cn/rule-set/accelerated-domains.china.conf.srs"
    local_file: "accelerated-domains.china.conf.srs"
    enabled: true
    
  - name: "ip-cn"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/cn.srs"
    local_file: "ip-cn.srs"
    enabled: true
    
  - name: "site-cn"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/cn.srs"
    local_file: "site-cn.srs"
    enabled: true
    
  - name: "geolocation-not-cn"
    url: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/geolocation-!cn.srs"
    local_file: "geolocation-!cn.srs"
    enabled: true
```

## 環境變量

可以使用環境變量覆蓋配置（優先級：環境變量 > 配置文件 > 默認值）：

| 環境變量 | 配置路徑 | 說明 | 默認值 |
|---------|---------|------|--------|
| `CACHE_DIR` | `cache.dir` | 本地緩存目錄路徑 | `./cache` |
| `DOWNLOAD_TIMEOUT` | `cache.timeout_seconds` | HTTP 下載超時時間（秒） | `30` |
| `MAX_RETRIES` | `cache.max_retries` | 最大重試次數 | `3` |
| `SYNC_INTERVAL` | `sync.interval_minutes` | 同步間隔（分鐘） | `60` |
| `AUTO_START` | `sync.auto_start` | 是否自動啟動定時同步 | `false` |

使用示例：

```bash
export CACHE_DIR=/var/cache/singbox
export DOWNLOAD_TIMEOUT=60
export MAX_RETRIES=5
export SYNC_INTERVAL=30

./singbox-rule-cache start
```

## 信號處理

程序支持優雅關閉：
- 按 `Ctrl+C` 或發送 `SIGTERM` 信號
- 等待當前下載完成
- 清理資源後退出

## 日誌級別

- **默認**：Info 級別，顯示重要信息
- **Verbose**：使用 `--verbose` 標誌啟用 Debug 級別，顯示詳細調試信息

```bash
./singbox-rule-cache --verbose sync
```

## 高級特性

### 原子文件替換

下載過程使用臨時文件，驗證通過後才原子替換目標文件，確保：
- 外部消費者（如 sing-box）永遠不會看到不完整的文件
- 即使在更新過程中斷電或崩潰，也不會損壞現有緩存

### 條件請求優化

支持 HTTP 條件請求：
- 使用 ETag（`If-None-Match`）
- 使用 Last-Modified（`If-Modified-Since`）
- 服務器返回 304 Not Modified 時，跳過下載，節省帶寬

### 併發下載

- 所有啟用的規則同時下載，提高效率
- 單個規則失敗不影響其他規則（graceful degradation）
- 使用 atomic.Int64 確保併發安全

### SRS 文件驗證

每次下載後自動驗證：
1. Magic Bytes 檢查（必須是 0x53, 0x52, 0x53）
2. 版本號驗證（支持 1-5）
3. zlib 解壓測試
4. 規則數量解析

驗證失敗的文件會被標記並可通過 `CleanInvalid` 清理。

## 常見問題

### Q: 如何查看哪些規則下載失敗了？

A: 使用 `--verbose` 標誌可以看到詳細的下載日誌，包括失敗的規則和錯誤原因。

### Q: 緩存文件存儲在哪裡？

A: 默認在 `./cache/` 目錄下，可以通過 `--cache-dir` 標誌或 `CACHE_DIR` 環境變量修改。

### Q: 如何清理無效的緩存文件？

A: 目前需要手動刪除，未來版本可能會添加 `clean` 子命令。

### Q: 支持多少個規則同時下載？

A: 理論上無限制，但建議根據網絡帶寬和系統資源合理配置。當前實現沒有併發數限制。
