---
name: no-em-dashes
description: Never use em dashes in any text for this project; use commas, colons, semicolons, or periods
metadata:
  type: feedback
---

The user wants **no em dashes (—) anywhere** in this project's text, established 2026-07-23. The whole repo was swept to remove them (243 occurrences across 40 files).

**Why:** style preference for this platform's writing (em dashes read as an AI tell / are unwanted).

**How to apply:** In any prose I write or edit here (UI copy, atlas.json content, docs, code comments, commit messages), substitute the most appropriate punctuation for the context, not another dash:
- parenthetical/appositive aside → commas, or parentheses;
- clause join of two independent clauses → semicolon or period (a comma there is a splice, avoid);
- "statement, then its explanation" → colon;
- browser-title / label separators → middle dot "·" (the app already uses "·").
Do not reintroduce em dashes in new text. En dashes and spaced hyphens as a dash substitute are also not wanted.
