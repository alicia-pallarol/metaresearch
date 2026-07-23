<script setup lang="ts">
/**
 * The condensed 12 problems × 12 research areas grid, the hero visual.
 *
 * Two switches read the same grid four ways (see lib/community.ts for the
 * arithmetic):
 *   · SOURCE    our curated ratings, or the community's maturity votes. Which
 *               agendas address which problem is curated in both, only the
 *               colour changes hands.
 *   · SCENARIO  the strongest quarter of the ratings behind each cell, the middle
 *               of them, or the weakest quarter.
 *
 * Interaction: a single tab stop with arrow-key movement (the standard grid
 * pattern), Enter/Space to drill down, and a popover beside the cell on hover and
 * on focus.
 */
import { computed, nextTick, ref } from 'vue'
import { agendasReaching, atlas, getAreaByName, getProblem } from '@/lib/atlas'
import {
  agendaReading,
  gridCell,
  SCENARIO_HINTS,
  type GridCell,
  type GridSource,
  type Scenario,
} from '@/lib/community'
import { describeCell, tierAbbrev, tierStyle } from '@/lib/scales'
import { useSummary } from '@/composables/useSummary'
import ModeSwitch from '@/components/ModeSwitch.vue'
import type { Tier } from '@/types/atlas'

defineProps<{
  selected: { problemId: string; areaName: string } | null
}>()

const emit = defineEmits<{
  (e: 'select', value: { problemId: string; areaName: string }): void
}>()

const source = defineModel<GridSource>('source', { required: true })
const scenario = defineModel<Scenario>('scenario', { required: true })

const { summary, state } = useSummary()

const areas = atlas.area_matrix.area_order
const problems = atlas.problems

const cellRefs = ref<Record<string, HTMLButtonElement | null>>({})
const focused = ref({ row: 0, col: 0 })
const hovered = ref<{ problemId: string; areaName: string; x: number; y: number; above: boolean } | null>(null)

function cellKey(row: number, col: number): string {
  return `${row}:${col}`
}

function tagOf(areaName: string): string {
  return getAreaByName(areaName)?.tag ?? ''
}

/** Every cell under the current switches, recomputed when either one moves. */
const cells = computed(() => {
  const map = new Map<string, GridCell>()
  for (const problem of problems) {
    for (const area of areas) {
      map.set(
        `${problem.id}|${area}`,
        gridCell(source.value, scenario.value, tagOf(area), area, problem.id, summary.value),
      )
    }
  }
  return map
})

function cellOf(problemId: string, areaName: string): GridCell {
  return cells.value.get(`${problemId}|${areaName}`) ?? { tier: null, agendaCount: 0, ratings: 0, inherited: false }
}

/** No agenda in this area reaches the problem, a claim, not missing data. */
function isBlank(problemId: string, areaName: string): boolean {
  return cellOf(problemId, areaName).agendaCount === 0
}

/** The structure is there but nobody has voted on it yet. */
function isUnvoted(problemId: string, areaName: string): boolean {
  const cell = cellOf(problemId, areaName)
  return cell.agendaCount > 0 && cell.tier === null
}

function setCellRef(row: number, col: number, el: unknown): void {
  cellRefs.value[cellKey(row, col)] = (el as HTMLButtonElement | null) ?? null
}

async function moveFocus(row: number, col: number): Promise<void> {
  const r = Math.max(0, Math.min(problems.length - 1, row))
  const c = Math.max(0, Math.min(areas.length - 1, col))
  focused.value = { row: r, col: c }
  await nextTick()
  cellRefs.value[cellKey(r, c)]?.focus()
}

/**
 * The popover is anchored to the cell rather than to the pointer, so it sits in
 * the same place whether it was opened by hovering or by arrowing onto the cell.
 * It flips above when there is not room below, and is clamped to the viewport so
 * the rightmost column never pushes it off-screen.
 */
const TIP_WIDTH = 330

function openTip(el: EventTarget | null, problemId: string, areaName: string): void {
  const rect = (el as HTMLElement | null)?.getBoundingClientRect()
  if (!rect) return

  const above = rect.bottom + 210 > window.innerHeight && rect.top > 210
  const half = TIP_WIDTH / 2
  const x = Math.min(Math.max(rect.left + rect.width / 2, half + 8), Math.max(window.innerWidth - half - 8, half + 8))

  hovered.value = { problemId, areaName, x, y: above ? rect.top - 8 : rect.bottom + 8, above }
}

