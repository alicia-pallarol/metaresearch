<script setup lang="ts">
/**
 * The maturity legend. Always reachable: the heatmaps are unreadable without it,
 * and the tiers are the part readers are most likely to misread (Contested is a
 * stronger claim than Untested, not a weaker one).
 */
import { atlas, tierCounts, tierOrder } from '@/lib/atlas'
import { tierAbbrev, tierStyle } from '@/lib/scales'

defineProps<{ compact?: boolean }>()

const counts = tierCounts()
const blankMeaning = atlas.legend.tiers['(blank)'] ?? 'Not applicable.'
</script>

<template>
  <div class="legend" :class="{ 'legend--compact': compact }">
    <h3 class="legend__title">Evidence maturity</h3>
    <p v-if="!compact" class="legend__note">
      How mature the evidence is that an agenda addresses <em>that specific problem</em>, not how
      important the agenda is, and not how good the research area is overall.
    </p>

    <ul class="legend__list">
      <li v-for="tier in tierOrder" :key="tier" class="legend__item">
        <span
          class="legend__swatch"
          :class="{ hatched: tierStyle(tier).hatched }"
          :style="{ backgroundColor: tierStyle(tier).fill, color: tierStyle(tier).ink }"
          aria-hidden="true"
          >{{ tierAbbrev(tier) }}</span
        >
        <span class="legend__body">
          <span class="legend__name">
            {{ tier }}
            <span class="legend__count tabular">{{ counts[tier] ?? 0 }} cells</span>
          </span>
          <span v-if="!compact" class="legend__def">{{ atlas.legend.tiers[tier] }}</span>
        </span>
      </li>

      <li class="legend__item">
        <span class="legend__swatch legend__swatch--blank" aria-hidden="true"></span>
        <span class="legend__body">
          <span class="legend__name">Blank</span>
          <span v-if="!compact" class="legend__def">{{ blankMeaning }}</span>
        </span>
      </li>
    </ul>

    <p v-if="!compact" class="legend__note legend__note--last">
      The two hatched tiers are not rungs on the evidence ladder. <strong>Never demonstrated</strong>
      is a negative finding, looked for and not found. <strong>Contested</strong> means the evidence
      cuts against the agenda or a real dispute is live in the literature.
    </p>
  </div>
</template>

<style scoped>
.legend {
  background: var(--surface);
  border: 1px solid var(--rule);
  border-radius: var(--radius-lg);
  padding: 1.1rem 1.25rem;
}

.legend__title {
  font-size: 0.9375rem;
  margin-bottom: 0.35rem;
}

.legend__note {
  color: var(--ink-secondary);
  font-size: 0.8125rem;
  margin-bottom: 0.9rem;
}

.legend__note--last {
  margin: 0.9rem 0 0;
}

.legend__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.55rem;
}

.legend--compact .legend__list {
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
}

.legend__item {
  display: flex;
  gap: 0.6rem;
  align-items: flex-start;
}

.legend__swatch {
  flex: none;
  width: 2rem;
  height: 1.35rem;
  border-radius: 3px;
  display: grid;
  place-items: center;
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.legend__swatch--blank {
  border: 1px dashed var(--rule-strong);
  background: repeating-linear-gradient(45deg, var(--surface-sunken) 0 3px, transparent 3px 6px);
}

.legend__body {
  display: block;
  min-width: 0;
}

.legend__name {
  display: block;
  font-size: 0.875rem;
  font-weight: 600;
}

.legend__count {
  font-weight: 400;
  color: var(--ink-muted);
  font-size: 0.75rem;
  margin-left: 0.35rem;
}

.legend__def {
  display: block;
  color: var(--ink-secondary);
  font-size: 0.8125rem;
  line-height: 1.45;
}
</style>
