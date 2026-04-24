# AGENTS.md — rccc-api

This repository owns the canonical backend API for the Remote Career Command Center MVP.

## Purpose

Use `rccc-api` for:

- canonical job persistence
- ingestion-facing upsert validation and storage
- user-facing job browse and detail APIs
- saved jobs and application tracking
- profile and preference persistence
- deterministic fit scoring

Do not put source-specific parsing, scraping, or frontend UI logic in this repository.

## Technologies

- Go 1.24
- PostgreSQL as the system of record
- Redis is optional and should remain deferred unless runtime needs justify it
- REST over HTTP
- Docker for local execution

## Key Directories

- `cmd/api`: service entrypoint
- `internal/jobs`: job domain logic and query behavior
- `internal/applications`: saved jobs and application workflow
- `internal/profile`: user preference model and handlers
- `internal/scoring`: deterministic fit scoring logic
- `internal/platform/http`: transport and handler wiring
- `internal/platform/db`: database bootstrap and persistence wiring
- `migrations`: schema migrations

## Tools And Commands

Use the repo's existing commands first:

- `go run ./cmd/api`
- `go test ./...`
- `gofmt -w ./cmd ./internal`
- `docker compose up --build`
- `make run`
- `make test`
- `make fmt`

Primary environment variables:

- `PORT`
- `DATABASE_URL`
- `REDIS_URL`

## Conventions

- `rccc-api` is the system of record for canonical jobs, profile data, saved jobs, application states, and score output.
- Ingestion must write through explicit REST endpoints, not through direct database access.
- Keep request and response models stable and explicit. If a contract changes, update the relevant documentation and dependent repos.
- Use exactly these application statuses for MVP: `saved`, `applied`, `interview`, `offer`, `rejected`.
- Keep scoring deterministic and explainable. Return score reasons; do not add AI-based scoring.
- Keep filtering and pagination semantics server-side when they affect canonical browse results.
- Prefer simple handlers and domain services over framework-heavy abstractions.

## Patterns

- Keep transport, domain, and persistence concerns separate.
- Put canonical validation close to the API boundary, then pass validated models into domain services.
- Use repository-style persistence boundaries where they simplify tests and handler wiring.
- Keep migrations additive and explicit.
- Favor table-driven tests for handler and service behavior.

## Required Skills And Owning Agent

Primary skill:

- `rccc-api-backend`

Supporting skills:

- `rccc-contracts`
- `rccc-handoffs-json`
- `rccc-testing`
- `rccc-git-pr`
- optional `security-best-practices`

Expected owning agent:

- `backend-agent`

Invoke supporting skills when:

- a shared payload or endpoint contract changes: `rccc-contracts`
- another agent depends on your output: `rccc-handoffs-json`
- the task is primarily validation or test expansion: `rccc-testing`
- the task is complete and needs closeout: `rccc-git-pr`
## Coordination Rules

- Before starting work, claim or update the relevant task in the root [tasks.md](../tasks.md).
- If another agent will need your output, create or update `../handoffs/TASK-###-slug/` and record the task brief, handoff, decisions, and related files as JSON.
- Use the shared schemas in `../handoffs/schemas/` and keep all handoff `related_files` workspace-relative.
- For delegated multi-step work, update `tasks.md` and `handoff.json` after context is read, after meaningful edits, after validation, and before exit.`r`n- For live observability, create or update `../handoffs/TASK-###-slug/worker-<agent>.json` with heartbeat, current state, current step, touched files, and validation status.
- End every delegated run with an explicit `done`, `blocked`, or `partial` outcome in the returned message. If the outcome is `partial`, leave the task `in_progress` with a precise next step. If the outcome is `blocked`, set the task to `blocked` and record the blocker in `handoff.json`.
- When a repository task is complete, finish it with a git commit, push the branch, and open or update a pull request unless the task is explicitly documentation-only workspace coordination outside that repository.
- If work crosses into ingestion contracts or frontend payloads, update the shared contract task or create a new coordination task first.
- Do not expand scope into auth, AI features, or deployment architecture unless a task explicitly requires it.







