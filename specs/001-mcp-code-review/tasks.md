# Tasks: MCP Code Review Server

**Input**: Design documents from `/specs/001-mcp-code-review/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/mcp-tools.md ✅, quickstart.md ✅

**Tests**: Tests are REQUIRED per Constitution Principle III (Test-First, NON-NEGOTIABLE).
Tests MUST be written and confirmed failing before implementation begins.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create project directory structure and verify all declared dependencies are present in go.mod.

- [x] T001 Create directory structure: `cmd/mcp-code-review/`, `internal/provider/`, `internal/git/`, `internal/review/`, `internal/server/`, `tests/integration/`, `tests/unit/`
- [x] T002 [P] Verify go.mod contains all required dependencies: `modelcontextprotocol/go-sdk`, `anthropics/anthropic-sdk-go`, `openai/openai-go/v3`, `google.golang.org/genai`; run `go mod tidy` if needed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types, interfaces, and shared infrastructure that ALL user stories depend on. Must be complete before any user story implementation begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Types & Interfaces (define contracts first)

- [x] T003 [P] Define `ReviewInput`, `ReviewResult`, and `Finding` structs with json tags and validation rules in `internal/review/types.go`
- [x] T004 [P] Define `Provider` interface (`Review`, `Name`, `DefaultModel`) and `ConstructorFunc` type in `internal/provider/provider.go`

### Tests (write before implementation — must FAIL first)

- [x] T005 [P] Write unit tests covering staged/unstaged/commit diff extraction, empty diff, binary file skipping, git-not-found error, and bad repo path in `tests/unit/git_test.go`
- [x] T006 [P] Write unit tests covering prompt construction with/without language hint, ReviewResult JSON parsing, invalid severity normalization, and empty-message filtering in `tests/unit/review_test.go`
- [x] T007 [P] Write unit tests covering registry Register/Resolve, default provider resolution order (Anthropic→OpenAI→Google), missing-key error, and unsupported-provider error in `tests/unit/registry_test.go`

### Implementation (after tests are confirmed failing)

- [x] T008 Implement `DiffResult` struct and `StagedDiff`, `UnstagedDiff`, `CommitDiff` functions with binary-file detection in `internal/git/git.go`
- [x] T009 Implement system-instruction prompt and user-message formatter with optional language/context fields; include JSON fallback for unparseable model output in `internal/review/prompt.go`
- [x] T010 Implement `Registry` struct with `Register`, `Resolve`, and `DefaultProvider` methods; implement env-var-based default resolution (`ANTHROPIC_API_KEY` → `OPENAI_API_KEY` → `GOOGLE_API_KEY`) in `internal/provider/registry.go`

**Checkpoint**: Run `go test ./tests/unit/...` — all unit tests must pass before any user story work begins.

---

## Phase 3: User Story 1 — Review Git Staged Changes (Priority: P1) 🎯 MVP

**Goal**: A developer with staged changes can invoke `review_staged` and receive a `ReviewResult` JSON from an AI provider.

**Independent Test**: Stage a file change in any git repository, invoke the tool, and verify a structured review is returned without committing. Verify "no staged changes" message when nothing is staged.

### Tests for User Story 1 (REQUIRED — write first, must FAIL)

- [x] T011 [US1] Write integration test (`//go:build integration`) covering happy path (staged changes → ReviewResult), nothing-staged error, and 500+ line diff in `tests/integration/staged_test.go`

### Implementation for User Story 1

- [x] T012 [US1] Implement `AnthropicProvider` with `Review()` method calling Anthropic SDK, JSON response parsing, and model default `claude-sonnet-4-6` in `internal/provider/anthropic.go`
- [x] T013 [US1] Implement `ReviewStagedInput` struct and `handleReviewStaged` handler: call `git.StagedDiff`, validate non-empty, resolve provider, call `provider.Review`, return JSON text result; include slog logging for invocation and provider call in `internal/server/tools.go`
- [x] T014 [US1] Implement `NewServer` function: initialize slog handler from `LOG_LEVEL` env var, build provider registry, create MCP server, register `review_staged` tool with its handler in `internal/server/server.go`
- [x] T015 [US1] Implement `main()`: call `server.NewServer`, call `server.Run(ctx, mcp.NewStdioTransport())`, handle OS signals in `cmd/mcp-code-review/main.go`

**Checkpoint**: Build and run `go build ./cmd/mcp-code-review`. Pipe the MCP initialize message from quickstart.md and verify a valid JSON response. Run `go test ./tests/integration/... -tags integration` against a repo with staged changes.

