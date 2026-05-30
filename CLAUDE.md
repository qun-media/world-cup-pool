# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development
make dev-backend     # Run PocketBase backend on :8091 (disposable pb_data_dev)
make dev-frontend    # Run SvelteKit dev server on :5173 (proxies /api to :8091)

# Build
make build           # Build single Go binary with embedded frontend
make build-frontend  # Build SvelteKit SPA into internal/web/build (required before make build)

# Test
make test            # Run Go tests
npm run test         # Vitest unit tests (run from frontend/)
npm run test:e2e     # Playwright e2e tests (run from frontend/)
npm run check        # svelte-check + TypeScript type check (run from frontend/)

# Docker
make docker          # Build production Docker image
make docker-dev      # Run isolated test container on :8091
make stop-test       # Stop test container

# Reset
make reset           # Wipe pb_data_dev (dev database)
make clean           # Remove build artifacts
```

## Architecture

**Single-binary self-hosted app**: Go backend (PocketBase) with an embedded SvelteKit SPA. SQLite is the only database. Deployed as a single Docker container.

### Backend (`internal/` packages)

Each subdirectory is a feature module that registers its own PocketBase hooks and HTTP routes in `main.go`:

- **seed** — Seeds teams, groups, and fixture data from openfootball JSON on first boot (idempotent)
- **scoring** — Core engine: recomputes all match/forecast scores for all users/configs after any result change. All scoring is a full idempotent recompute — never partial.
- **leagues** — Private league lifecycle: create, join (by invite code), leaderboards, member management, direct invitations
- **tips** — Match predictions locked at kickoff; crowdsourcing aggregation
- **forecast** — Pre-tournament bracket/group predictions; locked once tournament starts
- **sync** — Fetches match results from API-Football or openfootball; manual admin overrides
- **standings** — Computes group standings and knockout bracket from results
- **odds** — Syncs bookmaker odds from The Odds API; falls back to FIFA-ranking estimates
- **account** — User deletion, personal stats
- **chat** — League chat with per-user read cursors
- **oauth** — Google OAuth integration
- **clock** — Server time abstraction; supports time simulation in dev mode (`WMP_DEV=1`)
- **dev** — Dev-only endpoints (bot generation, time override); only registered when `WMP_DEV=1`
- **web** — Serves embedded SvelteKit build; handles PWA manifest and invite link social metadata

### Frontend (`frontend/src/`)

SvelteKit SPA (no SSR, static adapter). File-based routing under `src/routes/`. Shared code in `src/lib/` (API client, auth store, theme). During dev, Vite proxies `/api/*` to the backend via `VITE_API_ORIGIN`.

### Data model highlights

Key PocketBase collections (schema defined in `migrations/0001_init.go` + 16 follow-on migrations):

- **matches** — 104 fixtures; `stage` values: `group`, `R32`, `R16`, `QF`, `SF`, `3RD`, `FINAL`
- **tips** — One per user/match; locked after `match.kickoff`
- **forecasts** — One per user; stores `groupOrder`, `thirdQualifiers`, `bracket` as JSON
- **scoring_configs** — Per-league JSON blob defining all point values
- **match_scores** / **forecast_scores** — Derived/computed; regenerated on every result change
- **league_invites** — Pending/accepted/declined; separate from `league_members`

### Environment variables (see `.env.example`)

| Var | Purpose |
|-----|---------|
| `HTTP_PORT` | Listening port (default 8090) |
| `WMP_DEV` | `1` enables dev-only endpoints and time simulation |
| `RESULTS_SOURCE` | `auto` / `openfootball` / `apifootball` |
| `API_FOOTBALL_KEY` | Optional; enables live result sync |
| `ODDS_API_KEY` | Optional; enables bookmaker odds sync |
| `PB_ADMIN_EMAIL` / `PB_ADMIN_PASSWORD` | Bootstrap admin account |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Optional Google OAuth |

### Frontend ↔ Backend contract

- All custom API routes are under `/api/` and registered in `main.go`
- PocketBase's built-in CRUD REST API is used directly for collections like `tips`, `league_chat`, `league_chat_reads`
- Auth uses PocketBase's `/api/auth/` endpoints; the frontend stores the token and passes it as `Authorization: Bearer <token>`

### Schema migrations

Migrations live in `migrations/` as Go files (not SQL). New migrations must be registered in `migrations/migrate.go` and applied via PocketBase's migration runner on startup.
