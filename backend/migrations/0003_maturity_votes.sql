-- 0003_maturity_votes.sql
--
-- Feedback is now a vote, not an opinion about our vote.
--
-- The agree/disagree question is gone: asking a researcher whether they agree
-- with a rating buys one bit and anchors them on ours. Instead every submission
-- carries the reader's own maturity rating for the research area, in the same six
-- tiers the map itself uses, and may carry a rating for individual agendas
-- (subareas) within that area.
--
-- Maturity here is a property of the research, not of a problem: how much work
-- exists, how well it has been done, what it has actually shown. That is why the
-- vote is scoped to an area (and optionally to its agendas) rather than to an
-- Area x Problem cell. The community grid is built by substituting these ratings
-- into the curated adjacency, which agendas address which problem, which stays
-- the authors' claim and is not voted on.
--
-- agree_with_rating is kept, not dropped: rows written before this migration are
-- real answers to a question that was really asked, and deleting them would be
-- rewriting the record. Nothing writes it any more.
--
-- Idempotent, like 0001 and 0002: safe to paste into the Neon console or let the
-- backend apply it on boot, and safe to run twice.

alter table feedback add column if not exists area_maturity    text;
alter table feedback add column if not exists subarea_maturity jsonb;

-- The six tiers of legend.tier_order in data/atlas.json. Spelled out rather than
-- referenced so the database rejects a typo even if the API ever stops checking.
-- If a future iteration renames a tier, this constraint is part of that change.
do $$
begin
  if not exists (select 1 from pg_constraint where conname = 'feedback_area_maturity_tier') then
    alter table feedback add constraint feedback_area_maturity_tier
      check (area_maturity is null or area_maturity in (
        'Robust (small scale)',
        'Strong existence proof',
        'Early / partial',
        'Untested',
        'Never demonstrated',
        'Contested'
      ));
  end if;
end $$;

-- subarea_maturity is an object of agenda id -> tier, e.g. {"IN6":"Untested"}.
-- The API validates the keys against the curated agenda list and the values
-- against the tier list; this only enforces the shape.
do $$
begin
  if not exists (select 1 from pg_constraint where conname = 'feedback_subarea_maturity_object') then
    alter table feedback add constraint feedback_subarea_maturity_object
      check (subarea_maturity is null or jsonb_typeof(subarea_maturity) = 'object');
  end if;
end $$;

-- Every new row names the area it votes on, including rows submitted from an
-- agenda page (the area is derived from the agenda id server-side). The area
-- aggregate groups on this column.
create index if not exists feedback_area_tag_idx on feedback (area_tag);

-- The per-agenda aggregate expands subarea_maturity with jsonb_each_text; the
-- containment index makes the "has a vote for this agenda" scan cheap once the
-- table is large enough for it to matter.
create index if not exists feedback_subarea_maturity_idx on feedback using gin (subarea_maturity);
