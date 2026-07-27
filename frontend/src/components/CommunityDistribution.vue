<script setup lang="ts">
/**
 * What everyone else said, before you say it.
 *
 * Two histograms, maturity and familiarity, with the summary statistics that
 * are meaningful for each. Maturity gets the same best/average/worst-case band
 * the grid itself reads (see lib/community.ts), never a mean: the six tiers
 * are names, not numbers, and two of them are not even rungs on the evidence
 * ladder. Familiarity is a 0–3 scale with equal steps, so it does have a mean.
 *
 * The bars are drawn from the same counts the grid's community view is built
 * from, so what a reader sees here is exactly what their vote will move.
 */
import { computed } from 'vue'
import { atlas, tierOrder } from '@/lib/atlas'
import { bandTier, numericStats, tierStats, totalOf } from '@/lib/community'
import { familiarityFill, familiarityInk, tierAbbrev, tierStyle } from '@/lib/scales'
import type { RatingSummary } from '@/types/atlas'

const props = defineProps<{
  /** What the histograms are about, e.g. "Interpretability" or "IN6". */
  subject: string
  stats: RatingSummary | undefined
  /** Shown when the subject has no votes of its own yet. */
  emptyNote?: string
}>()

const FAMILIARITY_LEVELS = [0, 1, 2, 3, 4]

const maturity = computed(() => {
  const counts = props.stats?.maturity ?? {}
  const rows = tierOrder.map((tier) => ({ tier, n: counts[tier] ?? 0 }))
  const most = Math.max(1, ...rows.map((r) => r.n))
  const tierCounts = Object.fromEntries(rows.map((r) => [r.tier, r.n]))
  return {
    rows: rows.map((r) => ({ ...r, width: (r.n / most) * 100 })),
    stats: {
      n: totalOf(tierCounts),
      best: bandTier(tierCounts, 'best'),
      typical: bandTier(tierCounts, 'typical'),
      worst: bandTier(tierCounts, 'worst'),
      bimodal: tierStats(tierCounts).bimodal,
    },
  }
})

const familiarity = computed(() => {
  const counts = props.stats?.familiarity ?? {}
  const rows = FAMILIARITY_LEVELS.map((level) => ({
    level,
    label: atlas.legend.familiarity_scale[String(level)] ?? String(level),
    n: counts[String(level)] ?? 0,
  }))
  const most = Math.max(1, ...rows.map((r) => r.n))
  return {
    rows: rows.map((r) => ({ ...r, width: (r.n / most) * 100 })),
    stats: numericStats(counts, FAMILIARITY_LEVELS),
  }
})

const total = computed(() => props.stats?.count ?? 0)
</script>