function closeTip(): void {
  hovered.value = null
}

/** Focus and hover share one popover, so keyboard readers get the same detail. */
function onCellFocus(event: FocusEvent, row: number, col: number, problemId: string, areaName: string): void {
  focused.value = { row, col }
  openTip(event.currentTarget, problemId, areaName)
}

function onKeydown(event: KeyboardEvent, row: number, col: number): void {
  const moves: Record<string, [number, number]> = {
    ArrowUp: [row - 1, col],
    ArrowDown: [row + 1, col],
    ArrowLeft: [row, col - 1],
    ArrowRight: [row, col + 1],
    Home: [row, 0],
    End: [row, areas.length - 1],
    PageUp: [0, col],
    PageDown: [problems.length - 1, col],
  }
  const target = moves[event.key]
  if (event.key === 'Escape') {
    closeTip()
    return
  }
  if (!target) return
  event.preventDefault()
  void moveFocus(target[0], target[1])
}

/** How many agendas to name in the popover before collapsing to "+N more". */
const MAX_TIP_AGENDAS = 5

/**
 * The popover is where the opaque codes get decoded: it names the problem in full
 * and lists the individual agendas behind the cell (strongest first), each with
 * the tier it carries under the switches currently set, so what you read always
 * matches what you see.
 */
const tooltip = computed(() => {
  const target = hovered.value
  if (!target) return null

  const cell = cellOf(target.problemId, target.areaName)
  const problem = getProblem(target.problemId)
  const areaTag = tagOf(target.areaName)
  const reaching = agendasReaching(target.areaName, target.problemId)

  const rows = reaching.map((row) => {
    const community = agendaReading(summary.value, areaTag, row.agenda.id, scenario.value)
    return {
      id: row.agenda.id,
      name: row.agenda.agenda,
      tier: source.value === 'research' ? row.tier : community.tier,
      inherited: source.value === 'community' && community.inherited,
      n: source.value === 'community' ? community.n : 0,
    }
  })

  return {
    x: target.x,
    y: target.y,
    above: target.above,
    // Same title as the drill-down panel: the column (area) first, then the row
    // (problem), horizontal coordinate before vertical, as coordinates read.
    title: {
      area: target.areaName,
      problemId: problem?.id ?? target.problemId,
      problemName: problem?.name ?? '',
    },
    tier: cell.tier,
    summary: tipSummary(cell),
    agendas: rows.slice(0, MAX_TIP_AGENDAS),
    more: Math.max(0, rows.length - MAX_TIP_AGENDAS),
  }
})

function tipSummary(cell: GridCell): string {
  if (cell.agendaCount === 0) return 'Blank: this area does not address this problem'
  const agendas = `${cell.agendaCount} agenda${cell.agendaCount === 1 ? '' : 's'}`
  if (source.value === 'research') return `${agendas} reach it, strongest first`
  if (cell.tier === null) return `${agendas} here, none of them rated by anyone yet`
  return `${cell.ratings} rating${cell.ratings === 1 ? '' : 's'} of this area, across ${agendas}`
}

/* --- Rendering ------------------------------------------------------------ */

function isSelected(problemId: string, areaName: string, selected: { problemId: string; areaName: string } | null): boolean {
  return selected?.problemId === problemId && selected.areaName === areaName
}

/** Blank stays blank: the shape of what is covered is information in itself. */
function cellStyle(problemId: string, areaName: string) {
  const tier = cellOf(problemId, areaName).tier
  if (tier === null) return { backgroundColor: 'var(--cell-blank)', color: 'var(--ink-muted)' }
  const style = tierStyle(tier)
  return { backgroundColor: style.fill, color: style.ink }
}

/**
 * The label is the count of agendas reaching the problem, the same number in
 * both views, so the two are read side by side. The tier is the cell's colour
 * (see the legend). Colour = how strong, number = how many.
 */
function cellLabel(problemId: string, areaName: string): string {
  const cell = cellOf(problemId, areaName)
  return cell.agendaCount > 0 ? String(cell.agendaCount) : ''
}

