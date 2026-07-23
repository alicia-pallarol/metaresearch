<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { atlas, getAreaByName, getAreaByTag, tierCounts, tierRank } from '@/lib/atlas'
import { gridCell, type GridSource, type Scenario } from '@/lib/community'
import { useSummary } from '@/composables/useSummary'
import AreaHeatmap from '@/components/AreaHeatmap.vue'
import AgendaHeatmap from '@/components/AgendaHeatmap.vue'
import DrillDownPanel from '@/components/DrillDownPanel.vue'
import FocusAreas from '@/components/FocusAreas.vue'
import TierLegend from '@/components/TierLegend.vue'
import SwissCheesePanel from '@/components/SwissCheesePanel.vue'

const { summary, totalSubmissions, uniqueSubmitters, state } = useSummary()

// Derived, never hard-coded: a new iteration changes these without a code edit.
const counts = tierCounts()
const ratedCells = computed(() => Object.values(counts).reduce((sum, n) => sum + n, 0))
const robustCells = computed(() => counts['Robust (small scale)'] ?? 0)

// How the grid is being read. Owned here because the drill-down below it has to
// answer for the same view the cell was clicked in.
const source = ref<GridSource>('research')
const scenario = ref<Scenario>('best')

const selected = ref<{ problemId: string; areaName: string } | null>(null)
const gridSection = ref<HTMLElement | null>(null)
const detailSlot = ref<HTMLElement | null>(null)

/**
 * The detail always opens in the same place, directly under the grid, and the
 * page comes to it, clicking a cell should never leave the reader hunting for
 * what changed further down.
 */
async function onSelect(cell: { problemId: string; areaName: string }): Promise<void> {
  const isSame = selected.value?.problemId === cell.problemId && selected.value.areaName === cell.areaName
  selected.value = isSame ? null : cell
  if (isSame) return

  await nextTick()
  detailSlot.value?.scrollIntoView?.({ behavior: 'smooth', block: 'nearest' })
}

/** "Show P1 in the grid": jump to the strongest area against that problem. */
async function focusProblem(problemId: string): Promise<void> {
  let best: { areaName: string; rank: number } | null = null
  for (const areaName of atlas.area_matrix.area_order) {
    const tag = getAreaByName(areaName)?.tag ?? ''
    const { tier } = gridCell(source.value, scenario.value, tag, areaName, problemId, summary.value)
    if (tier === null) continue
    const rank = tierRank(tier)
    if (!best || rank < best.rank) best = { areaName, rank }
  }
  if (best) selected.value = { problemId, areaName: best.areaName }

  await nextTick()
  gridSection.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
}

/** "Show Interpretability in the grid": jump to its strongest cell against any problem. */
async function focusArea(areaTag: string): Promise<void> {
  const areaName = getAreaByTag(areaTag)?.name
  if (areaName) {
    let best: { problemId: string; rank: number } | null = null
    for (const problem of atlas.problems) {
      const { tier } = gridCell(source.value, scenario.value, areaTag, areaName, problem.id, summary.value)
      if (tier === null) continue
      const rank = tierRank(tier)
      if (!best || rank < best.rank) best = { problemId: problem.id, rank }
    }
    if (best) selected.value = { problemId: best.problemId, areaName }
  }

  await nextTick()
  gridSection.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
}
</script>