<template>
  <section class="dist" aria-labelledby="dist-title">
    <h4 id="dist-title" class="dist__title">
      What researchers have said about {{ subject }}
      <span class="dist__n tabular">{{ total }} {{ total === 1 ? 'response' : 'responses' }}</span>
    </h4>

    <p v-if="total === 0" class="dist__empty">
      {{ emptyNote ?? 'Nobody has rated this yet. Yours would be the first.' }}
    </p>

    <template v-else>
      <div class="dist__block">
        <p class="dist__label">Maturity</p>
        <ul class="bars">
          <li v-for="row in maturity.rows" :key="row.tier" class="bar" :class="{ 'bar--zero': row.n === 0 }">
            <span
              class="bar__chip"
              :class="{ hatched: tierStyle(row.tier).hatched }"
              :style="{ backgroundColor: tierStyle(row.tier).fill, color: tierStyle(row.tier).ink }"
              >{{ tierAbbrev(row.tier) }}</span
            >
            <span class="bar__name">{{ row.tier }}</span>
            <span class="bar__track">
              <span
                class="bar__fill"
                :class="{ hatched: tierStyle(row.tier).hatched }"
                :style="{ width: `${row.width}%`, backgroundColor: tierStyle(row.tier).fill }"
              ></span>
            </span>
            <span class="bar__n tabular">{{ row.n }}</span>
          </li>
        </ul>
        <p class="dist__stats">
          Best case <strong>{{ maturity.stats.best ?? 'n/a' }}</strong> · Average
          <strong>{{ maturity.stats.typical ?? 'n/a' }}</strong> · Worst case
          <strong>{{ maturity.stats.worst ?? 'n/a' }}</strong>
          <template v-if="maturity.stats.bimodal">
            · <em>two peaks, the responses disagree rather than spread</em>
          </template>
        </p>
      </div>

      <div class="dist__block">
        <p class="dist__label">Familiarity of the people who answered</p>
        <ul class="bars">
          <li v-for="row in familiarity.rows" :key="row.level" class="bar" :class="{ 'bar--zero': row.n === 0 }">
            <span
              class="bar__chip tabular"
              :style="{ backgroundColor: familiarityFill(row.level), color: familiarityInk(row.level) }"
              >{{ row.level }}</span
            >
            <span class="bar__name">{{ row.label }}</span>
            <span class="bar__track">
              <span class="bar__fill" :style="{ width: `${row.width}%`, backgroundColor: familiarityFill(row.level) }"></span>
            </span>
            <span class="bar__n tabular">{{ row.n }}</span>
          </li>
        </ul>
        <p class="dist__stats">
          Mean <strong class="tabular">{{ familiarity.stats.mean }}</strong> · Median
          <strong class="tabular">{{ familiarity.stats.median }}</strong>
          <template v-if="familiarity.stats.bimodal"> · two peaks</template>
        </p>
      </div>

      <p class="dist__note">
        Maturity has no mean here on purpose: the six tiers are names, not numbers, and
        <em>Contested</em> and <em>Never demonstrated</em> are not rungs on the evidence ladder,
        one is a live dispute, the other a negative finding.
      </p>
    </template>
  </section>
</template>

<style scoped>
.dist {
  background: var(--surface-sunken);
  border-radius: var(--radius);
  padding: 0.85rem 1rem 0.95rem;
}

.dist__title {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem;
  font-size: 0.9375rem;
  margin: 0 0 0.75rem;
}

.dist__n {
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--ink-muted);
}

.dist__empty {
  margin: 0;
  font-size: 0.875rem;
  color: var(--ink-secondary);
}

.dist__block + .dist__block {
  margin-top: 0.9rem;
}

.dist__label {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  font-weight: 600;
  color: var(--ink-muted);
  margin: 0 0 0.35rem;
}

.bars {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.2rem;
}

.bar {
  display: grid;
  grid-template-columns: 1.5rem minmax(0, 11rem) minmax(3rem, 1fr) 1.6rem;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.75rem;
}

.bar--zero {
  opacity: 0.55;
}

.bar__chip {
  text-align: center;
  border-radius: 3px;
  padding: 0.05rem 0;
  font-size: 0.5625rem;
  font-weight: 700;
}

.bar__name {
  color: var(--ink-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bar__track {
  background: var(--surface);
  border-radius: 2px;
  height: 0.6rem;
  overflow: hidden;
}

.bar__fill {
  display: block;
  height: 100%;
  min-width: 1px;
  border-radius: 2px;
}

.bar__n {
  text-align: right;
  color: var(--ink-secondary);
  font-weight: 600;
}

.dist__stats {
  margin: 0.4rem 0 0;
  font-size: 0.75rem;
  color: var(--ink-secondary);
  max-width: none;
}

.dist__stats em {
  color: var(--ink);
}

.dist__note {
  margin: 0.75rem 0 0;
  font-size: 0.6875rem;
  color: var(--ink-muted);
  max-width: none;
}

@media (max-width: 560px) {
  .bar {
    grid-template-columns: 1.5rem minmax(0, 1fr) 1.6rem;
  }

  .bar__track {
    display: none;
  }
}
</style>