function ariaLabel(problemId: string, areaName: string): string {
  const cell = cellOf(problemId, areaName)
  const problem = getProblem(problemId)
  const head = `${problem?.id ?? problemId} ${problem?.name ?? ''}, ${areaName}.`

  if (source.value === 'research') return `${head} ${describeCell(cell.tier, cell.agendaCount || undefined)}`
  if (cell.agendaCount === 0) return `${head} ${describeCell(null)}`
  if (cell.tier === null) return `${head} Not rated by the community yet. ${cell.agendaCount} agendas here.`
  return `${head} Community: ${cell.tier}, from ${cell.ratings} rating${cell.ratings === 1 ? '' : 's'} across ${cell.agendaCount} agendas.`
}

const SOURCE_OPTIONS = [
  { value: 'research' as const, label: 'Our research' },
  { value: 'community' as const, label: 'The community' },
]

const SCENARIO_OPTIONS = [
  { value: 'best' as const, label: 'Best case' },
  { value: 'typical' as const, label: 'Typical' },
  { value: 'worst' as const, label: 'Worst case' },
]

const caption = computed(() => {
  const whose =
    source.value === 'research'
      ? 'what we read and concluded'
      : 'how researchers who answered rated the maturity of these areas'
  return `Colour is ${whose}, taken as ${SCENARIO_HINTS[scenario.value]}.`
})

/** Said plainly rather than hidden: an empty community view is not a broken one. */
const communityNotice = computed(() => {
  if (source.value !== 'community') return ''
  if (state.value === 'disabled') return 'Feedback collection is switched off on this deployment, so there is nothing to show here yet.'
  if (state.value === 'unavailable') return 'The feedback server is asleep or unreachable, so the community view is empty. Our own ratings are unaffected.'
  if ((summary.value?.totals.total_submissions ?? 0) === 0) return 'Nobody has voted yet. The first rating will show up here.'
  return ''
})

function tipChipStyle(tier: Tier | null) {
  const style = tierStyle(tier)
  return { backgroundColor: style.fill, color: style.ink }
}
</script>

