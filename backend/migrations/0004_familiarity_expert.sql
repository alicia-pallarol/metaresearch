-- 0004_familiarity_expert.sql
--
-- Familiarity gains a fifth level: 4 = expert, "I work directly in this area".
-- The scale is now 0..4, so the range check written in 0001 has to widen.
--
-- 0001 declared the check inline, so it has an auto-generated name. Rather than
-- guess that name, drop whichever check constraint currently bounds familiarity
-- and add a named one in its place. Running this twice is harmless: the second
-- pass drops the named constraint and re-adds an identical one.

do $$
declare c text;
begin
  for c in
    select conname
    from pg_constraint
    where conrelid = 'feedback'::regclass
      and contype = 'c'
      and pg_get_constraintdef(oid) ilike '%familiarity%'
  loop
    execute format('alter table feedback drop constraint %I', c);
  end loop;

  alter table feedback
    add constraint feedback_familiarity_range check (familiarity between 0 and 4);
end $$;
