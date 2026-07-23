---
name: focus-areas
description: The "Focus areas" section is computed from the grid by lib/focus.ts, not hardcoded
metadata:
  type: project
---

The homepage "Focus areas" section (was the hardcoded "priority tiers") is computed, decided 2026-07-23:

- **Two independent passes**, one after the other: research areas, then problems. Each sorts its units into three categories, **Easy to intervene** (best result strong but unfinished), **Bottleneck** (much depends on it and it's still weak, for problems = agenda count, for areas = how shared its problems are), **Underexplored** (thin: few agendas, many blanks, low maturity, not merely contested).
- **Source is synced with the grid's Research/Community switch** (shared `source` ref in HomeView, `v-model:source` on both `AreaHeatmap` and `FocusAreas`). Research → curated ratings; community → vote averages. Community only classifies units that have votes; near-Robust units are excluded (closest to solved).
- **All the logic + tunable constants live in [frontend/src/lib/focus.ts]** with heavy comments, the user explicitly wants to read/tweak thresholds there. Each lens is min-max normalised against a **fixed research-population reference** (not the survivor cohort) so a single voted unit doesn't collapse to 0. Per-item reason strings are assembled from the unit's own numbers (no runtime LLM, hard project rule).
- Validation: on iteration-0 data the classifier independently reproduces the curated editorial read (P1 "the hub" → bottleneck; AF/SO oversight+foundations → bottleneck areas; P3 sharp-left-turn → underexplored). Pinned in [frontend/src/lib/focus.test.ts].
- The curated `priorities` block stays in atlas.json (validated by atlas.test.ts) but is no longer rendered.

Builds on [[maturity-vote-model]], the community grid votes are what this reads in community mode.
