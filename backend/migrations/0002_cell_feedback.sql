-- 0002_cell_feedback.sql
--
-- Feedback can now be about a whole Area x Problem cell (e.g. "Interpretability's
-- coverage of P4"), not only a single agenda. A row targets either an agenda, or an
-- area + problem cell.
--
-- Idempotent, like 0001: safe to paste into the Neon console or let the backend
-- apply it on boot, and safe to run twice.

alter table feedback alter column agenda_id drop not null;

alter table feedback add column if not exists area_tag   text;
alter table feedback add column if not exists problem_id  text;

create index if not exists feedback_problem_id_idx on feedback (problem_id);

-- Every row must point at something: an agenda, or an area+problem cell. This is
-- defence in depth behind the API's own validation.
do $$
begin
  if not exists (select 1 from pg_constraint where conname = 'feedback_has_target') then
    alter table feedback add constraint feedback_has_target
      check (agenda_id is not null or (area_tag is not null and problem_id is not null));
  end if;
end $$;
