# GitHub Action Docker Build 規劃

> **⚠️ AI 生成聲明**：本文件由 AI 輔助生成，僅供參考。所有內容必須經過**人類 Review + 確認**後才能進入實施階段。未經確認前，不得開始任何開發工作。

**建立日期**: 2026-10-03
**狀態**: 🟢 完成
**優先級**: Medium

---

## 背景與目標

建立 GitHub Action 自動建構 Docker image，支援：
- Push to main branch 時自動建構並推送 `latest` 標籤
- Push tag 時自動建構並推送版本標籤（如 `v1.0.0`）
- Pull request 時建構但不推送（驗證用）
- 多架構支援（amd64, arm64）
- 使用 Docker Hub 作為映像倉庫

## 任務列表

| # | 任務 | 狀態 | Sub Agent | 備註 |
|---|------|------|-----------|------|
| 1 | 建立 GitHub Action workflow | ✅ 完成 | general | |
| 2 | Code Review | ✅ 完成 | code-reviewer | |

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

#### GitHub Action Workflow
- **觸發條件**：
  - `push` to `main` branch → 建構並推送 `latest`
  - `push` tag `v*` → 建構並推送版本標籤 + `latest`
  - `pull_request` to `main` → 只建構不推送（驗證用）
- **多架構支援**：`linux/amd64`, `linux/arm64`
- **Docker Hub 認證**：使用 GitHub Secrets (`DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`)
- **快取優化**：使用 GitHub Actions cache 加速建構

### 前置需求

需要在 GitHub Repository Settings → Secrets and variables → Actions 中設定：
- `DOCKERHUB_USERNAME` — Docker Hub 使用者名稱
- `DOCKERHUB_TOKEN` — Docker Hub access token（不是密碼）

### 使用方式

1. Push to main branch → 自動建構並推送 `mmhk/singbox-rule-cache:latest`
2. Create tag `v1.0.0` → 自動建構並推送 `mmhk/singbox-rule-cache:v1.0.0` + `latest`
3. Open pull request → 自動建構驗證（不推送）
