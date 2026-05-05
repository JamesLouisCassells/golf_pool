# Masters Pool

Masters Pool is a rewrite of a golf pool application being built as both a functioning app and a learning-focused portfolio project. The goal is not just to ship features, but to do the work in a way that shows professional engineering habits, clear reasoning, and steady iteration.

The current architecture and long-term direction live in `PLAN.md`. The execution checklist lives in `TODO.MD`. `AGENTS.md` explains how future agent work should stay aligned with both.

## Current Status

The repository currently has:

- a Go backend with config, auth, entry, and admin-config routes in `backend/`
- a Vue/Vite frontend with routed `Home`, `Enter`, `Entries`, `Standings`, and `Admin` views in `frontend/`
- Postgres migrations for `users`, `entries`, and `tournament_config`
- a health endpoint at `GET /healthz`
- a config endpoint at `GET /api/config/{year}`
- mock auth for local development and Clerk JWT validation for backend-protected routes
- entry create/edit/read flows and a public entries listing once the tournament starts
- an admin config form backed by admin-only config endpoints

This is no longer just scaffold state, but it is still incomplete. Live standings, frontend Clerk integration, and deployment are still pending.

## Repository Structure

```text
golf_pool/
├── backend/      # Go API, DB access, migrations
├── frontend/     # Vue 3 + Vite frontend app
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
- either mock-auth env vars or Clerk env vars for protected API routes

Mock auth and Clerk-related values are included in `.env.example`. Mock auth is still the temporary local-development path while the real Clerk browser flow is being wired.

The frontend now also expects its own Vite env file when you want the real Clerk browser flow:

```bash
cd frontend
cp .env.example .env
```

That file should include:

- `VITE_CLERK_PUBLISHABLE_KEY`

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

The frontend now has a routed app shell and initial data-backed views.

To run it locally:

```bash
cd frontend
npm install
npm run dev
```

What the current frontend does:

- provides a router with `Home`, `Enter`, `Entries`, `Standings`, and `Admin` routes
- provides an `Entries` route backed by the public entries endpoint
- provides an `Admin` route backed by admin-only config endpoints
- loads `GET /api/config/:year`
- loads `GET /api/entries/mine`
- loads `GET /api/entries`
- loads and saves `/api/admin/config/:year`
- submits `POST /api/entries` for a new entry
- submits `PUT /api/entries/:id` for an existing entry
- shows a countdown and locks the form after the deadline
- conditionally initializes the Clerk Vue SDK when a publishable key is present
- exposes login/logout controls in the app shell
- routes signed-out users to a dedicated sign-in screen in Clerk mode
- adds frontend guards for `/enter` and `/admin` in Clerk mode
- attaches Clerk bearer tokens to protected API requests from one shared fetch helper

What is still missing:

- live standings integration
- final end-to-end proof against a real Clerk instance
- polish around loading, error, and mobile states

## Current Auth Slice

The protected backend surface currently includes:

- `GET /api/me`
- `GET /api/entries/mine`
- `GET /api/entries`
- `POST /api/entries`
- `PUT /api/entries/:id`
- `GET /api/admin/config/:year`
- `PUT /api/admin/config/:year`

What it does today:

- supports a config-driven mock authenticated user for local development
- accepts either a bearer token or the `__session` cookie
- validates RS256 JWT signatures against Clerk JWKS
- validates Clerk `azp` against configured allowed origins when present
- upserts the local `users` row from token claims
- can fall back to the Clerk Backend API for the user profile if the session token does not include an email claim
- returns the local user record plus `is_admin`
- returns the authenticated user's entry for the active tournament year when one exists
- returns the public active-year entry list after the tournament has started
- creates an entry for the active tournament year before the deadline and blocks duplicates
- updates an owned active-year entry before the deadline and blocks cross-user edits
- allows admin-only reads and updates of tournament config
- allows admin-only list/edit/delete entry maintenance routes

What is still incomplete:

- the frontend Clerk wiring still needs real instance values
- no end-to-end proof has been run yet with a real Clerk browser session
- custom session claims still need to be configured in Clerk

## Revised Clerk Plan

The current plan is:

- use Clerk in the Vue frontend for sign-in/sign-up and browser session state
- keep backend request authentication lightweight by verifying Clerk session JWTs locally against JWKS
- avoid using Clerk's Backend API on every authenticated request
- use the local Postgres `users` table as the app's working copy of identity data
- add custom Clerk session claims for the fields the API needs on normal requests

That means the backend does not need a broad auth rewrite just to "use the Clerk SDK." The valuable part is getting the browser session flow in place and making backend request handling cheap.

### Intended request flow

1. The frontend initializes Clerk and signs the user in.
2. Protected frontend requests send Clerk session context to the API.
3. The Go backend verifies the session token locally using Clerk JWKS.
4. The Go backend reads needed identity and admin fields from token claims and upserts the local `users` row.
5. The app uses local database state for pool data instead of repeatedly calling Clerk.

### Why this plan matters

The free-tier concern is valid if the backend reaches out to Clerk's Backend API repeatedly. The safer design is to treat Clerk as the identity provider and keep request-time authorization mostly local:

- verify tokens locally
- keep Clerk claims small but sufficient
- only call Clerk's Backend API when token claims are missing something important

### Recommended custom session claims

For this app, the token should ideally carry:

- email
- display name
- admin role or admin flag

That removes the need to fetch a Clerk user profile during ordinary authenticated requests. Claims should stay small so cookie and session size does not become a problem.

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
- If you want email, display name, and admin state to come directly from the token, add custom session token claims in Clerk that match `CLERK_EMAIL_CLAIM`, `CLERK_NAME_CLAIM`, and the configured admin claim.
- If you do not add those claims, the backend can use `CLERK_SECRET_KEY` to fetch the user profile from Clerk's Backend API after token verification, but that should be the exception rather than the normal path.

Suggested local proof steps:

1. In Clerk, add small custom session claims for email, display name, and admin role if needed.
2. Run the backend with real Clerk values and set `MOCK_AUTH_ENABLED=false`.
3. In `frontend/.env`, set `VITE_CLERK_PUBLISHABLE_KEY` and sign in through the browser.
4. Call `GET /api/me` through the app flow and confirm the local `users` row is inserted or updated in Postgres.
5. Keep the Backend API profile fallback only for missing-claim edge cases.

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

- finish Clerk integration in the browser
- finish admin route-guarding in the frontend
- build standings ingestion and payout logic
- package local dev and deployment workflows

For the detailed step-by-step plan, use `TODO.MD`.