---

## Phase 4: User Story 2 — Select AI Provider (Priority: P2)

**Goal**: Users can pass `provider` (and optionally `model`) to any review tool; the correct provider API is called. Unspecified provider falls back to first available API key. Unsupported provider returns a clear error.

**Independent Test**: Invoke any review tool with each of `provider: "anthropic"`, `provider: "openai"`, `provider: "google"` (with `LOG_LEVEL=DEBUG`) and verify the correct provider is called. Invoke with `provider: "groq"` and verify a clear error listing supported providers.

### Tests for User Story 2 (REQUIRED — write first, must FAIL)

- [x] T016 [P] [US2] Write integration test (`//go:build integration`) covering provider switching across all three providers and unsupported-provider error in `tests/integration/arbitrary_test.go`

### Implementation for User Story 2

- [x] T017 [P] [US2] Implement `OpenAIProvider` with `Review()` method calling OpenAI SDK, JSON response parsing, and model default `gpt-4o`; support `MCP_REVIEW_OPENAI_MODEL` env override in `internal/provider/openai.go`
- [x] T018 [P] [US2] Implement `GoogleProvider` with `Review()` method calling Google Generative AI SDK, JSON response parsing, and model default `gemini-2.0-flash`; support `MCP_REVIEW_GOOGLE_MODEL` env override in `internal/provider/google.go`
- [x] T019 [US2] Register all three providers in `NewServer` (update `internal/server/server.go`); wire `MCP_REVIEW_ANTHROPIC_MODEL` env override for AnthropicProvider
- [x] T020 [US2] Add provider validation in tool handlers: resolve provider via Registry, return structured error listing supported providers on unknown name; update `internal/server/tools.go`

**Checkpoint**: Run integration tests with each provider key set. Confirm `LOG_LEVEL=DEBUG` logs show correct provider name. Confirm error message for `provider: "groq"`.

---

## Phase 5: User Story 3 — Review Git Unstaged Changes (Priority: P3)

**Goal**: A developer with modified but unstaged files can invoke `review_unstaged` and receive a `ReviewResult`. Returns a clear message when nothing is unstaged.

**Independent Test**: Modify a tracked file without staging; invoke `review_unstaged`; verify the review covers only working-directory changes. Confirm "no unstaged changes" message on a clean working tree.

### Tests for User Story 3 (REQUIRED — write first, must FAIL)

- [x] T021 [US3] Write integration test (`//go:build integration`) covering modified-unstaged happy path and no-unstaged-changes error in `tests/integration/unstaged_test.go`

### Implementation for User Story 3

- [x] T022 [US3] Implement `ReviewUnstagedInput` struct and `handleReviewUnstaged` handler in `internal/server/tools.go`
- [x] T023 [US3] Register `review_unstaged` tool in `NewServer` in `internal/server/server.go`

**Checkpoint**: Modify a tracked file, invoke `review_unstaged`, verify ReviewResult. Verify error on clean working tree.

---

## Phase 6: User Story 4 — Review a Specific Commit (Priority: P3)

**Goal**: A developer can pass any full (40-char) or abbreviated (≥7-char) commit SHA to `review_commit` and receive a `ReviewResult` for that commit's changes. Invalid or non-existent SHAs return a clear error.

**Independent Test**: Pass a valid commit SHA from repo history; verify the review reflects only that commit's diff. Pass a non-existent SHA; verify clear error.

### Tests for User Story 4 (REQUIRED — write first, must FAIL)

- [x] T024 [US4] Write integration test (`//go:build integration`) covering valid full SHA, valid abbreviated SHA, non-existent SHA error, and missing-sha error in `tests/integration/commit_test.go`

### Implementation for User Story 4

- [x] T025 [US4] Implement `ReviewCommitInput` struct and `handleReviewCommit` handler with SHA presence validation and abbreviated-SHA ambiguity error in `internal/server/tools.go`
- [x] T026 [US4] Register `review_commit` tool in `NewServer` in `internal/server/server.go`

**Checkpoint**: Invoke `review_commit` with `sha: "a5b6cbf"` (initial commit) per quickstart.md scenario 3. Verify ReviewResult is returned.

---

## Phase 7: User Story 5 — Review Arbitrary Code (Priority: P4)

**Goal**: A user can pass a code string directly to `review_code` (no git context required) and receive a `ReviewResult`. Empty `code` parameter returns a clear error.

**Independent Test**: Pass any code string with optional language hint; verify ReviewResult is returned with no git repository present. Pass empty string; verify clear error.

