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

Clerk keys are included in `.env.example` now so the repo is ready for upcoming auth work, but the auth middleware is not implemented yet.

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
