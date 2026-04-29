# Masters Pool — Development TODO

Work through these phases in order. Each phase should leave the app in a working state before moving on.

---

## Phase 1 — Project Setup

Get the skeleton in place before writing any real logic.

### Repo & tooling
- [x] Create new GitHub repo
- [x] Set up directory structure: `/backend`, `/frontend`, `/deploy`
- [x] Add `.gitignore`, `.env.example`, `README.md`
- [x] Invite collaborator, agree on branch strategy (main + feature branches)

### Go backend scaffold
- [x] `go mod init`, install `chi`, `pgx`, `golang-jwt`
- [x] `main.go`: HTTP server, health endpoint `GET /healthz`
- [x] Set up `internal/` package structure: `api/`, `auth/`, `db/`, `golf/`, `admin/`
- [x] `Dockerfile` for backend (Alpine)

### Vue frontend scaffold
- [x] `npm create vite@latest` with Vue template
- [ ] Install Tailwind CSS, Clerk Vue SDK
- [x] Stub out views: `Entries`, `Admin`
- [x] Stub out views: `Home`, `Standings`, `Enter`
- [x] Set up `router/index.js` with routes for each view
- [x] `vite.config.js`: proxy `/api/*` to `localhost:8080` in dev
- [ ] `Dockerfile` for frontend (node build stage → nginx:alpine)
- [ ] `nginx.conf`: serve static files, proxy `/api/*` to `localhost:8080`

### Local dev environment
- [ ] `docker-compose.yml`: `postgres` + `api` + `web` containers
- [ ] Confirm all three start and nginx proxies to Go correctly
- [ ] Add `Makefile` targets: `make dev`, `make build`, `make test`

---

## Phase 2 — Database & Auth

### Postgres + migrations
- [x] Install `golang-migrate` CLI
- [x] Migration 001: create `users` table
- [x] Migration 002: create `entries` table
- [x] Migration 003: create `tournament_config` table
- [x] Add DB connection pool to Go (`pgxpool`), log success on startup
- [x] Add typed config and user queries in `db/` (`GetConfig`, `GetUser`, `UpsertUser`)
- [x] Add typed entry lifecycle and listing queries (`CreateEntry`, entry lookup/update, active-year list)
- [ ] Add remaining typed query functions (`config update`, standings reads, admin entry operations)

### Clerk auth
- [ ] Create Clerk account, set up application, enable Google + email/password providers
- [ ] Add Clerk publishable key and secret key to `.env`
- [x] Write initial `auth/` middleware: fetch Clerk JWKS and validate JWT on incoming requests
- [x] Write initial `requireAdmin` middleware: check for admin role claim in token
- [x] Add temporary mock auth mode so protected routes can be developed before Clerk is fully wired
- [ ] Test end-to-end with a real valid Clerk token and confirm local `users` upsert
- [ ] Add Clerk Vue SDK to frontend — implement login/logout, expose auth state app-wide
- [x] On authenticated API requests, upsert a `users` row using the Clerk user ID and email

---

## Phase 3 — Core Features

### Tournament config
- [ ] Seed `tournament_config` for the active year (migration or admin endpoint)
- [x] `GET /api/config/:year` — return group definitions, player lists, deadline
- [x] Admin page: form to view and edit config (deadline, multipliers, pool payouts, FRL winner)
- [x] `PUT /api/admin/config/:year` — save config changes

### Entry creation & editing
- [x] `POST /api/entries` — create entry, enforce deadline check
- [x] `PUT /api/entries/:id` — edit own entry, enforce deadline + ownership check
- [x] `GET /api/entries/mine` — return my entry for the current year
- [x] `GET /api/entries` — public list after tournament start
- [x] `Enter.vue`: form built dynamically from config groups, pre-filled if entry already exists
- [x] Show deadline countdown on the entry form
- [x] Disable form and show a locked message after deadline passes

### Golf scores + payout calculation
- [ ] Port RapidAPI polling logic from Node.js to Go (`golf/` package)
- [ ] Port payout calculation and tie distribution logic to Go
- [ ] Port golfer name matching logic to Go
- [ ] Implement poll interval logic (active tournament hours vs. off-season)
- [ ] Apply mutt/old-mutt multipliers and FRL payout from `tournament_config`
- [ ] `GET /api/standings/:year` — return entries sorted by total payout with per-group breakdown

### Standings page
- [ ] `Standings.vue`: fetch standings, auto-refresh every 5 minutes
- [ ] Port `EntryCard` from existing JS into `EntryCard.vue` — expand/collapse, golfer scores, payout per group
- [ ] Port side leaderboard into `SideLeaderboard.vue`
- [ ] `Countdown.vue`: show countdown when tournament hasn't started
- [ ] `GolferBadge.vue`: score, position, cut indicator — reuse existing styling

---

## Phase 4 — Admin Panel

- [ ] `GET /api/admin/entries` — list all entries for the active year
- [ ] `PUT /api/admin/entries/:id` — edit any entry (admin override)
- [ ] `DELETE /api/admin/entries/:id` — delete entry
- [ ] `POST /api/admin/lock` — lock all entries (sets deadline to now)
- [ ] `POST /api/admin/refresh` — manually trigger a score data refresh
- [ ] `Admin.vue`: entry table with edit/delete actions, config form, refresh + lock buttons
- [ ] Guard `/admin` route in Vue Router — redirect non-admins to home
- [ ] Guard admin API routes in Go middleware — return `403` for non-admins

---

## Phase 5 — Kubernetes Deploy

### GitHub Actions CI
- [ ] Workflow: build and push `masters-pool-api` image to GHCR on push to `main`
- [ ] Workflow: build and push `masters-pool-web` image to GHCR on push to `main`
- [ ] Tag with `sha-<short>` and semver from git tags

### Kubernetes manifests
- [ ] `deploy/app/deployment.yaml` — two-container pod (Go API + Nginx)
- [ ] `deploy/app/service.yaml` + `ingress.yaml` — expose to internet with TLS
- [ ] `deploy/app/configmap.yaml` — Nginx config, `ACTIVE_YEAR`, tournament env vars
- [ ] Sealed Secret — RapidAPI key, Clerk keys, DB connection string
- [ ] `deploy/postgres/cluster.yaml` — CNPG cluster with Backblaze B2 backup config
- [ ] Confirm Flux picks up new image tags and deploys cleanly end-to-end

---

## Phase 6 — Data Migration & Polish

- [ ] Write a script to import 2024 and 2025 entries from existing CSVs into the DB
- [ ] Verify historical standings pages render correctly for both years
- [x] `Entries.vue`: grid of all entries for the active year (visible after tournament starts)
- [ ] Navigation: consistent nav bar across all pages with login/logout state
- [ ] Mobile responsive check on all pages
- [ ] Basic error handling — friendly messages for API failures and loading states
- [ ] Smoke test full flow: register, create entry, view standings, admin edits entry, lock all entries
