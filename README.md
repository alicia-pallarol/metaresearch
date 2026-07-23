# AI Safety Agendas × Problems Map

An interactive, manually curated map of **58 AI safety research agendas** across
**12 research areas**, rated against **12 open problems**, plus a small feedback
endpoint so researchers can tell us where the ratings are wrong.

Every cell answers one question: *how mature is the evidence that this line of work
addresses this specific problem?* Nothing on the site is model-generated, and the
running application never calls a model API. That is a credibility requirement, not
only a cost one.

---

## Architecture

```
[ Vue 3 SPA · Cloudflare Pages ]
        │
        ├── data/atlas.json  ─────────  bundled at build time.
        │                               The whole map renders from this file, so the
        │                               site works with the API down or asleep.
        │
        └── HTTPS ──▶ [ Go API · Render ] ──▶ [ Postgres · Neon ]
                        POST /api/feedback        the ONLY write path
                        GET  /api/summary         aggregates only, cached ~45s
                        GET  /api/health          also wakes a cold instance
                        GET  /api/admin/export    bearer token, raw rows
```

The split is the point:

- **The curated map is static.** It ships in the frontend bundle, is free to serve,
  and cannot be altered through the API, there is no endpoint that writes it.
- **The database holds nothing but researcher feedback.** One table, one write path.
- **The public only ever reads aggregates.** No individual submission is reachable
  from the web; names, emails and notes leave the database only through the
  admin-token export or the Neon SQL console.

## Repository layout

```
data/atlas.json          the curated dataset, authoritative, committed, the only required input
data/source/*.xlsx       OPTIONAL provenance workbooks (absent here; not needed to build)
swiss_layers.json        interpretive Swiss-cheese config, empty and inert by default (see below)

backend/                 Go 1.22+ service: main.go, internal/{api,store,config,atlas}, migrations/, Dockerfile
frontend/                Vue 3 + Vite + TypeScript app

scripts/validate_atlas.py    checks every invariant in atlas.json
scripts/gen_agenda_ids.py    regenerates the agenda id list the backend validates against
scripts/build_atlas.py       optional regenerator from the .xlsx workbooks (no-op when absent)

MANUAL_SETUP.md          the operator runbook: from zero accounts to a live site
```

## Local development

Requirements: Go 1.22+, Node 20+, Python 3.10+ (only for the data scripts).

### Backend

```bash
cd backend
cp .env.example .env          # then edit
go run .
```

`go run .` loads `backend/.env` automatically (via `godotenv`), Go does not read
`.env` files on its own, so this is what makes `cp .env.example .env` actually take
effect for local dev. It is a no-op wherever the file is absent (Render, Docker,
CI) and never overrides a variable the real environment already set.

`ALLOWED_ORIGIN` is required (use `http://localhost:5173` for Vite). `DATABASE_URL`
is optional locally: without it the service starts in **degraded mode**, the map
and every read still work, `/api/health` reports `degraded`, and feedback writes
answer 503. That is the fastest way to work on the frontend.

With a database (a free Neon branch is easiest), migrations apply automatically on
boot.

```bash
go vet ./...
go test ./...            # add -race on an amd64 toolchain; see "Known limitation"
gofmt -l .               # should print nothing
```

There is no CI workflow yet, the project is early enough that a human running
these commands before pushing is enough. Worth adding once there's a shared
remote and more than one contributor.

### Frontend

```bash
cd frontend
cp .env.example .env          # VITE_API_BASE=http://localhost:8080
npm install
npm run dev                   # http://localhost:5173
```

```bash
npm run typecheck             # vue-tsc
npm test                      # vitest
npm run build                 # typecheck + production build into dist/
```

### Data

```bash
python scripts/validate_atlas.py     # every invariant, with what it checked
python scripts/gen_agenda_ids.py     # after any change to the agenda list
```

## How to cut a new iteration

