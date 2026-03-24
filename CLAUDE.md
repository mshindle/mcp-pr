# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`mcp-pr` is a Go-based MCP (Model Context Protocol) tool at `github.com/mshindle/mcp-pr` (Go 1.26). It uses the Specify framework for specification-driven, AI-assisted feature development.

## Commands

```bash
# Build
go build ./...

# Test
go test ./...
go test ./path/to/package -run TestName   # single test

# Lint / Format
go vet ./...
go fmt ./...
```

## Key Dependencies

- `anthropics/anthropic-sdk-go` — Claude AI integration
- `modelcontextprotocol/go-sdk` — MCP protocol support
- `openai/openai-go/v3` — OpenAI SDK
- `google.golang.org/genai` — Google Generative AI
- `gorilla/websocket` — WebSocket support
- `tidwall/gjson`, `tidwall/sjson` — JSON manipulation

## Specify Workflow

Features are developed through a structured pipeline using slash commands:

1. `/speckit.specify` — Creates `specs/[###-feature-name]/spec.md` from natural language
2. `/speckit.clarify` — Asks up to 5 clarifying questions and encodes answers into the spec
3. `/speckit.plan` — Generates `plan.md`, `data-model.md`, `research.md` under the feature directory
4. `/speckit.tasks` — Generates `tasks.md` with dependency-ordered, parallel-safe tasks
5. `/speckit.implement` — Executes tasks from `tasks.md` sequentially/in parallel
6. `/speckit.analyze` — Cross-checks consistency across `spec.md`, `plan.md`, `tasks.md`

Tasks in `tasks.md` are annotated with `[P]` (parallel-safe), story number `[###]`, and file path `[Path]`. Each user story should be independently testable.

## Branch Naming

Feature branches follow sequential numbering: `001-feature-name`, `002-feature-name`, etc. Use `.specify/scripts/powershell/create-new-feature.ps1` to create branches (outputs JSON with branch metadata).

## Project Constitution

`.specify/memory/constitution.md` governs architectural constraints and development principles. Run `/speckit.constitution` to initialize it before starting feature work. Once ratified, it takes precedence over ad-hoc decisions.
