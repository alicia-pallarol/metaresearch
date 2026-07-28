<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
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

/** The hero title is a shortcut to the grid: same smooth scroll as a cell click. */
function scrollToGrid(): void {
  gridSection.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
}

// A small decorative square beside the hero: a grid of cells in the four maturity
// blues, echoing the real grid below. Cells slowly fade in, shift colour and fade
// out; hovering one holds it lit and recoloured. Purely ornamental, hidden from
// assistive tech. The fill is passed as a custom property so the hover rule in
// CSS can override the colour (an inline background could not be).
const artColumns = 6
const artRows = 6
const artCycle = 18 // seconds; see cellCycle below

// Deterministic 0..1 hash, so the layout is scattered (not diagonal) but stable
// across renders rather than reshuffling on every paint.
const hash = (n: number): number => {
  const x = Math.sin(n) * 43758.5453
  return x - Math.floor(x)
}

const blues = ['var(--tier-untested)', 'var(--tier-early)', 'var(--tier-strong)', 'var(--tier-robust)']
const artCells = Array.from({ length: artColumns * artRows }, (_, i) => ({
  // Random colour per cell, so the blues are scattered rather than striped.
  fill: blues[Math.floor(hash(i + 1) * blues.length)],
  // Random negative offset desyncs the cells: at any moment some are fading in
  // while others fade out, with no diagonal wave running through them.
  delay: `-${(hash((i + 1) * 1.7) * artCycle).toFixed(2)}s`,
}))

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
        <h1 class="hero__title">
          <a class="hero__title-link" href="#grid" @click.prevent="scrollToGrid">The Map</a>
        </h1>

        <div class="hero__inner">
          <div class="hero__body">
            <p class="hero__para">
              A crowdsourced view of where AI safety research stands.<br />
              Anyone can weigh in, and we want you to!
            </p>
            <p class="hero__para">
              The <a class="hero__maplink" href="#grid" @click.prevent="scrollToGrid">Map</a> covers
              {{ atlas.agendas.length }} research agendas across {{ atlas.areas.length }} areas,
              rated against {{ atlas.problems.length }} open problems. Each rating answers one
              question: how mature is the evidence that this line of work addresses this problem?
            </p>
            <p class="hero__para">
              Open any cell to see the reasoning behind a rating and what others have submitted, then
              add your own. Every response feeds back into the map.
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
          </div>

          <div class="hero__art" aria-hidden="true">
            <div class="cellgrid" :style="{ '--cols': artColumns, '--rows': artRows }">
              <span
                v-for="(cell, i) in artCells"
                :key="i"
                class="cellgrid__cell"
                :style="{ '--fill': cell.fill, animationDelay: cell.delay }"
              />
            </div>
          </div>
        </div>

        <p class="hero__cta">
          <a class="btn" href="#priorities">Where to focus</a>
          <a class="btn btn--quiet" href="#agendas">Browse every agenda →</a>
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
            cell, with the box for adding your own rating.
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

/* The paragraphs + counters sit on the left; the decorative square on the right.
   align-items: stretch makes the square exactly as tall as that text block, from
   the first paragraph down to the bottom of the counters. It collapses to a
   single column (art hidden) below the breakpoint so nothing crowds on phones. */
.hero__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: clamp(2rem, 5vw, 4rem);
}

/* Sized to its content (the ~60ch paragraphs), never growing, so space-between
   can park the square out at the right margin with the text staying on the left. */
.hero__body {
  min-width: 0;
  flex: 0 1 auto;
}

.hero__title {
  max-width: 20ch;
}

/* The title doubles as a jump link to the grid. Underlined so it reads as
   clickable; it stays the same colour as the heading rather than link-blue. */
.hero__title-link {
  color: inherit;
  text-decoration: underline;
  text-decoration-thickness: 2px;
  text-underline-offset: 5px;
  text-decoration-color: var(--tier-early);
  cursor: pointer;
  transition: text-decoration-color 0.15s ease;
}

.hero__title-link:hover,
.hero__title-link:focus-visible {
  text-decoration-color: var(--tier-robust);
}

/* --- Decorative maturity-blue cell grid ------------------------------------ */
.hero__art {
  flex: 0 0 auto;
  display: none; /* wide screens only, enabled in the media query below */
}

/* A square, sized by its WIDTH (bounded by the viewport) with the height derived
   from aspect-ratio. Sizing width-first is what keeps the flex box exactly as
   wide as the square, so it can never spill past the right edge. The width lands
   close to the text block's height, so it still reads as "about as tall as the
   text". Equal rows and columns keep the cells square too. */
.cellgrid {
  display: grid;
  grid-template-columns: repeat(var(--cols, 6), 1fr);
  grid-template-rows: repeat(var(--rows, 6), 1fr);
  gap: 7px;
  width: clamp(240px, 28vw, 340px);
  aspect-ratio: 1 / 1;
}

/* Each cell slowly fades in, shifts colour, and fades out again, on a long loop
   (18s) with a desynced start, so cells appear and disappear independently. */
.cellgrid__cell {
  aspect-ratio: 1;
  border-radius: 4px;
  background: var(--fill);
  opacity: 0.55;
  animation: cellCycle 18s ease-in-out infinite;
  transition:
    background 0.35s ease,
    opacity 0.35s ease;
}

@keyframes cellCycle {
  0% {
    opacity: 0;
    background: var(--fill);
  }
  40% {
    opacity: 0.82;
    background: var(--fill);
  }
  70% {
    opacity: 0.82;
    background: var(--focus);
  }
  100% {
    opacity: 0;
    background: var(--focus);
  }
}

/* Hovering drops the loop and holds the cell full and recoloured; the transition
   makes it ease in and out rather than snap. */
.cellgrid__cell:hover {
  animation: none;
  opacity: 1;
  background: var(--focus);
}

/* Only show the square once the row is wide enough to hold the text and the
   square side by side. Below this it hides. */
@media (min-width: 1024px) {
  .hero__art {
    display: flex;
  }
}

/* For readers who ask for less motion, hold the cells still (no twinkle) but keep
   the hover recolour, which is a colour change rather than movement. */
@media (prefers-reduced-motion: reduce) {
  .cellgrid__cell {
    animation: none;
    opacity: 0.55;
    transform: none;
  }
}

.hero__para {
  font-size: 1.0625rem;
  color: var(--ink-secondary);
  max-width: 60ch;
  margin: 0 0 1rem;
}

/* The first paragraph carries the invitation, so give it a touch more presence. */
.hero__para:first-of-type {
  color: var(--ink);
}

/* Inline "Map" link inside the paragraph, in the usual link colour. */
.hero__maplink {
  color: var(--link);
  text-underline-offset: 2px;
  cursor: pointer;
}

/* No bottom margin: the counters are the last thing in the block, so the block
   (and the square beside it) ends exactly where the counters end. */
.hero__stats {
  display: flex;
  flex-wrap: wrap;
  gap: 2.25rem;
  margin: 1.6rem 0 0;
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

/* The CTA now spans full width below the text/art row, so it needs its own top
   margin: without it the buttons crowd the counters directly above. */
.hero__cta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  margin-top: 2.25rem;
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
