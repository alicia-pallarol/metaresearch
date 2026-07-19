#!/usr/bin/env python3
"""No-Docker local dev: start a userspace Postgres and run the Go backend.

For contributors who don't have Docker. It boots a throwaway Postgres inside
this repo (no root, no system install), applies migrations automatically via
the backend, and runs the API on http://localhost:8080.

    python3 scripts/dev-stack.py

Then, in a second terminal:

    cd frontend && npm install && npm run dev

Stop with Ctrl-C. The database persists in ./.localpg between runs; delete that
folder to reset.
"""
import os
import pathlib
import signal
import subprocess
import sys
import time

REPO = pathlib.Path(__file__).resolve().parent.parent
PGDATA = REPO / ".localpg"


def ensure_pgserver():
    try:
        import pgserver  # noqa: F401
    except ImportError:
        print("Installing 'pgserver' (bundled Postgres, userspace, no root)…")
        subprocess.run([sys.executable, "-m", "pip", "install", "--quiet", "pgserver"], check=True)
    import pgserver
    return pgserver


def main():
    pgserver = ensure_pgserver()
    PGDATA.mkdir(exist_ok=True)
    print(f"Starting Postgres in {PGDATA} …")
    db = pgserver.get_server(PGDATA)

    # Create the application database once.
    exists = db.psql("SELECT 1 FROM pg_database WHERE datname='saige'").strip()
    if "1" not in exists:
        db.psql("CREATE DATABASE saige")
        print("Created database 'saige'.")

    sock = str(PGDATA.resolve())
    database_url = f"postgresql://postgres:@/saige?host={sock}"

    env = dict(
        os.environ,
        DATABASE_URL=database_url,
        ALLOWED_ORIGIN=os.environ.get("ALLOWED_ORIGIN", "http://localhost:5173"),
        PORT=os.environ.get("PORT", "8080"),
        IP_HASH_SALT=os.environ.get("IP_HASH_SALT", "local-dev-salt-not-for-production"),
        MIGRATIONS_PATH=str(REPO / "migrations"),
        GO111MODULE="on",
    )

    print("Starting Go backend on http://localhost:%s …" % env["PORT"])
    print("  DATABASE_URL=%s" % database_url)
    print("  (migrations run automatically on startup)\n")

    backend = subprocess.Popen(["go", "run", "."], cwd=str(REPO / "backend"), env=env)

    def shutdown(*_):
        print("\nShutting down…")
        backend.terminate()
        try:
            backend.wait(timeout=10)
        except subprocess.TimeoutExpired:
            backend.kill()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)
    backend.wait()


if __name__ == "__main__":
    main()
