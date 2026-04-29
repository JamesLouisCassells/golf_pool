# Masters Pool

Masters Pool is a rewrite of a golf pool application being built as both a functioning app and a learning-focused portfolio project. The goal is not just to ship features, but to do the work in a way that shows professional engineering habits, clear reasoning, and steady iteration.

The current architecture and long-term direction live in `PLAN.md`. The execution checklist lives in `TODO.MD`. `AGENTS.md` explains how future agent work should stay aligned with both.

## Current Status

The repository currently has:

- a Go backend scaffold in `backend/`
- a Vue/Vite frontend scaffold in `frontend/`
- Postgres migrations for `users`, `entries`, and `tournament_config`
- a health endpoint at `GET /healthz`
- a config endpoint at `GET /api/config/{year}`

This is still an early project state. Auth, entry workflows, standings logic, admin features, and deployment are not complete yet.

## Repository Structure

```text
golf_pool/
├── backend/      # Go API, DB access, migrations
├── frontend/     # Vue 3 + Vite frontend scaffold
├── deploy/       # Deployment manifests, to be filled in later
├── PLAN.md       # Target architecture and technical direction
├── TODO.MD       # Phased execution checklist
└── AGENTS.md     # Working rules for future agent contributions
```

## Prerequisites

Install these tools locally:

- Go
- Docker Desktop
- Homebrew
- `golang-migrate`

Useful checks:

```bash
go version
docker --version
brew --version
migrate -version
```

If `golang-migrate` is missing:

```bash
brew install golang-migrate
```

## Environment Setup

Create your local environment file from the example:

```bash
cp .env.example .env
```

The backend currently expects:

- `HTTP_ADDR`
- `DATABASE_URL`
- either mock-auth env vars or `CLERK_JWKS_URL` for protected API routes

Mock auth and Clerk-related values are included in `.env.example`. Mock auth is the current local-development path while Clerk wiring is still in progress.

## Running the Backend

Start Postgres in Docker:

```bash
docker run --name golf-pool-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=golf_pool \
  -p 5432:5432 \
  -d postgres:17
```

If the container already exists, use:

```bash
docker start golf-pool-postgres
```

Apply migrations:

```bash
cd backend
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/golf_pool?sslmode=disable" up
```

Run the API:

```bash
cd backend
DATABASE_URL="postgres://postgres:postgres@localhost:5432/golf_pool?sslmode=disable" go run ./cmd/api
```

Test the health endpoint:

```bash
curl http://localhost:8080/healthz
```

Test the config endpoint after inserting a row into `tournament_config`:

```bash
curl http://localhost:8080/api/config/2026
```

## Frontend

The frontend was scaffolded with Vue and Vite, but it is still at the starter-template stage.

To run it locally:

```bash
cd frontend
npm install
npm run dev
```

The frontend is not yet integrated with the backend API in a meaningful way.

## Current Auth Slice

The first protected backend route now exists:

- `GET /api/me`
- `GET /api/entries/mine`
- `POST /api/entries`
- `PUT /api/entries/:id`

What it does today:

- supports a config-driven mock authenticated user for local development
- accepts either a bearer token or the `__session` cookie
- validates RS256 JWT signatures against Clerk JWKS
- validates Clerk `azp` against configured allowed origins when present
- upserts the local `users` row from token claims
- falls back to the Clerk Backend API for the user profile if the session token does not include an email claim
- returns the local user record plus `is_admin`
- returns the authenticated user's entry for the active tournament year when one exists
- creates an entry for the active tournament year before the deadline and blocks duplicates
- updates an owned active-year entry before the deadline and blocks cross-user edits

What is still incomplete:

- no frontend auth integration yet
- no admin-only routes are wired yet, even though the initial admin middleware exists
- no local helper exists yet for generating or capturing a dev token flow

## Mock Auth Dev Setup

For now, local development can use mock auth instead of Clerk:

- set `MOCK_AUTH_ENABLED=true`
- set the mock user values in `.env`
- run the backend normally

When mock auth is enabled, protected routes behave as if the configured mock user is logged in. That keeps pool workflows moving without coupling feature work to Clerk setup timing.

## Clerk Dev Setup

To finish the real auth proof locally, create `.env` from `.env.example` and fill in these Clerk values from your development instance:

- `CLERK_SECRET_KEY`
- `CLERK_JWKS_URL`
- `CLERK_ISSUER`
- `CLERK_AUTHORIZED_PARTIES`

Optional values:

- `CLERK_EMAIL_CLAIM`
- `CLERK_NAME_CLAIM`
- `CLERK_ADMIN_CLAIM`
- `CLERK_ADMIN_VALUE`

Important:

- Clerk's default session token claims include `sub`, `iss`, `exp`, `nbf`, and often `azp`, but not `email` by default.
- If you want email and display name to come directly from the token, add custom session token claims in Clerk that match `CLERK_EMAIL_CLAIM` and `CLERK_NAME_CLAIM`.
- If you do not add those claims, the backend will use `CLERK_SECRET_KEY` to fetch the user profile from Clerk's Backend API after token verification.

Suggested local proof steps:

1. Run the backend with your real Clerk values.
2. Sign in through the frontend once Clerk is wired, or manually capture a valid session token from your Clerk dev app.
3. Call `GET /api/me` with either the `Authorization: Bearer <token>` header or the `__session` cookie.
4. Confirm the `users` row is inserted or updated in Postgres.

## Learning Notes

This project is being built with two priorities:

1. Learn each layer of the stack in a professional way.
2. End with a real app that is worth showing in a portfolio.

That means the repo should show not just code, but decision-making:

- why a piece of structure exists
- how backend responsibilities are separated
- how planning documents stay current as the implementation changes

## Next Steps

The next major work areas are:

- finish Phase 2 database functions
- add Clerk auth scaffolding
- build entry creation and tournament config workflows
- connect the frontend to real backend data

For the detailed step-by-step plan, use `TODO.MD`.
