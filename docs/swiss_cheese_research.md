# A Swiss Cheese Defense Model for AI Safety Research Areas

> **Disclaimer: AI-generated content.** This document was drafted with AI assistance as a working conceptual mapping. It has not been peer-reviewed. It is presented on the platform with a visible label to that effect.

This file is the source for the conceptual mapping rendered on the platform's
Swiss cheese page. The machine-readable version consumed by the frontend lives at
`frontend/src/data/swissCheese.js`; keep the two in sync when revising.

## 1. Background

The Swiss cheese model (James Reason, 1990) pictures safety as a stack of
defensive layers. Each layer is imperfect — it has holes. Harm occurs when holes
in successive layers align. Implications: no single layer must be perfect if the
layers are diverse and their holes do not align; reducing risk means both adding
layers and shrinking holes; correlated weaknesses across layers are the most
dangerous failure mode.

Applied to AI safety, the hazard is a misaligned, misused, or unsafe AI system
causing harm; the layers are stages at which the field can catch or prevent that
harm. Community *familiarity* with each research area is used as a proxy for how
much attention and expertise reinforces each layer.

## 2. Proposed defense layers and mapping to research areas

- **Layer 1 — Alignment by construction** (before/during training): EA, AF.
- **Layer 2 — Transparency and verification**: IN, SA, CE.
- **Layer 3 — Oversight and control**: SO, AR.
- **Layer 4 — Deployment and security**: DS, SH.
- **Layer 5 — Societal and institutional**: ST, TG, SM (cross-cutting).

Each layer's hazard framing and typical holes are transcribed verbatim in
`frontend/src/data/swissCheese.js`.

## 3. Interpreting the visualization

Slice shading is proportional to aggregate community familiarity with the areas in
that layer. Holes are illustrative, not exhaustive. Familiarity is a proxy for
attention, not a measure of a layer's actual effectiveness.

## 4. Sources for further reading (real, verifiable)

- Reason, J. (1990). *Human Error*. Cambridge University Press.
- Reason, J. (2000). "Human error: models and management." *BMJ* 320:768–770.
- Hendrycks, D. et al. (2023). "An Overview of Catastrophic AI Risks."
- Defense-in-depth discussions in frontier-lab safety frameworks.

*(The layer groupings in §2 are this document's own construction and should be
revised by the research team.)*
