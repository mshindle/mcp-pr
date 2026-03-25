# Feature Specification: MCP Code Review Server

**Feature Branch**: `001-mcp-code-review`
**Created**: 2026-03-24
**Status**: Draft
**Input**: User description: "Write an MCP server for doing code reviews. Options for reviewing arbitrary code, git staged, unstaged, or a specific commit. Use the go-sdk/mcp library for the MCP. Use the OpenAI, Google, Anthropic SDKs for connecting to models for the review."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Review Git Staged Changes (Priority: P1)

A developer has staged changes with `git add` and wants an AI-powered code review before
committing. They invoke the review tool requesting staged diff, receive a structured review
with findings, and decide whether to proceed with the commit.

**Why this priority**: Staged review is the most natural integration point in a developer's
workflow — it runs right before a commit, providing the highest-value feedback at the lowest
interruption cost.

**Independent Test**: Can be fully tested by staging a file change in any git repository,
invoking the tool, and verifying a structured review is returned without committing.

**Acceptance Scenarios**:

1. **Given** a git repository with staged changes, **When** the user invokes the staged-review
   tool, **Then** a structured code review is returned covering all staged files.
2. **Given** a git repository with no staged changes, **When** the user invokes the
   staged-review tool, **Then** the tool returns a clear message indicating nothing is staged.
3. **Given** staged changes exceeding 500 changed lines, **When** the review runs, **Then**
   the tool returns a review without error (may summarize at a higher level of abstraction).

---

### User Story 2 - Select AI Provider (Priority: P2)

A user wants to choose which AI provider (Anthropic, OpenAI, or Google) performs the review,
and optionally specify a model. Different users have different API keys, cost preferences,
or quality expectations.

**Why this priority**: Provider selection is a prerequisite for any review to succeed — each
user brings their own API credentials and may have access to only one provider.

**Independent Test**: Can be tested by invoking any review tool with a provider parameter
set to each supported provider and verifying the correct provider's API is called.

**Acceptance Scenarios**:

1. **Given** a provider parameter set to "anthropic", "openai", or "google", **When** any
   review tool is invoked, **Then** the review is performed using that provider's API.
2. **Given** no provider parameter is specified, **When** a review tool is invoked, **Then**
   the tool selects the default provider based on available environment API keys.
3. **Given** an unsupported provider name is specified, **When** the tool is invoked,
   **Then** a clear error listing supported providers is returned.

---

### User Story 3 - Review Git Unstaged Changes (Priority: P3)

A developer is mid-edit and wants an AI review of their working-directory changes before
staging them. They invoke the review tool requesting unstaged changes and receive actionable
feedback.

**Why this priority**: Unstaged review enables earlier feedback in the development loop but
is less critical than the pre-commit staged review.

**Independent Test**: Can be tested by modifying a tracked file without staging, invoking the
tool, and verifying the review covers only the working-directory diff.

**Acceptance Scenarios**:

1. **Given** a git repository with modified but unstaged files, **When** the user invokes the
   unstaged-review tool, **Then** a review covering all unstaged modifications is returned.
2. **Given** a repository with no unstaged modifications, **When** the user invokes the
   unstaged-review tool, **Then** the tool returns a clear message indicating no unstaged
   changes exist.

---

### User Story 4 - Review a Specific Commit (Priority: P3)

A developer or reviewer wants to understand and critique a past commit. They provide a commit
identifier and receive a review of exactly the changes introduced by that commit.

**Why this priority**: Retrospective commit review is useful but less time-sensitive than
in-flight pre-commit reviews.

**Independent Test**: Can be tested by passing any valid commit SHA from a repository's
history and verifying the review reflects only that commit's diff.

**Acceptance Scenarios**:

1. **Given** a valid commit SHA (full or abbreviated), **When** the user invokes the
   commit-review tool with that SHA, **Then** a review of the changes in that commit is
   returned.
2. **Given** an invalid or non-existent commit SHA, **When** the user invokes the tool,
   **Then** a clear error message is returned indicating the commit was not found.

