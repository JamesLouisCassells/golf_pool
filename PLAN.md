# Masters Pool — Rewrite Architecture

*v1.0 — April 2026*

---

## Overview

A ground-up rewrite of the Masters Pool app. Goals: clean maintainable codebase, form-based entry creation, user login, admin panel, and live tournament standings — all running in Kubernetes.

**Key principles:**
- Go backend serving clean JSON — no templating, no server-side HTML
- Vue 3 frontend consuming that API — component-based, easy to learn, existing look preserved
- External auth via Clerk — no rolling our own login or session management
- PostgreSQL for all state — entries, users, tournament config
- Everything containerized and GitOps-deployed via Flux

---

## Tech Stack

### Backend — Go
- `chi` router for HTTP handling — lightweight, idiomatic
- `pgx` for Postgres — fast, well-maintained Go driver
- JWT middleware to validate Clerk tokens on protected routes
- Built-in scheduler for polling the golf scores API
- Returns JSON on all routes — no HTML rendering

### Frontend — Vue 3 + Vite
- Vue 3 with single-file components (`.vue`) — component-based, easy to learn
- Vite for build tooling — fast, minimal config
- Vue Router for client-side navigation
- Custom CSS is currently in use; Tailwind remains optional rather than adopted
- Clerk Vue SDK is now wired for auth state and login/logout
- The existing entry cards, side leaderboard, and standings layout are still intended to be ported into Vue components

### Auth — Clerk
- Handles all login flows: email/password, Google OAuth, GitHub OAuth
- Free tier up to 10,000 MAU
- Clerk is initialized in the Vue frontend so the browser owns sign-in and session state
- Go backend validates Clerk session JWTs locally against JWKS — no passwords or auth data stored locally
- The backend avoids depending on Clerk Backend API lookups during ordinary authenticated requests
- Small custom session claims carry the identity fields the API needs most often
- Admin role is a custom claim set in Clerk's dashboard, checked by Go middleware

### Database — PostgreSQL
- CloudNativePG (CNPG) operator running a StatefulSet in-cluster
- WAL archiving + base backups shipped to Backblaze B2 (S3-compatible, essentially free at this scale)
- Schema migrations managed with `golang-migrate`

### Infrastructure
- Two containers per pod: Go API (`:8080`) + Nginx (`:80`)
- Nginx serves the built Vue static files and proxies `/api/*` to the Go container over localhost
- Flux GitOps for all deployments — manifests live in `/deploy`

---

## Architecture

```
┌──────────────────────────────────────────────────┐
│                Kubernetes Cluster                │
│                                                  │
│  ┌───────────────────────────────┐               │
│  │            App Pod            │               │
│  │  ┌──────────┐  ┌───────────┐ │  ┌──────────┐ │
│  │  │  Nginx   │  │  Go API   │ │  │  CNPG    │ │
│  │  │  :80     │──│  :8080    │─┼─▶│ Postgres │ │
│  │  │ Vue dist │  │ /api/*    │ │  │ + B2 bkp │ │
│  │  └──────────┘  └───────────┘ │  └──────────┘ │
│  └───────────────────────────────┘               │
│                                                  │
│  Flux GitOps manages all resources               │
└──────────────────────────────────────────────────┘

External:
  Browser  → Clerk    (login UI / JWT issuance)
  Go API   → Clerk    (JWKS endpoint for JWT validation)
  Go API   → Clerk    (Backend API only when claims are missing and profile fallback is needed)
  Go API   → RapidAPI (golf scores, polled on schedule)
  CNPG     → Backblaze B2 (WAL + base backups)
```

---

## Project Structure

```
masters-pool/
├── backend/
│   ├── Dockerfile
│   ├── main.go
│   ├── migrations/           # SQL migration files
│   └── internal/
│       ├── api/              # HTTP handlers
│       ├── auth/             # Clerk JWT middleware
│       ├── db/               # pgx query functions
│       ├── golf/             # RapidAPI polling + payout calc
│       └── admin/            # admin-only handlers
│
├── frontend/
│   ├── vite.config.js        # proxy /api/* to :8080 in dev
│   ├── index.html
│   └── src/
│       ├── main.js
│       ├── App.vue
│       ├── router/index.js
│       ├── views/
│       │   ├── Home.vue
│       │   ├── Standings.vue
│       │   ├── Enter.vue      # entry creation / edit
│       │   ├── Entries.vue    # all entries view
│       │   └── Admin.vue
│
├── deploy/
│   ├── app/
│   │   ├── deployment.yaml   # two-container pod
│   │   ├── service.yaml
│   │   ├── ingress.yaml
│   │   └── configmap.yaml    # nginx config, env vars
│   └── postgres/
│       └── cluster.yaml      # CNPG cluster + B2 backup config
│
└── backend/Dockerfile
```

