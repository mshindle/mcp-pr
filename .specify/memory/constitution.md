<!--
SYNC IMPACT REPORT
==================
Version change: unversioned → 1.0.0 (MAJOR: initial ratification)

Modified principles: N/A (all new)

Added sections:
  - Core Principles (I–V)
  - Technology Constraints
  - Development Workflow
  - Governance

Removed sections: N/A

Templates:
  ✅ .specify/templates/plan-template.md — Constitution Check section present; gates now
     resolvable against concrete principles. No structural change needed.
  ✅ .specify/templates/spec-template.md — User-story and requirements structure aligns
     with all principles. No change needed.
  ⚠ .specify/templates/tasks-template.md — "Tests are OPTIONAL" comment conflicts with
     Principle III (Test-First NON-NEGOTIABLE). Updated to reflect mandatory tests.
  ✅ .specify/templates/checklist-template.md — Not read; no known conflict. Verify on
     next checklist generation.

Follow-up TODOs: None deferred.
-->

# mcp-pr Constitution

## Core Principles

### I. MCP-First Architecture

Every feature MUST be exposed exclusively as an MCP tool via `modelcontextprotocol/go-sdk`.
The MCP protocol is the sole external interface for this project. Adding HTTP REST, gRPC, or
other transport layers is prohibited unless a dependency mandates it, and any such exception
MUST be documented in the feature's plan.md Complexity Tracking table.

**Rationale**: Consistent MCP exposure keeps the tool composable with any MCP-compatible host
and prevents interface proliferation.

### II. Multi-Provider AI

The system MUST treat AI providers (Anthropic, OpenAI, Google GenAI) as interchangeable
behind a shared Go interface. No feature may be hard-coded to a single provider. Provider
selection MUST be runtime-configurable (e.g., via environment variable or flag). Feature
implementations that call an AI provider MUST program to the interface, not the concrete SDK.

**Rationale**: The go.mod includes three distinct AI SDKs; coupling to one defeats the
multi-provider intent and makes future provider additions expensive.

### III. Test-First (NON-NEGOTIABLE)

Tests MUST be written and confirmed failing before implementation begins. The Red-Green-Refactor
cycle is strictly enforced:

1. Write test → confirm it fails (`go test` reports failure)
2. Write minimum implementation to pass
3. Refactor under green tests

Integration tests MUST exercise real provider interfaces; mocking AI provider behavior is
prohibited. Unit tests may mock non-AI dependencies (filesystem, network, etc.).

**Rationale**: MCP tool correctness is critical; mocked AI behavior has historically masked
real integration failures. Real provider tests are required to catch protocol drift.

### IV. Simplicity

YAGNI is enforced. Abstractions MUST NOT be introduced until 3+ concrete, existing usages
justify them. Every non-trivial design choice (additional layers, patterns, packages) MUST be
justified in plan.md's Complexity Tracking table with the simpler alternative explicitly
rejected. Prefer flat package structures; avoid deep nesting.

**Rationale**: Early-stage projects accumulate accidental complexity quickly. Explicit
justification creates a forcing function against premature generalization.

### V. Observability

All MCP tool invocations MUST emit structured log entries (at minimum: tool name, input
summary, outcome, duration). Errors MUST propagate to stderr; successful tool results to
stdout. AI provider interactions MUST be logged at DEBUG level with request/response
summaries (not full payloads by default).

**Rationale**: MCP tools are called by external hosts; silent failures are extremely difficult
to diagnose without structured traces.

## Technology Constraints

- **Language / runtime**: Go 1.26, module `github.com/mshindle/mcp-pr`
- **Approved dependencies** (additions require plan.md justification):
  - `anthropics/anthropic-sdk-go`, `openai/openai-go/v3`, `google.golang.org/genai` — AI providers
  - `modelcontextprotocol/go-sdk` — MCP protocol
  - `gorilla/websocket` — transport (WebSocket)
  - `tidwall/gjson`, `tidwall/sjson` — JSON manipulation
- Adding a dependency not in the approved list MUST be justified in the relevant plan.md
  before the `go.mod` entry is committed.
- All new code MUST pass `go vet ./...` and `go fmt ./...` without warnings before merge.

## Development Workflow

- Feature branches follow sequential numbering: `###-feature-name` (created via
  `.specify/scripts/powershell/create-new-feature.ps1`).
- All features MUST traverse the full Specify pipeline:
  `specify → (clarify) → plan → tasks → implement → analyze`
- Every PR MUST reference a `specs/###-feature-name/spec.md`.
- The Constitution Check in plan.md MUST be completed and passing before Phase 0 research
  begins and re-verified after Phase 1 design.
- Each user story MUST be independently testable and deployable as a standalone MVP increment.

## Governance

This constitution supersedes all other project conventions and practices. When a conflict
arises between this document and any other guideline, the constitution prevails.

**Amendment procedure**:

1. Propose the change with rationale (can be inline in a PR description or a spec).
2. Increment `CONSTITUTION_VERSION` per semantic versioning rules (see version policy below).
3. Update `LAST_AMENDED_DATE` to the amendment date.
4. Run `/speckit.constitution` to propagate changes to dependent templates.
5. Reference the new version in the PR that enacts the amendment.

**Version policy**:
- MAJOR: Principle removed, redefined, or governance incompatibly changed.
- MINOR: New principle or section added, or materially expanded guidance.
- PATCH: Clarification, wording, or non-semantic refinement.

**Compliance**: All PRs must be reviewed against this constitution before merge. Complexity
violations not documented in plan.md's Complexity Tracking table are grounds for PR rejection.

---

**Version**: 1.0.0 | **Ratified**: 2026-03-24 | **Last Amended**: 2026-03-24