---

### User Story 5 - Review Arbitrary Code (Priority: P4)

A user wants a code review for a text snippet they provide directly, without any git context.
They pass code as text input and receive a review.

**Why this priority**: Useful for reviewing code outside of a git repository or for ad-hoc
snippets, but less integrated into a standard developer workflow.

**Independent Test**: Can be tested by passing any text string containing code and verifying
a review is returned without any git repository present.

**Acceptance Scenarios**:

1. **Given** a string of source code provided as input, **When** the user invokes the
   arbitrary-code-review tool, **Then** a review of that code is returned.
2. **Given** an empty string as code input, **When** the user invokes the tool, **Then**
   the tool returns a clear message that no code was provided.

---

### Edge Cases

- What happens when the git binary is not found on the host system?
- How does the system handle binary files in a diff (images, compiled assets)?
- What if an AI provider API key is missing or expired?
- What if the repository is in a detached HEAD state when requesting staged/unstaged changes?
- What if a diff is too large for the selected model's context window?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose each review mode (staged, unstaged, commit, arbitrary)
  as a distinct MCP tool.
- **FR-002**: The system MUST support Anthropic, OpenAI, and Google Generative AI as
  interchangeable review providers selectable at tool-call time.
- **FR-003**: Users MUST be able to specify the AI provider and model via tool input
  parameters; unspecified provider MUST fall back to a default derived from available
  environment API keys.
- **FR-004**: The system MUST return structured review output including: overall summary,
  per-file findings, and severity indicators (issue / suggestion / praise).
- **FR-005**: The system MUST return a clear, human-readable error when a git operation fails
  (no repository found, bad SHA, nothing staged, etc.).
- **FR-006**: The system MUST return a clear error when an AI provider call fails (missing
  key, rate limit, network timeout).
- **FR-007**: The arbitrary-code-review tool MUST accept source code as a direct string
  parameter, with an optional language hint parameter.
- **FR-008**: The commit-review tool MUST accept both full (40-char) and abbreviated (≥7-char)
  commit SHAs.
- **FR-009**: Binary files encountered in a diff MUST be skipped with a note in the review
  output rather than causing an error.
- **FR-010**: The system MUST log all tool invocations and AI provider interactions at
  structured, configurable verbosity levels.

### Key Entities

- **ReviewRequest**: Encapsulates the source of code to review (type: staged | unstaged |
  commit | arbitrary), provider selection, model selection, and optional language hint.
- **ReviewResult**: Structured output containing an overall summary and a list of per-file
  findings, each with severity, optional location (file + line range), and message.
- **Provider**: An AI backend (Anthropic / OpenAI / Google) identified by name, with an
  associated model identifier and runtime client.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can complete a full staged-diff review in under 30 seconds for
  a typical change set (under 200 changed lines) on a standard internet connection.
- **SC-002**: All four review modes work correctly against all three providers in end-to-end
  validation.
- **SC-003**: Switching providers requires only a parameter change at call time — no
  configuration file edits or server restarts.
- **SC-004**: A clear, actionable error message is returned in 100% of invalid-input and
  provider-failure scenarios with no silent failures or crashes.
- **SC-005**: The MCP server starts and successfully registers all tools within 2 seconds
  of launch on a standard developer workstation.

## Assumptions

- The git binary is available in the PATH on the host system where the MCP server runs;
  diff extraction is performed via subprocess calls to git.
- API keys for the chosen AI provider are supplied via environment variables at runtime;
  the server does not store or manage credentials.
- Review quality (accuracy, depth) is entirely delegated to the underlying AI model; the
  server is responsible only for correct input construction and output parsing.
- A single review request targets one provider and one model; fan-out or consensus across
  multiple providers is out of scope for v1.
- The MCP server runs as a local process (stdio or SSE transport); multi-tenant or
  cloud-hosted server deployment is out of scope.
- Diffs larger than a model's context window are out of scope for v1 and MUST return a
  clear error rather than silently truncating input.