<template>
  <figure class="heatmap">
    <figcaption class="heatmap__head">
      <div class="heatmap__intro">
        <h2 id="area-heatmap-title" class="heatmap__title">Research areas against problems</h2>
        <p class="heatmap__sub">
          {{ caption }} The <strong>number</strong> is how many agendas in that area reach the problem
          at all. Hover a cell to name them, strongest first; click to open them.
        </p>
      </div>

      <div class="heatmap__switches">
        <ModeSwitch v-model="source" name="grid-source" label="Rated by" :options="SOURCE_OPTIONS" />
        <ModeSwitch v-model="scenario" name="grid-scenario" label="Reading" :options="SCENARIO_OPTIONS" />
      </div>
    </figcaption>

    <p v-if="communityNotice" class="heatmap__notice">{{ communityNotice }}</p>

    <div class="heatmap__scroll">
      <div
        class="grid"
        role="grid"
        aria-labelledby="area-heatmap-title"
        :style="{ '--cols': areas.length }"
      >
        <div class="grid__row grid__row--head" role="row">
          <div class="grid__corner" role="columnheader">Problem \ Area</div>
          <div v-for="area in areas" :key="area" class="grid__colhead" role="columnheader">
            <abbr :title="area">{{ getAreaByName(area)?.tag ?? area }}</abbr>
          </div>
        </div>

        <div v-for="(problem, row) in problems" :key="problem.id" class="grid__row" role="row">
          <div class="grid__rowhead" role="rowheader" :title="`${problem.id} · ${problem.name}`">
            <span class="grid__pid tabular">{{ problem.id }}</span>
            <span class="grid__pname">{{ problem.name }}</span>
          </div>

          <button
            v-for="(area, col) in areas"
            :key="area"
            :ref="(el) => setCellRef(row, col, el)"
            type="button"
            role="gridcell"
            class="cell"
            :class="{
              'cell--blank': isBlank(problem.id, area),
              'cell--unvoted': isUnvoted(problem.id, area),
              'cell--selected': isSelected(problem.id, area, selected),
              hatched: tierStyle(cellOf(problem.id, area).tier).hatched,
            }"
            :style="cellStyle(problem.id, area)"
            :tabindex="focused.row === row && focused.col === col ? 0 : -1"
            :aria-label="ariaLabel(problem.id, area)"
            :aria-selected="isSelected(problem.id, area, selected)"
            @click="emit('select', { problemId: problem.id, areaName: area })"
            @focus="onCellFocus($event, row, col, problem.id, area)"
            @blur="closeTip"
            @mouseenter="openTip($event.currentTarget, problem.id, area)"
            @mouseleave="closeTip"
            @keydown="onKeydown($event, row, col)"
          >
            <span class="cell__label tabular">{{ cellLabel(problem.id, area) }}</span>
          </button>
        </div>
      </div>
    </div>

    <p class="heatmap__scrollhint" aria-hidden="true">Scroll the grid sideways to see every area →</p>

    <!-- Visual only: the same detail is in each cell's accessible name, so a
         screen reader is not read the grid twice. -->
    <Teleport to="body">
      <div
        v-if="tooltip"
        class="tip"
        :class="{ 'tip--above': tooltip.above }"
        :style="{ left: `${tooltip.x}px`, top: `${tooltip.y}px`, width: `${TIP_WIDTH}px` }"
        aria-hidden="true"
      >
        <div class="tip__head">
          <!-- prettier-ignore -->
          <span class="tip__title">{{ tooltip.title.area }} <span class="tip__x">×</span> <span class="tabular">{{ tooltip.title.problemId }}</span> {{ tooltip.title.problemName }}</span>
          <span class="tip__summary">{{ tooltip.summary }}</span>
        </div>

        <ul v-if="tooltip.agendas.length" class="tip__agendas">
          <li v-for="a in tooltip.agendas" :key="a.id" class="tip__agenda">
            <span
              class="tip__chip"
              :class="{ hatched: tierStyle(a.tier).hatched, 'tip__chip--none': a.tier === null }"
              :style="tipChipStyle(a.tier)"
              >{{ a.tier ? tierAbbrev(a.tier) : '–' }}</span
            >
            <span class="tip__aid tabular">{{ a.id }}</span>
            <span class="tip__aname">{{ a.name }}</span>
            <span class="tip__atier">{{ a.tier ?? 'no votes' }}<template v-if="a.inherited"> ·&nbsp;area</template></span>
          </li>
          <li v-if="tooltip.more > 0" class="tip__more">+{{ tooltip.more }} more, click to open all</li>
        </ul>

        <p v-if="source === 'community'" class="tip__foot">
          <template v-if="tooltip.agendas.some((a) => a.inherited)">
            “area” marks an agenda nobody has rated on its own yet, it is showing its area's votes.
          </template>
          <template v-else-if="tooltip.tier">Voted agenda by agenda.</template>
        </p>
      </div>
    </Teleport>
  </figure>
</template>

<style scoped>
.heatmap {
  margin: 0;
}

.heatmap__head {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem 1.5rem;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.heatmap__intro {
  flex: 1 1 24rem;
}

.heatmap__switches {
  display: grid;
  gap: 0.45rem;
  justify-items: end;
}

.heatmap__title {
  margin-bottom: 0.25rem;
}

.heatmap__sub {
  color: var(--ink-secondary);
  font-size: 0.9375rem;
  margin: 0;
}

.heatmap__notice {
  font-size: 0.875rem;
  color: var(--ink-secondary);
  background: var(--surface-sunken);
  border-radius: var(--radius);
  padding: 0.6rem 0.85rem;
  margin: 0 0 0.9rem;
  max-width: var(--measure);
}

.heatmap__scroll {
  overflow-x: auto;
  padding-bottom: 0.35rem;
}

.grid {
  min-width: 700px;
  display: grid;
  gap: 2px;
}

.grid__row {
  display: grid;
  grid-template-columns: minmax(190px, 250px) repeat(var(--cols), minmax(38px, 1fr));
  gap: 2px;
}

.grid__corner,
.grid__colhead {
  font-size: 0.6875rem;
  color: var(--ink-muted);
  font-weight: 600;
  letter-spacing: 0.03em;
  padding-bottom: 0.3rem;
  align-self: end;
}

.grid__colhead {
  text-align: center;
}

.grid__colhead abbr {
  text-decoration: none;
  border-bottom: 1px dotted var(--rule-strong);
  cursor: help;
}

/* Sticky so the problem stays readable while the areas scroll under it. */
.grid__rowhead {
  display: flex;
  gap: 0.45rem;
  align-items: baseline;
  font-size: 0.8125rem;
  padding-right: 0.5rem;
  min-width: 0;
  position: sticky;
  left: 0;
  z-index: 3;
  background: var(--page);
}

.grid__corner {
  position: sticky;
  left: 0;
  z-index: 3;
  background: var(--page);
}

.heatmap__scrollhint {
  display: none;
  font-size: 0.75rem;
  color: var(--ink-muted);
  margin: 0.4rem 0 0;
}

.grid__pid {
  color: var(--ink-muted);
  font-weight: 600;
  font-size: 0.75rem;
  flex: none;
}

.grid__pname {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ink-secondary);
}

