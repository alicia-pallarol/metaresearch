# Operator runbook, getting this live

Everything you have to do by hand, in order, assuming you have **no account with any
of these providers**. All three tiers are free, none asks for a credit card, and no
domain is needed: you will end up on `*.pages.dev` and `*.onrender.com`.

Budget about 45 minutes the first time. Nothing here is reversible-hard; if a step
goes wrong you can delete the resource and redo it.

You will create, in this order:

| # | Provider | What | Free tier |
|---|----------|------|-----------|
| 1 | [Neon](https://neon.tech) | Postgres database | 0.5 GB, permanent, no card |
| 2 | GitHub | the repo both deploys read from |, |
| 3 | [Render](https://render.com) | the Go API | 512 MB, no card, **sleeps when idle** |
| 4 | [Cloudflare Pages](https://pages.cloudflare.com) | the site | unlimited bandwidth, no card |

> **Why Neon and not Render's own Postgres:** Render's free Postgres **expires after
> 90 days** and takes your data with it. Neon's free tier is permanent. Keep the API
> on Render and the database on Neon.

---

## Step 0, Put the repo on GitHub

Both Render and Cloudflare deploy from a Git repository.

```bash
cd ai-safety-atlas
git init                      # if it is not a repo yet
git add .
git commit -m "AI Safety Agendas x Problems Map"
gh repo create ai-safety-atlas --private --source=. --push
```

Without the `gh` CLI: create an empty repository on github.com, then

```bash
git remote add origin https://github.com/<you>/ai-safety-atlas.git
git branch -M main
git push -u origin main
```

Private is fine, both providers can read private repos once you authorise them.

Check before you push: `git status` should show no `.env` file. Only `.env.example`
belongs in the repo.

---

## Step 1, Create the database (Neon)

1. Go to <https://neon.tech> and sign up (GitHub login works, no card).
2. **Create a project.** Name it `ai-safety-atlas`. For region pick **EU (Frankfurt)**:
   the audience and the operator are in the EU, and the privacy notice on the site
   says the data is stored in the EU.
3. On the project dashboard, open **Connection string** and pick the **Pooled
   connection** (it says "Pooled connection" or shows `-pooler` in the host). Copy it.
   It looks like:

   ```
   postgresql://user:PASSWORD@ep-something-pooler.eu-central-1.aws.neon.tech/neondb?sslmode=require
   ```

   Keep this somewhere safe for Step 3. It is a password, do not commit it.

### Apply the schema

There are two ways, and **you only need one**. They cannot conflict: the migration is
written to be idempotent, so whichever runs second finds the table already there.

**(a) By hand, now, recommended, so you can see it worked.**
In the Neon console open **SQL Editor** and run each file in
[`backend/migrations/`](backend/migrations/) in order, paste
[`0001_init.sql`](backend/migrations/0001_init.sql), run it, then
[`0002_cell_feedback.sql`](backend/migrations/0002_cell_feedback.sql), run it.
Both are idempotent, so re-running is harmless. Then check:

```sql
select column_name, data_type, is_nullable from information_schema.columns
where table_name = 'feedback' order by ordinal_position;
```

You should see 16 columns ending with `created_at`, including `area_tag` and
`problem_id` (both nullable), and `agenda_id` should be nullable too.

**(b) Automatically, on first boot.**
Do nothing. The API applies every migration in `backend/migrations/` when it starts,
in order, inside an advisory lock, and records what it applied in a
`schema_migrations` table. If you skipped (a), this is what will happen in Step 3,
check the Render deploy log for `"migration applied"` (once per new migration).

---

## Step 2, Generate the two secrets you need

```bash
# Admin token: guards GET /api/admin/export
openssl rand -hex 32

# IP hash salt: optional, see below
openssl rand -hex 32
```

No `openssl`? In PowerShell:

```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Max 256) })
```

About the salt: leave `IP_HASH_SALT` **unset** unless you have a reason not to. Unset
means a fresh random salt at every boot, so stored IP hashes are not comparable
across restarts, the most private option, and it costs you only the ability to
recognise a repeat submitter by address.

---

## Step 3, Deploy the API (Render)

1. Sign up at <https://render.com> with GitHub. No card.
2. **New → Web Service** → connect your `ai-safety-atlas` repository.
3. Settings:

   | Field | Value |
   |---|---|
   | Name | `ai-safety-atlas-api` |
   | Language / Runtime | **Docker** |
   | Dockerfile Path | `backend/Dockerfile` |
   | Docker Build Context Directory | `backend` |
   | Branch | `main` |
   | Instance Type | **Free** |
   | Region | Frankfurt (match your Neon region) |

4. **Environment variables**, add these before the first deploy:

   | Key | Value |
   |---|---|
   | `DATABASE_URL` | the pooled Neon string from Step 1 |
   | `ALLOWED_ORIGIN` | `https://example.pages.dev`, a placeholder for now; you will fix it in Step 5 |
   | `ADMIN_TOKEN` | the first random hex from Step 2 |
   | `ITERATION` | `0` |

   Optional: `TURNSTILE_SECRET` (Step 6), `IP_HASH_SALT` (Step 2). Do **not** set
   `PORT`, Render sets it.

   `ALLOWED_ORIGIN` must be a full origin with scheme and no trailing path. The
   service refuses to start on `*` or on a bare hostname, on purpose: it is the only
   thing standing between the write endpoint and every other website.

5. **Create Web Service** and watch the log. You want to see:

   ```
   {"level":"INFO","msg":"migration applied","version":"0001_init.sql"}
   {"level":"INFO","msg":"listening","port":"10000", ...}
   ```

   If it exits immediately, the log line says which variable is wrong.

6. Copy the service URL (`https://ai-safety-atlas-api.onrender.com`) and test it:

   ```bash
   curl https://ai-safety-atlas-api.onrender.com/api/health
   # {"status":"ok"}
   ```

   The **first** request after an idle period takes up to ~50 seconds, the instance
   is asleep and has to boot. That is the free tier working as intended, and the
   frontend is built for it (it shows "waking the server…" and retries once).

---

## Step 4, Deploy the site (Cloudflare Pages)

1. Sign up at <https://dash.cloudflare.com>. No card.
2. **Workers & Pages → Create → Pages → Connect to Git**, pick the repository.
3. Build settings:

   | Field | Value |
   |---|---|
   | Framework preset | **None** (or Vue) |
   | Build command | `npm install && npm run build` |
   | Build output directory | `frontend/dist` |
   | Root directory | `frontend` |

4. **Environment variables (Production)**:

   | Key | Value |
   |---|---|
   | `VITE_API_BASE` | your Render URL, e.g. `https://ai-safety-atlas-api.onrender.com`, no trailing slash |
   | `VITE_CONTACT_EMAIL` | the address for GDPR access/erasure requests, shown on `/privacy` |
   | `NODE_VERSION` | `20` |

   Optional: `VITE_TURNSTILE_SITE_KEY` (Step 6).

5. **Save and Deploy.** Copy the resulting URL, e.g.
   `https://ai-safety-atlas.pages.dev`.

The build copies `data/atlas.json` into the site (so `/data/atlas.json` is
downloadable) and bundles it into the app. `frontend/public/_redirects` already
handles SPA routing, so `/agenda/DS1` works on a hard refresh.

---

## Step 5, Close the loop and test end to end

1. Back in **Render → your service → Environment**, set `ALLOWED_ORIGIN` to the exact
   Pages URL from Step 4 (`https://ai-safety-atlas.pages.dev`, no trailing slash) and
   save. Render redeploys automatically.

2. Open `frontend/public/_headers` and set the `connect-src` entry of the
   Content-Security-Policy to your API origin instead of `https://*.onrender.com` if
   you want it tight. Commit and push; Pages redeploys.

3. Now walk the site as a researcher would:

   - Open the Pages URL. The grid renders immediately (it does not wait for the API).
   - Click a cell → the drill-down opens and lists that area's agendas.
   - Open any agenda → **Your reading of …** → pick a familiarity level → **Send
     feedback**. The first submission may sit on "waking the server…" for up to a
     minute.
   - The hero counter ("Researcher responses") should go up. Reload to confirm it
     came from the server rather than the optimistic local bump.

4. Confirm the row landed:

   ```bash
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
        https://ai-safety-atlas-api.onrender.com/api/admin/export?format=json
   ```

   or in the Neon SQL editor:

   ```sql
   select id, agenda_id, familiarity, is_anonymous, created_at from feedback order by id desc limit 10;
   ```

If the submission fails with a CORS error in the browser console, `ALLOWED_ORIGIN`
does not match the Pages URL exactly, check scheme, subdomain and trailing slash.

---

## Step 6, Optional: turn on Turnstile

The honeypot and the per-IP rate limit (5 writes/minute, 30/hour) are on by default
and are probably enough. If you start seeing junk:

1. Cloudflare dashboard → **Turnstile** → **Add site**. Free.
2. Add the **site key** as `VITE_TURNSTILE_SITE_KEY` in Cloudflare Pages.
3. Add the **secret key** as `TURNSTILE_SECRET` in Render.
4. Redeploy both. Set both or neither: a site key without a server secret renders a
   widget nobody checks, and a secret without a site key rejects every submission.

---

## Reading the feedback

There is **no admin screen on the website**, that keeps the public site read-only
and its attack surface tiny. You read submissions in one of two places, both of
which you already have: the **admin export endpoint** (bearer token) or the **Neon
SQL console**. (An internal dashboard could be added later; it is deliberately not
part of the public app.)

Each row targets **either an agenda** (`agenda_id` set) **or a whole Area × Problem
cell** (`area_tag` + `problem_id` set, `agenda_id` null). Cell feedback carries no
`agree_with_rating`, it is the free-text "what did we get wrong" plus familiarity.

**Everything, with names and emails**, admin export:

```bash
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
     "https://ai-safety-atlas-api.onrender.com/api/admin/export?format=json" > feedback.json
```

**In SQL**, Neon console:

```sql
-- who said what, most recent first (agenda rows and cell rows together)
select coalesce(agenda_id, area_tag || ' × ' || problem_id) as about,
       familiarity, agree_with_rating, notes,
       case when is_anonymous then '(anonymous)' else coalesce(submitter_name, '(no name)') end as who,
       created_at
from feedback
order by created_at desc;

-- per-agenda: where researchers disagree with us most
select agenda_id,
       count(*) filter (where agree_with_rating = 'disagree') as disagree,
       count(*) as responses,
       round(avg(familiarity)::numeric, 1) as avg_familiarity
from feedback
where agenda_id is not null
group by agenda_id
having count(*) filter (where agree_with_rating = 'disagree') > 0
order by disagree desc, avg_familiarity desc;

-- cell-level feedback (comments on a whole area × problem)
select area_tag, problem_id, familiarity, notes, created_at
from feedback
where area_tag is not null
order by created_at desc;

-- who consented to being contacted
select distinct submitter_name, submitter_email
from feedback
where contact_consent and submitter_email is not null;
```

Two things to respect, because the site promised them:

- **`reuse_consent = false` means do not quote it**, attributed or not.
- **Anonymous rows cannot be traced back** to a person. That is the trade the
  contributor made, and there is no lookup that undoes it.

---

## Cutting Iteration 1

1. Edit `data/atlas.json`. Bump `meta.iteration` to `1`, update `meta.iteration_label`
   (it is displayed in the header and the footer) and `meta.source_check_date`.
   If you keep the source workbooks, put them in `data/source/` and run
   `python scripts/build_atlas.py` (dry run) then `--write`.
2. `python scripts/validate_atlas.py`, it must print `OK`.
3. `python scripts/gen_agenda_ids.py`, then
   `cd backend && gofmt -w internal/atlas/ids.go && go test ./...`.
4. Set `ITERATION=1` on Render (so new submissions are stamped with it) and save.
5. `git push`. Cloudflare Pages rebuilds the site by itself.

Existing feedback keeps the iteration number it was given under, and the public
counters keep counting everything, a contributor's answer does not stop existing
because you re-cut the map.

---

## What to expect from the free tiers

- **The API sleeps.** After ~15 minutes idle, the next request takes up to ~50
  seconds. The map does not wait for it; only the counters and the feedback form do,
  and both say so on screen. If you are about to share the link somewhere, hit
  `/api/health` first to warm it.
- **Neon scales to zero** too, which adds a second or so to the first query.
- **Neither costs anything**, and neither expires.
- Nothing on the site tracks readers: no analytics, no cookies, no third-party
  scripts unless you enable Turnstile.

---

## The Swiss-cheese view is switched off, and that is a decision

`swiss_layers.json` ships empty and the section is hidden. It is not an unfinished
feature, it is a feature that would be dishonest to ship on this data.

A Swiss-cheese diagram makes three claims: that defences form an **order**, that they
are **roughly independent**, and that each has **holes** against a hazard. Only the
third is in the dataset (a weak tier is a big hole, derived deterministically). The
ordering and the independence are research judgments the workbook does not contain,
and the map's own reading note argues the defences are **correlated**: the
load-bearing problems (P1, P6) all lean on measurement and oversight, and the single
Robust cell in the whole matrix (sandboxing against P7) is the least AI-specific
agenda in it, inheriting its reliability from ordinary systems security. A diagram
implying independence would say the opposite of what the data says.

To turn it on you would need to author, and be prepared to defend:

1. an ordered list of defensive layers, outermost first;
2. for each, which areas or agendas belong to it;
3. a rationale per layer for *why it sits there in the order*;
4. honestly, something about how correlated the layers are, because the residual-risk
   number the view prints multiplies the holes as if they were independent, and it
   says on screen that this is an illustration of the assumption, not an estimate.

Then set `enabled: true` in `swiss_layers.json` and redeploy the frontend. The
permanent caveat stays on screen; it is not configurable, by design.

---

## When you buy a domain later

Nothing structural changes. In **Cloudflare Pages → your project → Custom domains**,
add the domain and follow the DNS instructions (if the domain is registered with
Cloudflare this is two clicks). Then update exactly two settings: `ALLOWED_ORIGIN`
on Render to the new origin, and, if you tightened it, the `connect-src` line in
`frontend/public/_headers`. `VITE_API_BASE` only changes if you also move the API to
a subdomain of your own. Redeploy both, and re-run the end-to-end test in Step 5.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Render service exits at boot | a required env var is missing or malformed | the log line names it, usually `ALLOWED_ORIGIN` without a scheme |
| Submission fails, console shows CORS | `ALLOWED_ORIGIN` ≠ the site's origin | copy the Pages URL exactly, no trailing slash, then redeploy |
| Counters show "not available right now" | API asleep or unreachable | expected on first load; hit `/api/health` and reload |
| `/api/health` returns `degraded` | `DATABASE_URL` unset or Neon unreachable | check the variable, and that the Neon project is not deleted |
| Export returns 401 | wrong or missing token | header must be `Authorization: Bearer <ADMIN_TOKEN>` |
| Export returns 401 with the right token | `ADMIN_TOKEN` is empty on Render | an empty token disables the endpoint entirely |
| Feedback returns 429 | rate limit (5/min, 30/hour per IP) | wait a minute; it is per-IP and in-memory |
| `/agenda/DS1` 404s on refresh | SPA fallback missing | `frontend/public/_redirects` must be in the build output |
