# Deployment guide

A complete, ordered checklist to take this platform from source code to a live
site on free tiers. It assumes you are a competent developer but have **never
used Neon, Render, or Cloudflare Pages**. Every step you must perform by hand is
listed. Expect ~30–45 minutes end to end.

**Architecture recap**

```
Cloudflare Pages (Vue SPA)  ──HTTPS──▶  Render (Go API)  ──▶  Neon (Postgres)
        your-project.pages.dev            saige-api.onrender.com
```

You will end up with three URLs. Two of them refer to each other, so you set them
in this order: **Neon → Render → Cloudflare**, then come back and connect
Cloudflare's URL to Render's CORS setting.

---

## 0. Prerequisites

- [ ] The repository is pushed to a GitHub (or GitLab) repo you control. Render and
      Cloudflare both deploy by connecting to that repo.
- [ ] Generate a rate-limit salt now and keep it handy:
      ```bash
      openssl rand -hex 32
      ```
      Call this value `IP_HASH_SALT` below.

---

## 1. Neon — create the Postgres database

1. [ ] Go to <https://neon.tech> and sign up (GitHub login is easiest). Free tier
       requires no card.
2. [ ] Click **Create project**. Pick a name (e.g. `saige`), keep the default
       Postgres version, and choose the **region closest to where Render runs**
       (Render free defaults to Oregon, US West — pick a US West/US East Neon
       region to keep latency low).
3. [ ] After the project is created, open **Dashboard → Connection Details**.
4. [ ] Select **Connection string**, ensure **Pooled connection** is on, and copy
       the string. It looks like:
       ```
       postgres://USER:PASSWORD@ep-xxxx-pooler.REGION.aws.neon.tech/neondb?sslmode=require
       ```
       This is your **`DATABASE_URL`**. Keep it secret.

> Migrations are applied **automatically** by the backend on startup, so you do not
> need to run anything against Neon by hand. (Manual fallback is in §6.)

---

## 2. Render — deploy the Go backend

Render clones the whole repo, builds the Docker image (which bundles
`/migrations`), and runs the server.

1. [ ] Go to <https://render.com> and sign up (GitHub login). Free tier, no card.
2. [ ] **New → Web Service**, then **Connect** your GitHub repository (authorize
       Render to read it if prompted).
3. [ ] Configure the service:
       - **Name:** `saige-api` (this becomes `https://saige-api.onrender.com`)
       - **Region:** the one nearest your Neon region
       - **Branch:** `main` (or your default)
       - **Root Directory:** leave **blank** (the Dockerfile lives at the repo root)
       - **Runtime / Language:** **Docker** (Render auto-detects the `Dockerfile`)
       - **Instance Type:** **Free**
       - No build/start command is needed — the Dockerfile defines them.

   > **Not using Docker?** If you prefer Render's native Go runtime instead of the
   > Dockerfile: set **Root Directory** to `backend`, **Build Command** to
   > `go build -o server .`, **Start Command** to `./server`, and add an env var
   > `MIGRATIONS_PATH=../migrations` so the server can find the SQL files.

4. [ ] Under **Environment**, add these variables (**Add Environment Variable**):
       | Key | Value |
       |-----|-------|
       | `DATABASE_URL` | the Neon pooled connection string from §1 |
       | `IP_HASH_SALT` | the `openssl rand -hex 32` value from §0 |
       | `ALLOWED_ORIGIN` | `http://localhost:5173` **for now** — you'll update it in §4 once you have the Pages URL |
       - Do **not** set `PORT`; Render injects it and the app reads it.
5. [ ] Click **Create Web Service**. Watch the deploy log. On success you'll see
       `migrations applied` and `listening on :PORT`.
6. [ ] Verify: open `https://saige-api.onrender.com/api/health` — it should return
       `{"status":"ok"}`. (The first hit after inactivity may take 30–60s because
       the free instance sleeps; the frontend handles this gracefully.)

Copy your backend URL (e.g. `https://saige-api.onrender.com`) — you need it next.

---

## 3. Cloudflare Pages — deploy the frontend

1. [ ] Go to <https://dash.cloudflare.com> and sign up (free, no card).
2. [ ] In the left sidebar: **Workers & Pages → Create → Pages →
       Connect to Git**. Authorize Cloudflare and select your repository.
3. [ ] Configure the build:
       - **Project name:** e.g. `saige` (this becomes `https://saige.pages.dev`)
       - **Production branch:** `main`
       - **Framework preset:** `Vue` (or `None`)
       - **Build command:** `npm run build`
       - **Build output directory:** `dist`
       - **Root directory (Advanced → set):** `frontend`
4. [ ] Under **Environment variables (Build)**, add:
       | Key | Value |
       |-----|-------|
       | `VITE_API_BASE_URL` | your Render URL from §2, e.g. `https://saige-api.onrender.com` |
       - `NODE_VERSION` = `18` (or higher) if the build complains about Node.