1. Edit `data/atlas.json`, either by hand, or with
   `python scripts/build_atlas.py --write` if you keep the source workbooks in
   `data/source/` (see the script's docstring for its limitations).
2. Bump `meta.iteration` and `meta.iteration_label`, and update
   `meta.source_check_date`.
3. Run `python scripts/validate_atlas.py`. It fails loudly on drift: wrong counts, a
   tier outside `tier_order`, a condensed cell that no longer matches the agenda
   grid it summarises.
4. Run `python scripts/gen_agenda_ids.py`, then `cd backend && gofmt -w internal/atlas/ids.go && go test ./...`.
   The backend's id set is what stops feedback arriving for agendas that do not exist.
5. Set `ITERATION` to the new number on the Render service so new submissions are
   stamped with it. Old feedback keeps its own iteration number.
6. Redeploy the frontend (a push to the default branch does this on Cloudflare Pages).

The headline counts in the hero, the legend and the tier tallies are all derived
from the dataset, so none of them need a code change.

## Design decisions worth knowing

**Two colour scales, never blended.** Maturity uses a single-hue blue ramp for the
four tiers that form an evidence axis (Untested → Early/partial → Strong existence
proof → Robust). *Contested* and *Never demonstrated* are **not** rungs on that
ladder, one is a live dispute, the other a negative finding, so they take
off-axis hues plus a diagonal hatch. Blank is not a tier and gets no fill.
Community familiarity uses a separate green ramp. Every step was checked with a
palette validator for lightness monotonicity, CVD separation and contrast against
both the light and dark surfaces; the reasoning lives in `frontend/src/lib/scales.ts`.

**The condensed grid always shows its agenda count**, because a cell is only as
strong as its best agenda, and one strong agenda in a thin area is a different claim
from one strong agenda among six.

**Aggregates span iterations.** `/api/summary` counts every submission ever made,
not just the current iteration's: a researcher's familiarity with an agenda does not
expire when the map is re-cut, and the contributor counter should not drop to zero.
Each row still records the iteration it was given against, so the split is available
in the export.

**Feedback is a maturity vote, not a verdict on ours.** A submission carries the
reader's own maturity rating for a research area (`area_maturity`, required, one of
the six tiers) and optionally a rating per agenda inside it (`subarea_maturity`, an
agenda→tier map). This replaced an agree/disagree question, which bought one bit and
anchored the reader on our rating; a rating in the map's own vocabulary can be
aggregated, disagreed with, and drawn. Maturity is a property of the *research*, how
much work exists and what it has shown, not of a problem, so a vote is scoped to an
area even when the form is opened from one Area × Problem cell. The `target`
(`agenda_id`, or `area_tag` + `problem_id`) is only provenance, which page produced
the answer; the area a vote lands in is derived from the agenda server-side, never
taken on trust. `agree_with_rating` is retired but not dropped: rows written while it
was asked keep their answers.

**The community view of the grid substitutes votes into the curated skeleton.**
Which agendas address which problem stays our claim; `GET /api/summary` returns a
per-area and a per-agenda maturity histogram, and the front end fills the same cells
with community tiers. Best/typical/worst-case bands are quantiles of those
histograms summarised by their median, never a mean, because two tiers are off the
evidence axis. Below eight ratings a band is a single rating, so *research · best
case* reproduces the historical "best of N agendas" grid exactly.

**Focus areas are computed, not curated.** The old hand-written priority tiers are
replaced by a classifier (`frontend/src/lib/focus.ts`) that reads the grid, under
whichever source the grid is set to, and sorts research areas, and separately
problems, into *easy to intervene* / *bottleneck* / *underexplored*. It excludes
what is already near-Robust, scores each unit on three lenses, and files it under
its strongest. Every threshold and weight is a named constant with a comment, so
the rule is editable in one place; the per-item reasons are assembled from the
unit's own numbers, never generated as prose (the no-runtime-LLM rule holds here
too). The curated `priorities` block stays in `atlas.json`, it is just no longer
rendered. On the current data the classifier independently reproduces the curated
editorial read (the measurement hub P1 and the oversight/foundations areas land as
bottlenecks; the sharp-left-turn problem lands as underexplored), which is the main
evidence it is doing something real.

**Anonymity is enforced server-side.** If `is_anonymous` is set, the API blanks
name, email and contact consent before the row is written, whatever the client sent.
A name is optional and may be a nickname, pseudonymous feedback is a first-class
middle ground between fully anonymous and fully named.

**No raw IPs.** Stored as a salted SHA-256. With `IP_HASH_SALT` unset a random salt
is generated per boot, so hashes are not even comparable across restarts, the more
private default.

## Known limitation

`go test -race` needs an amd64 toolchain with a C compiler; on a 386 Go build (no
C compiler) the race detector cannot run. The concurrent paths (the rate limiter
and the summary cache) have dedicated concurrency tests that are written to catch
races under `-race` wherever it's available, run them there before trusting a
change to either file.

## Deployment

See **[MANUAL_SETUP.md](MANUAL_SETUP.md)** for the full runbook: Neon → Render →
Cloudflare Pages, with no account, no credit card and no domain assumed. All three
tiers are free; the map costs nothing to serve and the API sleeps when unused.

## The Swiss-cheese view

`swiss_layers.json` is empty, and the feature is hidden until it is not. This is
deliberate. A defence-in-depth diagram needs an *ordered* set of layers that are
*roughly independent*; the dataset supplies neither, and its own reading note argues
the defences are correlated, the load-bearing problems all lean on measurement and
oversight, and the single Robust cell inherits its reliability from ordinary systems
security rather than from anything we understand about models. Only the hole sizes
are derived from the matrix. Authoring the layers is a research claim; the file and
`MANUAL_SETUP.md` say what it would take to make one honestly.

## Licence

MIT for the code. The curated dataset in `data/atlas.json` is the authors' research
judgment, see [LICENSE](LICENSE).
