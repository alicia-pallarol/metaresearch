---
name: maturity-vote-model
description: How community feedback maps onto the grid, maturity is about the research, not the problem
metadata:
  type: project
---

The operator's mental model for the map, established 2026-07-22, load-bearing for the feedback/community-grid design:

- **Maturity is a property of the research itself**, how much work exists, how well it was done, what it has actually shown, **not** of how much of a problem it solves. The curated Area×Problem tiers are "only a starting place"; an agenda can be mature research and still only chip at a hard problem.
- Therefore community feedback is a **maturity vote scoped to a research area** (required) plus optional per-agenda (subarea) votes, never scoped to a single Area×Problem cell.
- The **community view of the grid keeps the curated adjacency** (which agendas address which problem) and substitutes community maturity votes into it. The community does not vote on the structure, only the maturity.
- Best/typical/worst-case = quantile **bands** over the tiers behind a cell, summarised by **median, never mean** (two tiers, Contested, Never demonstrated, are off the evidence axis). Below 8 ratings a band is a single rating, so **research + best case reproduces the historical "best of N agendas" grid exactly**, this invariant is pinned in [frontend/src/lib/community.test.ts].

The agree/disagree question was deliberately retired (it "bought one bit and anchored the reader on our rating"). See [feedback-vote-not-verdict].