5. [ ] Click **Save and Deploy**. When it finishes, note your Pages URL, e.g.
       `https://saige.pages.dev`.

The `_headers` and `_redirects` files in `frontend/public/` are picked up
automatically — they set security headers and the SPA fallback for client-side
routes.

---

## 4. Connect the two — set CORS

The backend only accepts browser requests from origins in `ALLOWED_ORIGIN`. Now
that you have the Pages URL, point the backend at it.

1. [ ] Back in **Render → your service → Environment**, edit `ALLOWED_ORIGIN` to
       your exact Pages origin, no trailing slash:
       ```
       https://saige.pages.dev
       ```
       (You may list several comma-separated origins, e.g. to also allow a preview
       domain: `https://saige.pages.dev,http://localhost:5173`.)
2. [ ] Save. Render redeploys automatically. Wait for it to go live.

---

## 5. Verify the deployment

1. [ ] Open your Pages URL. The framework table should load (the first load may
       show a brief "Waking the server…" message while Render's free instance
       spins up — this is expected).
2. [ ] Click **Submit your familiarity**, fill the form, and submit. You should see
       a success message and the heatmap should update (`n = 1`).
3. [ ] Submit again with the **same email** — you should see "your previous response
       was updated", and `n` should stay the same (deduplication works).
4. [ ] Open the **Defense model** page; confirm the layers shade according to
       familiarity and the AI-assisted disclaimer is visible.
5. [ ] In the browser devtools **Network** tab, confirm requests to
       `…onrender.com/api/...` succeed with no CORS errors. If you see CORS errors,
       re-check that `ALLOWED_ORIGIN` exactly matches the Pages origin (scheme +
       host, no trailing slash) and that Render finished redeploying.

---

## 6. Migrations — how they run (and the manual fallback)

- **Automatic (default):** the backend runs all pending `*.up.sql` files from
  `/migrations` on every startup, tracked in a `schema_migrations` table. Re-runs
  are no-ops. You normally never touch this.
- **Manual fallback**, if you ever want to apply migrations yourself (e.g. with
  [`golang-migrate`](https://github.com/golang-migrate/migrate)):
  ```bash
  migrate -path ./migrations \
    -database "$DATABASE_URL" up
  ```
  Or apply the SQL files in filename order with `psql`:
  ```bash
  psql "$DATABASE_URL" -f migrations/0001_init.up.sql
  psql "$DATABASE_URL" -f migrations/0002_seed_research_areas.up.sql
  ```

---

## 7. Later: attaching a custom domain

No code changes are ever required — everything is env-driven. When you have a
domain (say `example.org` for the site and `api.example.org` for the backend):

**Frontend (Cloudflare Pages):**
1. [ ] **Workers & Pages → your project → Custom domains → Set up a custom domain.**
2. [ ] Enter `example.org` (or `www.example.org`). If the domain's DNS is on
       Cloudflare, records are added automatically; otherwise add the shown CNAME
       at your registrar. TLS is provisioned automatically.

**Backend (Render):**
3. [ ] **Render → your service → Settings → Custom Domains → Add Custom Domain**,
       enter `api.example.org`, and add the shown CNAME record at your DNS provider.
       Render provisions TLS automatically.

**Reconnect the two env vars:**
4. [ ] **Render → Environment:** set `ALLOWED_ORIGIN` to `https://example.org`
       (add `https://www.example.org` too if you use www). Save → redeploy.
5. [ ] **Cloudflare Pages → Settings → Environment variables:** set
       `VITE_API_BASE_URL` to `https://api.example.org`. Then **Deployments →
       Retry/Redeploy** so the new value is baked into the build.
6. [ ] Re-run the §5 verification against the new domain.

That's it — the same code now serves the custom domain with correct CORS.

---

## Troubleshooting

| Symptom | Likely cause / fix |
|---------|--------------------|
| Frontend stuck on "Waking the server…" for >1 min | Render free instance cold start or a bad `DATABASE_URL`. Check Render logs; hit `/api/health` directly. |
| CORS error in console | `ALLOWED_ORIGIN` doesn't exactly match the Pages origin (scheme/host, no trailing slash), or Render hasn't finished redeploying. |
| `IP_HASH_SALT is required` in Render logs | The env var wasn't set. Add it and redeploy. |
| Submit returns 429 | Rate limit hit (default 5/hour per IP hash). Adjust `RATE_LIMIT_PER_HOUR` if needed. |
| Build fails on Cloudflare with Node error | Set `NODE_VERSION=18` (or higher) in Pages build env vars. |
| Client-side route 404s on refresh | Ensure `frontend/public/_redirects` deployed (SPA fallback). |
