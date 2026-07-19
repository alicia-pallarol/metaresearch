# AI Safety Research Familiarity

A public platform where researchers record their familiarity (0–3) with each area
of a framework of AI safety research. Submissions are aggregated and visualized as
an interactive heatmap, and mapped onto a Swiss cheese defense model.

- **Framework page** — the research areas as an interactive, expandable grid and a
  heatmap of average community familiarity, with honest sample-size reporting.
- **Submission flow** — a compact 0–3 matrix with anonymity handling, server-side
  validation, a honeypot, and rate limiting. No accounts, no CAPTCHA.
- **Swiss cheese page** — an interactive, familiarity-linked defense model. Its
  conceptual mapping is AI-drafted and labeled as not yet peer-reviewed.

## Stack

| Layer     | Technology |
|-----------|------------|
| Frontend  | Vue 3 (`<script setup>`), Vue Router, Vite, hand-rolled SVG/CSS |
| Backend   | Go (`net/http` + chi), REST JSON, `pgx` |
| Database  | Postgres (Neon in production) |
| Hosting   | Cloudflare Pages (frontend) · Render (backend) · Neon (DB) — all free tier |

Deployment to those services is documented step by step in
[`DEPLOYMENT.md`](DEPLOYMENT.md).

## Repository layout

```
.
├── backend/            Go API (config, store, migration runner, handlers, tests)
├── frontend/           Vue 3 SPA (Vite)
├── migrations/         Plain SQL migrations (schema + seed), applied on startup
├── docs/               Source material (Swiss cheese mapping)
├── Dockerfile          Backend image (bundles migrations)
├── docker-compose.yml  Local Postgres + API
├── render.yaml         Optional Render blueprint
└── .env.example        Backend configuration contract
```

## Local development

### Prerequisites
- Go 1.21+ and Node.js 18+ (for running outside Docker)
- Docker + Docker Compose (recommended for the database)

### 1. Start Postgres and the API

The quickest path uses Docker Compose, which starts Postgres and the API and runs
migrations automatically:

```bash
cp .env.example .env          # optional; compose sets its own dev values
docker compose up --build
# API on http://localhost:8080  (GET /api/health returns {"status":"ok"})
```

### No Docker? One-command backend (userspace Postgres)

If you don't have Docker, this helper boots a throwaway Postgres inside the repo
(no root, no system install) and runs the Go API with migrations applied:

```bash
python3 scripts/dev-stack.py
# API on http://localhost:8080 ; database persists in ./.localpg
```

Leave it running and start the frontend in a second terminal (step 2 below). This
is the quickest way to get the table/heatmap populated locally.

### Prefer running the Go API on the host yourself?

Start only the database and export the backend env vars:

```bash
docker compose up -d db
cd backend
export DATABASE_URL='postgres://saige:saige@localhost:5432/saige?sslmode=disable'
export ALLOWED_ORIGIN='http://localhost:5173'
export IP_HASH_SALT='local-dev-salt'
export MIGRATIONS_PATH='../migrations'
go run .
```

### 2. Start the frontend

```bash
cd frontend
npm install
cp .env.example .env          # VITE_API_BASE_URL defaults to http://localhost:8080
npm run dev
# App on http://localhost:5173
```

Open http://localhost:5173. With no submissions you'll see the empty state and the
full framework; submit a response to watch the heatmap fill in.

## API

All responses are JSON. Emails and individual responses are **never** exposed.

| Method | Path                   | Purpose |
|--------|------------------------|---------|
| GET    | `/api/health`          | Liveness + DB check (used for cold-start detection) |
| GET    | `/api/research-areas`  | Framework + per-area aggregates (n, average, distribution) |
| POST   | `/api/submissions`     | Create/replace a submission (dedup by email) |
| GET    | `/api/contributors`    | Non-anonymous names only; disabled unless `EXPOSE_CONTRIBUTORS=true` |

### Submission body

```json
{
  "name": "Ada Lovelace",
  "email": "ada@example.com",
  "anonymous": false,
  "contact_consent": true,
  "website": "",
  "ratings": { "EA": 2, "IN": 1, "SA": null }
}
```

- `email` is required and validated; a repeat email **replaces** the prior response.
- `anonymous: true` forces the stored name empty and contact consent off.
- `website` is a honeypot — it must stay empty; filled submissions are silently
  discarded.
- `ratings` maps a research-area tag to `0`–`3`; omit a tag or send `null` to skip.

## Configuration

Backend reads config from environment variables only (see [`.env.example`](.env.example)):
`DATABASE_URL`, `ALLOWED_ORIGIN`, `PORT`, `IP_HASH_SALT`, and optional
`LOW_SAMPLE_THRESHOLD`, `RATE_LIMIT_PER_HOUR`, `MIGRATIONS_PATH`,
`EXPOSE_CONTRIBUTORS`.

Frontend reads `VITE_API_BASE_URL` at build time (see `frontend/.env.example`).

## Tests

```bash
cd backend
go test ./...                                   # unit tests (validation, rate limiter)

# Integration tests (aggregation smoke test, handler happy path, dedup) need a DB:
docker compose up -d db
TEST_DATABASE_URL='postgres://saige:saige@localhost:5432/saige?sslmode=disable' go test ./...
```

Unit tests run with no database; integration tests skip cleanly unless
`TEST_DATABASE_URL` is set.

## Security notes

- All SQL is parameterized; input is validated server-side.
- CORS is locked to `ALLOWED_ORIGIN`; security headers are set on both the API and
  the frontend (`frontend/public/_headers`).
- Only a salted, irreversible hash of the submitter IP is stored, for rate limiting.
- No secrets are committed; configuration is entirely env-driven.

## License

Provided as-is for the maintainer's use. Add a license file before publishing.
