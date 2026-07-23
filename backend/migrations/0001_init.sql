-- 0001_init.sql, AI Safety Agendas x Problems Map
--
-- The curated map is static and ships with the frontend. The database holds
-- exactly one table: researcher-submitted feedback.
--
-- This file is idempotent on purpose. You can either paste it into the Neon SQL
-- editor or let the backend apply it on boot; running it twice is harmless.

create table if not exists feedback (
  id                bigserial   primary key,
  agenda_id         text        not null,
  familiarity       smallint    not null check (familiarity between 0 and 3),
  agree_with_rating text        check (agree_with_rating in ('agree','disagree','unsure')),
  notes             text,
  submitter_name    text,
  submitter_email   text,
  is_anonymous      boolean     not null default true,
  contact_consent   boolean     not null default false,
  reuse_consent     boolean     not null default false,
  submitter_token   uuid        not null,
  iteration         int         not null default 0,
  ip_hash           text,
  user_agent        text,
  created_at        timestamptz not null default now()
);

create index if not exists feedback_agenda_id_idx       on feedback (agenda_id);
create index if not exists feedback_submitter_token_idx on feedback (submitter_token);