.cell {
  border: 0;
  border-radius: 3px;
  min-height: 34px;
  cursor: pointer;
  display: grid;
  place-items: center;
  font-size: 0.8125rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  padding: 0;
  transition: transform 100ms ease, box-shadow 100ms ease;
}

.cell:hover {
  transform: scale(1.06);
  box-shadow: var(--shadow);
  z-index: 1;
}

.cell--blank {
  border: 1px dashed var(--rule);
  cursor: default;
}

/* Rated by us, not yet by anyone else: the structure shows, the colour does not. */
.cell--unvoted {
  border: 1px dotted var(--rule-strong);
  color: var(--ink-muted);
}

.cell--selected {
  box-shadow: 0 0 0 2px var(--ink);
  z-index: 2;
}

.cell__label {
  pointer-events: none;
}

@media (max-width: 860px) {
  .heatmap__scrollhint {
    display: block;
  }

  .heatmap__switches {
    justify-items: start;
  }
}

@media (max-width: 640px) {
  .grid__row {
    grid-template-columns: minmax(120px, 150px) repeat(var(--cols), minmax(34px, 1fr));
  }

  .grid {
    min-width: 560px;
  }

  .grid__pname {
    font-size: 0.75rem;
  }
}
</style>

<style>
/* Teleported to <body>, so not scoped: it must not be clipped by the grid's
   horizontal scroll container. Pointer-transparent, so it can never sit between
   the pointer and the cell it describes. */
.tip {
  position: fixed;
  z-index: 60;
  transform: translateX(-50%);
  pointer-events: none;
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--rule-strong);
  border-radius: var(--radius);
  background: var(--surface);
  box-shadow: var(--shadow);
  font-size: 0.8125rem;
  line-height: 1.4;
}

.tip--above {
  transform: translate(-50%, -100%);
}

.tip__head {
  display: grid;
  gap: 0.1rem;
  margin-bottom: 0.4rem;
}

.tip__title {
  font-size: 0.875rem;
  font-weight: 600;
}

.tip__x {
  color: var(--ink-muted);
  font-weight: 400;
}

.tip__summary {
  color: var(--ink-secondary);
  font-size: 0.75rem;
}

.tip__agendas {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.2rem;
}

.tip__agenda {
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  min-width: 0;
}

.tip__chip {
  flex: none;
  align-self: center;
  min-width: 1.5rem;
  text-align: center;
  border-radius: 3px;
  padding: 0.05rem 0.3rem;
  font-size: 0.5625rem;
  font-weight: 700;
}

.tip__chip--none {
  border: 1px dotted var(--rule-strong);
}

.tip__aid {
  flex: none;
  font-weight: 700;
  font-size: 0.6875rem;
  color: var(--ink-muted);
}

.tip__aname {
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  flex: 1 1 auto;
}

.tip__atier {
  flex: none;
  color: var(--ink-muted);
  font-size: 0.6875rem;
}

.tip__more,
.tip__foot {
  color: var(--ink-muted);
  font-size: 0.6875rem;
}

.tip__more {
  padding-left: 1.9rem;
}

.tip__foot {
  margin: 0.4rem 0 0;
  max-width: none;
}
</style>