### Tests for User Story 5 (REQUIRED — write first, must FAIL)

- [x] T027 [P] [US5] Write integration test (`//go:build integration`) covering code snippet with language hint, empty-code error, and provider fallback in `tests/integration/arbitrary_test.go` (extend file from T016 if it already exists)

### Implementation for User Story 5

- [x] T028 [US5] Implement `ReviewCodeInput` struct and `handleReviewCode` handler: validate `code` non-empty, build ReviewInput directly (no git call), call provider.Review in `internal/server/tools.go`
- [x] T029 [US5] Register `review_code` tool in `NewServer` in `internal/server/server.go`

**Checkpoint**: Invoke `review_code` with `{"code": "func add(a, b int) int { return a - b }", "language": "go"}` per quickstart.md scenario 4. Verify finding notes the subtraction bug.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge-case hardening, format validation, and final acceptance.

- [x] T030 [P] Verify binary file skipping: ensure `DiffResult.IsBinary` prepends a note to the review summary in tool handlers; add binary-file test case in `tests/unit/git_test.go`
- [x] T031 [P] Add context-window-exceeded error: return `"diff exceeds model context limit; reduce staged changes or split the commit"` in all git-diff tool handlers when the diff length exceeds a configurable threshold; update `internal/server/tools.go`
- [x] T032 [P] Run `go vet ./...` and `go fmt ./...` — both must complete with no output; fix any issues
- [ ] T033 Run all quickstart.md validation scenarios (sections 1–6) end-to-end; confirm all pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) — BLOCKS all user stories
- **User Stories (Phase 3–7)**: All depend on Foundational phase completion
  - Stories can proceed in priority order (P1 → P2 → P3 → P4)
  - US3 and US4 are both P3 and are independent of each other
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — no story dependencies
- **US2 (P2)**: Can start after Phase 2 — extends US1 provider infrastructure; tools from US1 serve as test vehicle
- **US3 (P3)**: Can start after Phase 2 — independent of US1/US2 beyond shared git.go/provider infra
- **US4 (P3)**: Can start after Phase 2 — independent of US3; shares git.go infra
- **US5 (P4)**: Can start after Phase 2 — does not call git.go at all; simplest story

### Within Each User Story

1. Write integration test → confirm it FAILS
2. Implement provider (if new) in `internal/provider/`
3. Implement tool handler struct and function in `internal/server/tools.go`
4. Register tool in `internal/server/server.go`
5. Run integration test → confirm it PASSES

### Parallel Opportunities

Within Phase 2:
- T003 and T004 can run in parallel (different files, no dependencies on each other)
- T005, T006, T007 can all run in parallel (different test files)
- T008 depends on T003; T009 depends on T003; T010 depends on T004

Within Phase 4:
- T017 and T018 can run in parallel (different provider files)

Within Phase 7:
- T027 can run in parallel with other Phase 7 tasks if T016 file is already created

---

## Parallel Execution Examples

### Phase 2 Parallel Batch 1 (types + interface)

```
Task: T003 — internal/review/types.go
Task: T004 — internal/provider/provider.go
```

### Phase 2 Parallel Batch 2 (unit tests — write all at once)

```
Task: T005 — tests/unit/git_test.go
Task: T006 — tests/unit/review_test.go
Task: T007 — tests/unit/registry_test.go
```

### Phase 4 Parallel Batch (provider implementations)

```
Task: T017 — internal/provider/openai.go
Task: T018 — internal/provider/google.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (staged review via Anthropic)
4. **STOP and VALIDATE**: Build binary, run quickstart.md scenario 1, run integration test
5. Demo / validate with real developer workflow before proceeding

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready, all unit tests green
2. Phase 3 (US1) → `review_staged` works → **MVP demo**
3. Phase 4 (US2) → All three providers selectable on any tool
4. Phase 5 (US3) → `review_unstaged` works
5. Phase 6 (US4) → `review_commit` works
6. Phase 7 (US5) → `review_code` works → **All four tools complete**
7. Phase 8 → Polish, `go vet`, quickstart validation

---

## Notes

- `[P]` tasks target different files with no blocking dependencies — safe to parallelize
- `[Story]` labels map each task to a specific user story for traceability
- Integration tests require at least one provider API key; unit tests require none
- Run `go test ./tests/unit/...` after Phase 2 to confirm baseline before user story work
- Run `go test ./tests/integration/... -tags integration` after each user story phase
- Log to stderr only (`log/slog`); stdout is reserved for MCP protocol frames
- All error messages must match the exact strings defined in `contracts/mcp-tools.md`
