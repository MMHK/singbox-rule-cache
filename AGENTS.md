# AGENTS.md — singbox-rule-cache

This document is the behavior contract for this project. All AI-generated code MUST strictly follow these rules.
Violating any mandatory rule → rewrite and fix immediately.

---

## 0. Project Overview

**singbox-rule-cache** is a tool to cache [sing-box](https://sing-box.sagernet.org/) remote rule-set `.srs` (binary format) files to local storage.

### Core Features
- **Remote SRS Download**: Fetch `.srs` files from remote URLs with concurrent downloads
- **SRS Validation**: Verify file integrity (magic bytes, version, zlib compression, rule parsing)
- **Local Caching**: Store downloaded files with metadata tracking in JSON sidecar files
- **Cache Management**: Update, invalidate, and list cached rule-sets
- **Periodic Sync**: Automatic synchronization with configurable intervals
- **Atomic File Replacement**: Download to temp file first, then atomically rename to prevent corruption
- **Conditional Requests**: Support ETag/If-Modified-Since to avoid unnecessary downloads
- **CLI Interface**: Command-line tool with `start`, `sync`, `validate`, `list` subcommands
- **Graceful Degradation**: Individual rule failures don't affect other rules

### Tech Stack
- **Go 1.24+** (standard library preferred)
- **Key concepts**: HTTP download, file I/O, caching strategies, sing-box rule-set format, concurrency safety

---

## 1. Core Principles (Highest Priority, Violation = Rewrite)

1. **No over-engineering**
   - Unless the requirement explicitly states "high scalability", "multi-team maintenance", "DDD/microservices/clean architecture" — always use the simplest, most understandable approach.
   - No premature abstraction, layering, future-proofing code, or early design patterns.

2. **Follow existing patterns**
   - New code must match existing code style, naming, and structure.
   - Before starting a new task, check if there are any implementation plans or docs in the repo.

---

## 2. Directory Structure (Established, No New Subdirectories)

This project has an established directory structure. **Do not add new subdirectories or reorganize existing ones.**

```
.
├── main.go                      # Entry point (package main): init, CLI setup
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
│
├── pkg/                         # [ESTABLISHED] Core business modules
│   ├── cache/                   # Cache management: download, store, metadata
│   ├── config/                  # Config loading, validation
│   └── cli/                     # CLI commands and flags
│
└── doc/                         # Technical docs (only allowed doc directory)
    ├── spec.md                  # Project spec and architecture (if exists)
    └── plans/                   # Implementation plans (if exists)
```

### Directory Rules
- **Forbidden**: `internal/`, `domain/`, `application/`, `infrastructure/`, `adapter/`, `lib/`, `common/`, `layers/`, `modules/`, `features/`
- **No new subdirectories under `pkg/`** except those already listed — existing 3 subdirs are sufficient
- Prefer placing new files in existing subdirectories rather than root
- Keep the structure flat and simple

---

## 3. External Dependencies (Allowlist)

### Approved Dependencies (No unlisted additions)

| Package | Purpose | Notes |
|---------|---------|-------|
| `github.com/spf13/cobra` | CLI framework | Optional, only if needed |
| `gopkg.in/yaml.v3` | YAML parsing | For config files |
| `github.com/stretchr/testify` | Test assertions | Test-only |

### Preferred Stdlib (Use These First)
- `net/http` — HTTP downloads
- `os`, `io`, `path/filepath` — File operations
- `encoding/json` — Metadata serialization
- `log/slog` — Logging
- `context` — Request cancellation
- `time` — Timeouts, TTL

### Dependency Rules
- **New dependencies require explicit user approval**
- Prefer stdlib over external packages
- Before adding a dependency, check if existing deps already provide the same functionality
- Test-only deps (testify) are limited to `*_test.go` files

---

## 4. Code Style Rules

### Function Length
- **New code**: Functions/methods ≤ 40 lines (split if exceeded)
- When splitting, only create 2–3 smaller functions; no over-splitting

### Struct Fields
- General structs ≤ 8 fields
- Config-type structs may have more fields since they map to environment variables or config files
- When exceeding 15 fields, ask the user whether grouping is needed

### Error Handling
- Use basic patterns only: `if err != nil` / `return err` / `return fmt.Errorf("...: %w", err)`
- No custom multi-layer error types / error wrappers unless the existing code already uses them
- Non-fatal errors: use `slog.Warn()` + graceful degradation

### Logging
- **New code must use stdlib `log/slog`**
  - `slog.Debug()` / `slog.Info()` / `slog.Warn()` / `slog.Error()`
- Do not use logrus or other logging frameworks

### JSON / Serialization
- JSON tags use `snake_case` (e.g., `source_url`, `last_updated`)
- Use `camelCase` only when external APIs require it

### Naming Conventions
- Go identifiers follow Go conventions (`CamelCase` for exported, `camelCase` for unexported)
- Environment variables use `UPPER_SNAKE_CASE`
- Configuration keys use `snake_case`

---

## 5. Testing Rules

### Core Principle
- **No bulk test suite generation**
  Unless the requirement explicitly states "write complete tests from scratch" or "coverage ≥ 90%", add at most 3–5 test cases per change.

### Incremental Test Modification (Order is mandatory)
1. Check if `*_test.go` files already exist in context
2. If yes → only **append** new test functions at the end, or fix existing failing tests
3. If context says "happy path is already covered" → only add edge case / error case
4. If the module has no tests at all → only then create a new test file, still limited to 3–5 cases

### Test Style
- Each test ≤ 20 lines (project uses table-driven tests, continue to allow)
- Use `testify/assert` and `testify/require` (project standard)
- When API keys or network access are required, use `t.Skip()` if unavailable

### Test Coverage Gaps
The following modules should have tests prioritized:
- `pkg/cache/` — download and cache operations
- `pkg/config/` — configuration loading and validation

---

## 6. Documentation Rules

### No New Standalone Doc Files
Unless explicitly requested, do not create:
- README.md (new version)
- docs/, architecture.md, CHANGELOG.md, CONTRIBUTING.md

### doc/ Directory Convention
- `doc/` is the only allowed documentation directory
- After code changes that affect existing interfaces or behavior, update the corresponding `doc/` documentation

### Documentation Principles
- godoc per function: max 6 lines (function + params + return + brief example)
- Only add comments within existing files, or modify a section of existing README

---

## 7. Architecture Patterns (Must Understand Before Modifying)

### 7.1 Cache Storage Strategy
- Cached files stored in configurable directory (default: `./cache/`)
- Metadata tracked in JSON sidecar files (e.g., `rules.srs.meta.json`)
- Metadata includes: source URL, download time, file size, ETag, last modified, validation status
- Sidecar file naming: `<filename>.meta.json` (e.g., `geoip-cn.srs.meta.json`)

### 7.2 HTTP Download Pattern
- Use `http.Client` with timeout (default: 30s)
- Support conditional requests via `If-None-Match` (ETag) and `If-Modified-Since`
- Retry on transient failures (max 3 attempts with exponential backoff: 1s, 2s, 4s)
- **Atomic file replacement**: Download to temp file first (`os.CreateTemp()`), validate, then `os.Rename()` to target
  - This ensures external consumers (e.g., sing-box reading via HTTPS) never see partial/corrupted files
  - Temp files use pattern: `<random>.tmp` in the same directory as target

### 7.3 SRS Validation
- Validate magic bytes: `[0x53, 0x52, 0x53]` ("SRS")
- Validate version: uint8, supported versions 1-5
- Decompress zlib data and verify integrity
- Read uvarint rule count for basic structure validation
- Use `bytes.NewReader()` for zlib reader (not custom implementation)
- Return clear error types: `ErrInvalidMagicBytes`, `ErrUnsupportedVersion`, `ErrInvalidZlibData`, `ErrInvalidRuleCount`

### 7.4 Graceful Degradation
- On download failure: log warning, skip file, continue with others
- On cache read failure: treat as cache miss, attempt fresh download
- On validation failure: mark as invalid, allow cleanup via `CleanInvalid()`
- Never crash on individual file failures
- Use `sync/atomic` for concurrent counters (e.g., success/fail counts in parallel downloads)

### 7.5 CLI Design
- Subcommands: `start` (periodic sync), `sync` (manual sync), `validate` (check cache), `list` (show status)
- Flags: `--config`, `--cache-dir`, `--timeout`, `--verbose`, `--interval` (for start)
- Exit codes: 0 = success, 1 = error
- Signal handling: Support Ctrl+C (SIGINT) and SIGTERM for graceful shutdown
- Use `signal.NotifyContext()` for clean cancellation

### 7.6 Concurrency Safety
- Use `sync/atomic.Int64` for shared counters in goroutines (NOT plain int with ++ operator)
- Use `sync.Mutex` for protecting shared state (e.g., `running` flag in Syncer)
- Re-create channels when restarting components (e.g., `stopChan` in Syncer.Start())
- Use `context.Context` for cancellation propagation
- Prefer `sync.WaitGroup` for managing goroutine lifecycle

---

## 8. Pre-Output Self-Check Checklist

Before outputting any code, evaluate each item:
- Does this code create 2x+ more structure/files than the requirement needs?
- Are there unused interfaces/abstractions/design patterns/middleware?
- Was a new test file created when tests already exist?
- Were >5 new test cases added when the requirement was just a bug fix / field addition?
- Were new README/docs files created without being asked?
- Were dependencies added outside the allowlist?
- Does new code use `log/slog` instead of other logging frameworks?

→ If any is "yes" → immediately simplify to a minimal incremental version and note the reason.

---

## 9. Output Format (Always Follow)

1. First provide the "minimal change" version:
   - Code changes (diff or complete small file)
   - Test additions (only the appended portion)
   - Docstring / comment additions
2. Then confirm whether `doc/` documentation needs updating

---

## 10. Environment Variables & Config Management

- Project uses config files (YAML/JSON) or env vars for configuration
- Never hardcode secrets or URLs in code
- When adding new config options:
  1. Add struct field in `pkg/config/config.go`
  2. Provide sensible defaults
  3. Document in code comments

### Common Config Options
- `CACHE_DIR` — Local cache directory path (default: `./cache`)
- `DOWNLOAD_TIMEOUT` — HTTP timeout in seconds (default: 30)
- `MAX_RETRIES` — Max retry attempts (default: 3)
- `SYNC_INTERVAL` — Sync interval in minutes (default: 60)
- `AUTO_START` — Auto-start periodic sync on startup (default: false)

### Config Priority
Configuration loading follows this priority order (later overrides earlier):
1. **Default values** — Hardcoded defaults in code
2. **Config file** — YAML/JSON configuration file
3. **Environment variables** — System environment variables (highest priority)

This design allows:
- Development with config files for version control
- Production deployment with environment variables
- Easy configuration in Docker/K8s environments

---

## 11. Git Rules

### Forbidden Files
- `*.exe`, `*.dll`, `*.so`, `*.dylib` (binaries)
- `.env` (contains secrets)
- `vendor/` (dependency dir)
- `.idea/`, `.vscode/` (IDE config)
- `cache/` (runtime data)

### Commit Convention
- Use Conventional Commits format
- Check `git status` before committing to ensure no unexpected files
- Do not commit cached `.srs` files or metadata

---

## 12. Agent Orchestration & Workflow (Mandatory)

### 12.1 Role Separation
- **Main Agent** (this session): Orchestrator only. NEVER writes code directly.
  - Creates task plans in `doc/plans/` before any implementation
  - Delegates all code changes to **Sub Agents** via the `task` tool
  - Maintains long-running task loops until the project goal is complete
  - Tracks progress, resolves blockers, and coordinates sub agent outputs
- **Sub Agents**: Executors. Implement code, run tests, perform reviews.
  - `explore` agent: Codebase analysis, file search, architecture questions
  - `code-reviewer` agent: Code review, security checks, best practice verification
  - `general` agent: Code implementation, test writing, file modifications, multi-step tasks
  - `requirement-clarity` agent: Requirement clarification before implementation

### 12.2 Workflow (Every Task MUST Follow This Order)

```
1. PLAN       → Main agent creates/updates plan in doc/plans/<task-name>-plan.md
2. USER REVIEW → Main agent presents plan to user and WAITS for explicit approval before proceeding
3. DELEGATE   → Main agent spawns sub agent(s) to implement
4. REVIEW     → Main agent spawns sub agent to review code against plan
5. UPDATE     → Sub agent updates plan document with task status
6. LOOP       → Main agent checks progress, spawns next task or loops until done
```

**User Review Gate (Mandatory)**:
- After creating or significantly modifying a plan, the main agent MUST present the plan summary to the user using the `question` tool and wait for approval.
- The prompt MUST include: plan file path, task list, estimated scope, and any open questions.
- The user may: approve (proceed to DELEGATE), request changes (return to PLAN), or reject (cancel task).
- **NEVER proceed to DELEGATE without explicit user approval.**
- If the user requests changes, update the plan and re-present for review.

### 12.3 Plan Document Format (`doc/plans/<task-name>-plan.md`)

Every plan document MUST follow these rules:

#### 語言與聲明
- **必須使用繁體中文**撰寫規劃文件（標題、說明、備註等）
- 文件頂端**必須包含 AI 生成聲明**：

```markdown
> **⚠️ AI 生成聲明**：本文件由 AI 輔助生成，僅供參考。所有內容必須經過**人類 Review + 確認**後才能進入實施階段。未經確認前，不得開始任何開發工作。
```

- **人類 Review 是強制步驟**：即使 AI 規劃看起來完整正確，仍必須等待使用者明確確認（✅ Approved）才能進入 DELEGATE 階段
- 規劃文件的「人類 Review 記錄」區塊必須記錄每次 review 的結果

#### 文件結構

Every plan document MUST use this structure:

```markdown
# <任務標題>

> **⚠️ AI 生成聲明**：本文件由 AI 輔助生成，僅供參考。所有內容必須經過**人類 Review + 確認**後才能進入實施階段。未經確認前，不得開始任何開發工作。

**建立日期**: YYYY-MM-DD
**狀態**: 🟡 等待 Review | 🟢 完成 | 🔴 阻塞 | ⚪ 待處理
**優先級**: High | Medium | Low

---

## 背景與目標
簡要說明為什麼需要這個任務，以及預期達成什麼目標。

## 任務列表

| # | 任務 | 狀態 | Sub Agent | 備註 |
|---|------|------|-----------|------|
| 1 | 任務描述 | ⬜ 待處理 | — | |
| 2 | 任務描述 | 🔄 進行中 | general | |
| 3 | 任務描述 | ✅ 完成 | general | 已 Review |

### 狀態圖示說明
- ⬜ 待處理 — 尚未開始
- 🔄 進行中 — sub agent 執行中
- ✅ 完成 — 已實作並通過 review
- 🔴 阻塞 — 有依賴或問題需解決
- ⏭️ 跳過 — 不需要，附原因說明

## 人類 Review 記錄

| 日期 | 動作 | 備註 |
|------|------|------|
| YYYY-MM-DD | ⏳ 等待 Review / ✅ 已批准 / 🔄 要求修改 / ❌ 已拒絕 | |

## Review 日誌

| 日期 | Reviewer | 發現 | 處理方式 |
|------|----------|------|----------|
| YYYY-MM-DD | code-reviewer | 描述 | 修復方式 |

## 備註
任何額外的背景資訊、決策或阻塞問題。
```

### 12.4 Sub Agent Delegation Rules

- **Always include full context** in the prompt: file paths, expected behavior, constraints from AGENTS.md.
- **One sub agent per logical task** — do not batch unrelated changes.
- **Parallel when independent** — launch multiple sub agents simultaneously for unrelated tasks.
- **Sequential when dependent** — wait for prior sub agent to finish before spawning the next.
- Use `explore` agent before `general` agent when the task requires understanding unfamiliar code.
- Use `code-reviewer` agent after every implementation task (step 3 in workflow).

### 12.5 Main Agent Loop Responsibilities

For multi-task or long-running projects, the main agent MUST:
1. Break the project into discrete tasks in the plan document
2. Execute tasks one-by-one or in parallel batches via sub agents
3. After each batch: verify results, update plan status
4. If a sub agent fails or produces incorrect output: diagnose, adjust prompt, re-delegate
5. Continue the loop until ALL tasks in the plan are ✅ Complete or explicitly ⏭️ Skipped
6. After each major milestone, spawn a `code-reviewer` sub agent for a holistic review

### 12.6 Forbidden Main Agent Actions
The main agent MUST NOT:
- Write or edit source code files directly (use sub agents)
- Skip the plan document step
- Skip the user review step (NEVER delegate without explicit user approval of the plan)
- Skip the review step
- Mark tasks as complete without sub agent review confirmation
- Abandon a task loop without updating the plan document status to 🔴 Blocked with a reason

---

## 13. Lessons Learned & Best Practices

### 13.1 Common Pitfalls to Avoid

**❌ Data Races in Concurrent Code**
- Never use plain `int` with `++` operator in goroutines
- Always use `sync/atomic.Int64` for shared counters
- Example from code review: `successCount++` in parallel downloads caused data race

**❌ Channel Reuse After Close**
- Don't reuse closed channels when restarting components
- Re-create channels in `Start()` methods (e.g., `stopChan = make(chan struct{})`)
- Closed channels will immediately return zero value, causing silent failures

**❌ Reinventing Standard Library**
- Don't create custom `byteReader` — use `bytes.NewReader()`
- Don't manually implement substring search — use `strings.Contains()`
- Don't write custom error type checks — use `errors.As()` with `net.Error`

**❌ Unused Code**
- Regularly check for unused functions with `go vet`
- Remove dead code promptly to avoid confusion

### 13.2 Testing Best Practices

**E2E Testing**
- Use real URLs for end-to-end validation
- Test conditional requests (304 Not Modified) by downloading twice
- Verify metadata file generation and content
- Test cleanup of invalid cache files
- Use `t.TempDir()` for automatic cleanup

**Unit Testing**
- Test edge cases: empty files, wrong magic bytes, unsupported versions
- Use table-driven tests for multiple scenarios
- Mock network calls when possible, but keep some real HTTP tests
- Test concurrent behavior with `-race` flag

### 13.3 Documentation Standards

**CLI Documentation**
- Include real-world examples with actual URLs
- Document all flags and their effects
- Provide environment variable equivalents
- Show expected output format

**Code Comments**
- Godoc for exported functions (max 6 lines)
- Explain non-obvious design decisions
- Reference AGENTS.md sections when relevant
- Keep comments up-to-date with code changes

### 13.4 Performance Considerations

**Concurrent Downloads**
- No hard limit on concurrency (scales with rule count)
- Monitor system resources for large rule sets
- Consider adding `max_concurrent_downloads` config option if needed

**Memory Usage**
- SRS files are validated in streaming fashion (no full load into memory)
- Metadata files are small JSON (~500 bytes each)
- Temp files use same disk space as final files

**Network Efficiency**
- Conditional requests save bandwidth on unchanged files
- Exponential backoff prevents overwhelming servers
- Timeout settings prevent hanging connections

### 13.5 Security Considerations

**File Operations**
- Atomic replacement prevents partial file exposure
- Temp files in same directory ensure atomic rename works
- No path traversal vulnerabilities (use `filepath.Join()`)

**HTTP Requests**
- Always set timeouts (prevent DoS)
- Validate downloaded content before use
- No credential storage in code or config

**Input Validation**
- Validate config values (positive timeouts, non-negative retries)
- Check URL format before downloading
- Verify SRS file structure before caching

---

## Key Context

- **sing-box**: A universal proxy platform; rule-sets define routing rules in binary `.srs` format
- **Remote rule-sets**: Typically hosted at URLs like `https://example.com/rules.srs` or GitHub/jsDelivr CDNs
- **Goal**: Download and cache these remote `.srs` files locally for offline/reusable access
- **SRS format**: Binary format used by sing-box for efficient rule matching
  - Magic bytes: `0x53, 0x52, 0x53` ("SRS")
  - Version: uint8 (currently 1-5)
  - Payload: zlib-compressed rule list
- **Real-world usage**: This tool has been tested with 8 real SRS URLs from MetaCubeX and other providers
- **Production ready**: All features implemented, 24 tests passing, code review completed
