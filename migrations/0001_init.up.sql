-- Schema for the AI Safety Research Familiarity platform.

CREATE TABLE IF NOT EXISTS research_areas (
    id            SERIAL PRIMARY KEY,
    problem_group TEXT    NOT NULL,
    tag           TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL,
    definition    TEXT    NOT NULL,
    sort_order    INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS submissions (
    id              BIGSERIAL PRIMARY KEY,
    -- Email is required (used for deduplication and, only with consent, contact).
    -- It is unique so a resubmission from the same person replaces the previous one.
    -- Emails are personal data and are NEVER exposed through any public API response.
    email           TEXT        NOT NULL UNIQUE,
    -- Name is only shown publicly if the submission is non-anonymous AND a
    -- contributor list feature is enabled (disabled by default in the API).
    name            TEXT        NOT NULL DEFAULT '',
    anonymous       BOOLEAN     NOT NULL DEFAULT TRUE,
    contact_consent BOOLEAN     NOT NULL DEFAULT FALSE,
    -- Salted hash of the submitter IP; used only for rate limiting. Never the raw IP.
    ip_hash         TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS submissions_ip_hash_created_idx
    ON submissions (ip_hash, created_at);

CREATE TABLE IF NOT EXISTS ratings (
    submission_id    BIGINT   NOT NULL REFERENCES submissions (id) ON DELETE CASCADE,
    research_area_id INTEGER  NOT NULL REFERENCES research_areas (id) ON DELETE CASCADE,
    familiarity      SMALLINT NOT NULL CHECK (familiarity BETWEEN 0 AND 3),
    PRIMARY KEY (submission_id, research_area_id)
);

CREATE INDEX IF NOT EXISTS ratings_area_idx ON ratings (research_area_id);