<template>
  <div>
    <section class="hero">
      <div class="wrap">
        <p class="hero__badge">{{ atlas.meta.iteration_label }}</p>
        <h1 class="hero__title">{{ atlas.meta.title }}</h1>
        <p class="hero__lead">
          {{ atlas.agendas.length }} research agendas across {{ atlas.areas.length }} areas, rated
          against {{ atlas.problems.length }} open problems in AI safety. Every rating answers one
          question: <em>how mature is the evidence that this line of work addresses this specific
          problem?</em> The ratings are read and written by hand, one source at a time; no model
          generated any of this.
        </p>
        <p class="hero__invite">
          It is a draft put out for correction. Open any cell, then any agenda, to see the reasoning;
          and if a rating looks wrong or you know a source we missed, tell us right there.
        </p>

        <dl class="hero__stats">
          <div class="stat">
            <dt>Rated cells</dt>
            <dd class="tabular">{{ ratedCells }}</dd>
          </div>
          <div class="stat">
            <dt>Cells rated Robust</dt>
            <dd class="tabular">{{ robustCells }}</dd>
          </div>
          <div class="stat stat--live">
            <dt>Researcher responses</dt>
            <dd v-if="state === 'ready' || totalSubmissions > 0" class="tabular">
              {{ totalSubmissions }}
              <span v-if="uniqueSubmitters > 0" class="stat__sub">from {{ uniqueSubmitters }} people</span>
            </dd>
            <dd v-else-if="state === 'loading'" class="stat__pending">checking…</dd>
            <dd v-else class="stat__pending">not available right now</dd>
          </div>
        </dl>

        <p class="hero__cta">
          <a class="btn" href="#priorities">Start with where effort pays off</a>
          <RouterLink class="btn btn--quiet" to="/methodology">How to read the ratings →</RouterLink>
        </p>
      </div>
    </section>

    <section ref="gridSection" id="grid" class="section">
      <div class="wrap">
        <AreaHeatmap v-model:source="source" v-model:scenario="scenario" :selected="selected" @select="onSelect" />

        <!-- Always here, selected or not, so the grid never shifts under the
             pointer and the reader knows where the detail will appear. -->
        <div ref="detailSlot" class="detail" :class="{ 'detail--empty': !selected }">
          <DrillDownPanel
            v-if="selected"
            :problem-id="selected.problemId"
            :area-name="selected.areaName"
            :source="source"
            :scenario="scenario"
            @close="selected = null"
          />
          <p v-else class="detail__prompt">
            Click any cell above to open it here: the problem, the area, and every agenda behind that
            cell, with the box for telling us where the rating is wrong.
          </p>
        </div>

        <div class="legendwrap">
          <TierLegend />
        </div>
      </div>
    </section>

    <section id="priorities" class="section">
      <div class="wrap">
        <p class="eyebrow">Focus areas</p>
        <h2>Where the next unit of effort pays off most</h2>
        <FocusAreas
          v-model:source="source"
          @focus-problem="focusProblem"
          @focus-area="focusArea"
        />
      </div>
    </section>

    <section id="agendas" class="section">
      <div class="wrap">
        <AgendaHeatmap />
      </div>
    </section>

    <section id="swiss" class="section">
      <div class="wrap">
        <SwissCheesePanel />
      </div>
    </section>
  </div>
</template>

<style scoped>
.hero {
  padding-block: clamp(2.5rem, 1.5rem + 4vw, 5rem) clamp(2rem, 1rem + 3vw, 3.5rem);
}

.hero__badge {
  display: inline-block;
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.09em;
  font-weight: 600;
  color: var(--ink-secondary);
  border: 1px solid var(--rule-strong);
  border-radius: 999px;
  padding: 0.15rem 0.7rem;
  margin: 0 0 1rem;
}

.hero__title {
  max-width: 20ch;
}

.hero__lead {
  font-size: 1.0625rem;
  color: var(--ink-secondary);
  max-width: 62ch;
}

.hero__invite {
  font-size: 0.9375rem;
  color: var(--ink);
  max-width: 62ch;
  border-left: 3px solid var(--rule-strong);
  padding-left: 0.85rem;
}

.hero__stats {
  display: flex;
  flex-wrap: wrap;
  gap: 2.25rem;
  margin: 2rem 0 1.75rem;
}

.stat dt {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--ink-muted);
  font-weight: 600;
}

.stat dd {
  margin: 0.15rem 0 0;
  font-size: 1.75rem;
  font-weight: 600;
  line-height: 1.1;
}

.stat__sub {
  display: block;
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--ink-muted);
}

.stat__pending {
  font-size: 0.9375rem;
  font-weight: 400;
  color: var(--ink-muted);
}

.hero__cta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}

.hero__cta .btn {
  text-decoration: none;
}

.detail {
  scroll-margin-top: 1rem;
}

.detail--empty {
  margin-top: 1.25rem;
  border: 1px dashed var(--rule-strong);
  border-radius: var(--radius-lg);
  padding: 0.9rem 1.1rem;
}

.detail__prompt {
  margin: 0;
  font-size: 0.875rem;
  color: var(--ink-muted);
  max-width: var(--measure);
}

.legendwrap {
  margin-top: 1.5rem;
  max-width: 780px;
}
</style>
