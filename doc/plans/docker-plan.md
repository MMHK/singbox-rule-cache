# Docker 化規劃

> **⚠️ AI 生成聲明**：本文件由 AI 輔助生成，僅供參考。所有內容必須經過**人類 Review + 確認**後才能進入實施階段。未經確認前，不得開始任何開發工作。

**建立日期**: 2026-10-03
**狀態**: 🟢 完成
**優先級**: Medium

---

## 背景與目標

將 singbox-rule-cache 專案 Docker 化，支援：
- 多階段建構，產生最小運行映像
- 獨立 build service 與 run service
- 映像路徑：`mmhk/singbox-rule-cache`
- 支援推送到 Docker Hub

## 任務列表

| # | 任務 | 狀態 | Sub Agent | 備註 |
|---|------|------|-----------|------|
| 1 | 建立 Dockerfile（多階段建構） | ✅ 完成 | general | |
| 2 | 建立 docker-compose.yml（build + run service） | ✅ 完成 | general | |
| 3 | 建立 .dockerignore | ✅ 完成 | general | |
| 4 | Code Review | ✅ 完成 | code-reviewer | |

### 狀態圖示說明
- ⬜ 待處理 — 尚未開始
- 🔄 進行中 — sub agent 執行中
- ✅ 完成 — 已實作並通過 review
- 🔴 阻塞 — 有依賴或問題需解決
- ⏭️ 跳過 — 不需要，附原因說明

## 人類 Review 記錄

| 日期 | 動作 | 備註 |
|------|------|------|
| 2026-10-03 | ✅ 已批准 | 用戶確認規劃並要求開始實作 |

## Review 日誌

| 日期 | Reviewer | 發現 | 處理方式 |
|------|----------|------|----------|
| 2026-10-03 | code-reviewer | 無問題 | — |

## 備註

### 實作細節

#### 1. Dockerfile
- 多階段建構：builder (golang:1.24-alpine) → runtime (alpine:3.21)
- 靜態連結：`CGO_ENABLED=0`
- 優化體積：`-ldflags="-s -w"`
- 需要 ca-certificates 處理 HTTPS

#### 2. docker-compose.yml
- `app-build` service：建構映像，使用 `profiles: ["build"]` 控制
- `singbox-rule-cache` service：運行服務，掛載 config.yaml 和 cache volume
- 映像路徑：`mmhk/singbox-rule-cache:latest`

#### 3. .dockerignore
- 排除：.git, cache/, *.exe, *.srs, doc/, .env 等

### 使用方式

```bash
# 建構映像
docker compose build app-build

# 啟動服務
docker compose up -d

# 推送到 Docker Hub
docker build -t mmhk/singbox-rule-cache:latest .
docker push mmhk/singbox-rule-cache:latest
```
