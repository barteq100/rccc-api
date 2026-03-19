# rccc-api

Core backend API for the Remote Career Command Center MVP.

## Purpose

`rccc-api` is the system of record for:

- canonical jobs
- user profile and preferences
- saved jobs
- application tracking
- deterministic fit scoring
- REST contracts consumed by the web app and ingestion worker

## Tech Choice

- Go for a small, explicit HTTP service
- PostgreSQL for canonical persistence
- Redis is optional and deferred until runtime pressure justifies it

## Initial Structure

- `cmd/api`: process entrypoint
- `internal/jobs`: job domain logic
- `internal/applications`: saved job and application state
- `internal/profile`: user preferences
- `internal/scoring`: deterministic fit scoring
- `internal/platform/http`: transport concerns
- `internal/platform/db`: persistence bootstrap
- `migrations`: SQL schema changes

## Local Development

1. Copy `.env.example` to `.env`.
2. Run `docker compose up --build`.
3. For local Go execution, run `go run ./cmd/api`.

The current scaffold exposes a health endpoint placeholder only.

## Assumptions

- MVP supports a single-user workflow first.
- Ingestion writes through the API, not direct database writes.
- Auth is deferred for MVP scaffolding.

## MVP Backlog

- define canonical job schema and migrations
- implement ingestion upsert endpoint
- implement jobs browse and detail endpoints
- implement save/apply/status workflow
- implement profile/preferences endpoints
- add deterministic fit scoring service
- add repository and HTTP handler tests