---

## Database Schema

### users
Thin local record keyed to the Clerk user ID. No passwords stored here.

```sql
clerk_id     TEXT PRIMARY KEY
email        TEXT NOT NULL
display_name TEXT
created_at   TIMESTAMPTZ DEFAULT now()
```

### entries
One row per participant per year. Picks stored as JSONB so the group structure can vary year to year without a migration.

```sql
id           UUID PRIMARY KEY DEFAULT gen_random_uuid()
year         INT NOT NULL
clerk_id     TEXT REFERENCES users(clerk_id)
display_name TEXT NOT NULL
picks        JSONB NOT NULL  -- { "Group 1": "Scheffler", "WC": "Rose", ... }
in_overs     BOOLEAN DEFAULT false
locked       BOOLEAN DEFAULT false
created_at   TIMESTAMPTZ DEFAULT now()
updated_at   TIMESTAMPTZ DEFAULT now()
```

### tournament_config
One row per year. Controls all the moving parts — editable from the admin panel, no code changes needed to run a new year.

```sql
year                 INT PRIMARY KEY
entry_deadline       TIMESTAMPTZ
start_date           DATE
end_date             DATE
groups               JSONB     -- group names + player lists
mutt_multiplier      NUMERIC DEFAULT 2
old_mutt_multiplier  NUMERIC DEFAULT 3
pool_payouts         JSONB     -- { "1": 4475, "2": 2640, ... }
frl_winner           TEXT      -- set once first round is complete
frl_payout           INT DEFAULT 500000
active               BOOLEAN DEFAULT false
```

---

## API Routes

### Public — no auth required

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/standings/:year` | Live leaderboard with payouts |
| GET | `/api/config/:year` | Tournament dates, group definitions |
| GET | `/api/entries` | View all entries after tournament start |

### Authenticated — any logged-in user

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/entries` | Create entry (before deadline) |
| PUT | `/api/entries/:id` | Edit own entry (before deadline) |
| GET | `/api/entries/mine` | View my entry |
| GET | `/api/me` | Return the current authenticated user |

### Admin only

| Method | Route | Description |
|--------|-------|-------------|
| GET/PUT | `/api/admin/config/:year` | Manage tournament config |
| GET | `/api/admin/entries` | List all entries |
| PUT | `/api/admin/entries/:id` | Edit any entry |
| DELETE | `/api/admin/entries/:id` | Remove entry |
| POST | `/api/admin/refresh` | Trigger manual score refresh |
| POST | `/api/admin/lock` | Lock all entries immediately |

---

## Frontend Pages

| Route | View | Auth |
|-------|------|------|
| `/` | Home — links to years and current standings | Public |
| `/standings` | Live leaderboard, auto-refreshes | Public |
| `/enter` | Create or edit my entry | Intended authenticated flow; currently workable via mock auth |
| `/entries` | All entries for the year | Public (after tournament start) |
| `/admin` | Admin panel | Backend-enforced admin role; frontend route guard still pending |

Login/logout is handled by Clerk's hosted UI — no custom login page needed.

For routine authenticated API traffic, the intended design is:

- Clerk session established in the frontend
- session token presented to the API by cookie or bearer token
- backend verifies the token locally
- backend upserts or reads the local `users` row
- backend avoids per-request Clerk profile lookups

---

## Local Development

Target local development still assumes three services:

- `postgres` — standard Postgres image
- `api` — Go binary with hot-reload via `air`
- `web` — Vite dev server with `/api/*` proxied to the Go container

Today, local feature work is usually unblocked with mock auth while Clerk browser wiring is still pending.

The target next step is to replace that temporary path with:

- Clerk Vue SDK in the frontend
- real browser sign-in locally
- `MOCK_AUTH_ENABLED=false` during Clerk proof work
- custom session claims so the backend rarely needs Clerk Backend API calls

---

## Deployment

### CI — GitHub Actions
- On push to `main`: build Go binary, build Vue app, produce two Docker images, push to GHCR
- Tags: `sha-<short>` for Flux image automation, semver from git tags
- Two images: `ghcr.io/dgunzy/masters-pool-api` and `.../masters-pool-web`

### CD — Flux
- Flux watches GHCR for new image tags and updates the deployment automatically
- All year-to-year config lives in a ConfigMap and the `tournament_config` table — no image rebuild to run a new year
- Secrets (RapidAPI key, Clerk keys, DB credentials) managed as Sealed Secrets

### Postgres Backups
- CNPG ships WAL archives continuously to Backblaze B2
- Scheduled base backups daily
- Point-in-time recovery available from any backup
