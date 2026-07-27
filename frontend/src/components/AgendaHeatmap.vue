<script setup lang="ts">
/**
 * The full agenda-level grid: 58 agendas × 12 problems, grouped into collapsible
 * bands by research area.
 *
 * On wide screens each band is a grid. On narrow screens the same data is a list
 * of agendas with their tiers as wrapped chips, a 12-column table on a phone is
 * a scroll maze, and the chips carry the same information.
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { atlas, agendasOfArea, cellTier, coverageOf, getAgenda, getProblem, problemIds } from '@/lib/atlas'
import { describeCell, tierAbbrev, tierStyle } from '@/lib/scales'

// The first band starts open so the section is never a wall of closed rows.
const openAreas = ref<Record<string, boolean>>({ [atlas.areas[0]?.name ?? '']: true })

function toggleArea(name: string): void {
  openAreas.value[name] = !openAreas.value[name]
}

function expandAll(value: boolean): void {
  const next: Record<string, boolean> = {}
  for (const area of atlas.areas) next[area.name] = value
  openAreas.value = next
}

// Blank stays blank so the coverage pattern is legible.
function cellStyle(agendaId: string, problemId: string) {
  const tier = cellTier(agendaId, problemId)
  if (tier === null) return { backgroundColor: 'var(--cell-blank)', color: 'var(--ink-muted)' }
  const style = tierStyle(tier)
  return { backgroundColor: style.fill, color: style.ink }
}

function cellText(agendaId: string, problemId: string): string {
  return tierAbbrev(cellTier(agendaId, problemId))
}

/** Names both codes, so a cell explains itself on hover instead of "EA1 × P1". */
function cellTitle(agendaId: string, problemId: string): string {
  const tier = cellTier(agendaId, problemId)
  const agenda = getAgenda(agendaId)
  const problem = getProblem(problemId)
  const head = `${agendaId} ${agenda?.agenda ?? ''} × ${problemId} ${problem?.name ?? ''}`
  return `${head}, ${describeCell(tier)}`
}

/** "EA6, Model organisms of misalignment", for the row link hover. */
function agendaTitle(agendaId: string): string {
  const agenda = getAgenda(agendaId)
  return agenda ? `${agenda.id}, ${agenda.agenda}` : agendaId
}

/** "P1, Alignment measurement and observability", for the column header hover. */
function problemTitle(problemId: string): string {
  const problem = getProblem(problemId)
  return problem ? `${problem.id}, ${problem.name}: ${problem.short_def}` : problemId
}

/**
 * Interacting with a cell shows a small popover anchored to it, rather than
 * navigating away: it names the pair, gives the tier and the problem in words,
 * and carries the one link out to the agenda's feedback form.
 *
 * Hover (or keyboard focus) previews it; a click pins it so it holds still while
 * you reach for the link, until you click the same cell again, click a different
 * cell, click away, press Escape, or scroll. A pin always wins over a hover, so a
 * pinned popover does not shift as the pointer wanders over other cells.
 */
const POP_WIDTH = 300
type PopAnchor = { agendaId: string; problemId: string; x: number; y: number; above: boolean }

const hoverPop = ref<PopAnchor | null>(null)
const pinnedPop = ref<PopAnchor | null>(null)
const activePop = computed(() => pinnedPop.value ?? hoverPop.value)

let leaveTimer: ReturnType<typeof setTimeout> | undefined

function anchorAt(el: EventTarget | null, agendaId: string, problemId: string): PopAnchor | null {
  const rect = (el as HTMLElement | null)?.getBoundingClientRect()
  if (!rect) return null
  const above = rect.bottom + 190 > window.innerHeight && rect.top > 190
  const half = POP_WIDTH / 2
  const x = Math.min(Math.max(rect.left + rect.width / 2, half + 8), Math.max(window.innerWidth - half - 8, half + 8))
  return { agendaId, problemId, x, y: above ? rect.top - 8 : rect.bottom + 8, above }
}

function cancelLeave(): void {
  if (leaveTimer) {
    clearTimeout(leaveTimer)
    leaveTimer = undefined
  }
}

function showHover(el: EventTarget | null, agendaId: string, problemId: string): void {
  cancelLeave()
  hoverPop.value = anchorAt(el, agendaId, problemId)
}

// A short grace period so moving between cells, or from a cell onto the popover
// itself, does not flicker it shut in the gap between the leave and the enter.
function scheduleLeave(): void {
  cancelLeave()
  leaveTimer = setTimeout(() => {
    hoverPop.value = null
    leaveTimer = undefined
  }, 90)
}

function togglePin(el: EventTarget | null, agendaId: string, problemId: string): void {
  pinnedPop.value =
    pinnedPop.value?.agendaId === agendaId && pinnedPop.value.problemId === problemId
      ? null
      : anchorAt(el, agendaId, problemId)
}

function isPinned(agendaId: string, problemId: string): boolean {
  return pinnedPop.value?.agendaId === agendaId && pinnedPop.value.problemId === problemId
}

function closeAll(): void {
  cancelLeave()
  pinnedPop.value = null
  hoverPop.value = null
}

const popInfo = computed(() => {
  const p = activePop.value
  if (!p) return null
  const agenda = getAgenda(p.agendaId)
  const problem = getProblem(p.problemId)
  return {
    x: p.x,
    y: p.y,
    above: p.above,
    agendaId: p.agendaId,
    agendaName: agenda?.agenda ?? '',
    problemId: p.problemId,
    problemName: problem?.name ?? '',
    problemDef: problem?.short_def ?? '',
    tier: cellTier(p.agendaId, p.problemId),
    to: `/agenda/${p.agendaId}#p-${p.problemId}`,
  }
})

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') closeAll()
}

// Only a pinned popover needs dismissing by an outside click; a click on a cell
// or inside the popover is handled by their own handlers, so it is left alone.
function onDocClick(e: MouseEvent): void {
  if (!pinnedPop.value) return
  const t = e.target as HTMLElement | null
  if (t?.closest('.pop, .agrid__cellbtn, .alist__chip')) return
  pinnedPop.value = null
}

function onScroll(): void {
  if (pinnedPop.value || hoverPop.value) closeAll()
}

onMounted(() => {
  window.addEventListener('keydown', onKey)
  document.addEventListener('click', onDocClick)
  // Capture, so a scroll inside the grid's own overflow container counts too.
  window.addEventListener('scroll', onScroll, { passive: true, capture: true })
})
onBeforeUnmount(() => {
  cancelLeave()
  window.removeEventListener('keydown', onKey)
  document.removeEventListener('click', onDocClick)
  window.removeEventListener('scroll', onScroll, true)
})
</script>

<template>
  <div class="full">
    <div class="full__head">
      <div>
        <h2 class="full__title">Every agenda, every problem</h2>
        <p class="full__sub">
          The complete {{ atlas.agendas.length }} × {{ problemIds.length }} grid the condensed view
          summarises. Open an area to see its agendas.
        </p>
      </div>

      <div class="full__controls">
        <button type="button" class="btn btn--quiet" @click="expandAll(true)">Expand all</button>
        <button type="button" class="btn btn--quiet" @click="expandAll(false)">Collapse all</button>
      </div>
    </div>

    <section v-for="area in atlas.areas" :key="area.tag" class="band">
      <h3 class="band__headwrap">
        <button
          type="button"
          class="band__head"
          :aria-expanded="!!openAreas[area.name]"
          :aria-controls="`band-${area.tag}`"
          @click="toggleArea(area.name)"
        >
          <span class="band__tag tabular">{{ area.tag }}</span>
          <span class="band__name">{{ area.name }}</span>
          <span class="band__count">{{ area.agenda_ids.length }} agendas</span>
          <span class="band__caret" aria-hidden="true">{{ openAreas[area.name] ? '−' : '+' }}</span>
        </button>
      </h3>

      <div v-show="openAreas[area.name]" :id="`band-${area.tag}`" class="band__body">
        <p class="band__def">{{ area.definition }}</p>

        <!-- Wide screens: the grid -->
        <div class="band__scroll">
          <table class="agrid">
            <caption class="visually-hidden">
              {{ area.name }}: evidence maturity of each agenda against each problem
            </caption>
            <thead>
              <tr>
                <th scope="col" class="agrid__corner">Agenda</th>
                <th v-for="pid in problemIds" :key="pid" scope="col" class="agrid__colhead">
                  <abbr :title="problemTitle(pid)">{{ pid }}</abbr>
                </th>
                <th scope="col" class="agrid__colhead agrid__colhead--wide">Reach</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="agenda in agendasOfArea(area.name)" :key="agenda.id">
                <th scope="row" class="agrid__rowhead">
                  <RouterLink :to="`/agenda/${agenda.id}`" :title="agendaTitle(agenda.id)">
                    <span class="tabular agrid__id">{{ agenda.id }}</span>
                    <span class="agrid__name">{{ agenda.agenda }}</span>
                  </RouterLink>
                </th>
                <td
                  v-for="pid in problemIds"
                  :key="pid"
                  class="agrid__cell"
                  :class="{
                    'agrid__cell--blank': cellTier(agenda.id, pid) === null,
                    hatched: tierStyle(cellTier(agenda.id, pid)).hatched,
                  }"
                  :style="cellStyle(agenda.id, pid)"
                >
                  <button
                    v-if="cellTier(agenda.id, pid) !== null"
                    type="button"
                    class="agrid__cellbtn"
                    :class="{ 'agrid__cellbtn--on': isPinned(agenda.id, pid) }"
                    :aria-label="`${cellTitle(agenda.id, pid)}. Show details and feedback link.`"
                    @mouseenter="showHover($event.currentTarget, agenda.id, pid)"
                    @mouseleave="scheduleLeave"
                    @focus="showHover($event.currentTarget, agenda.id, pid)"
                    @blur="scheduleLeave"
                    @click="togglePin($event.currentTarget, agenda.id, pid)"
                  >
                    <span class="tabular">{{ cellText(agenda.id, pid) }}</span>
                  </button>
                  <span v-else class="agrid__cellblank">
                    <span class="tabular">{{ cellText(agenda.id, pid) }}</span>
                    <span class="visually-hidden">{{ cellTitle(agenda.id, pid) }}</span>
                  </span>
                </td>
                <td class="agrid__reach tabular">{{ coverageOf(agenda.id) }}/12</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Narrow screens: the same rows as chip lists -->
        <ul class="alist">
          <li v-for="agenda in agendasOfArea(area.name)" :key="agenda.id" class="alist__item">
            <RouterLink class="alist__link" :to="`/agenda/${agenda.id}`" :title="agendaTitle(agenda.id)">
              <span class="tabular agrid__id">{{ agenda.id }}</span> {{ agenda.agenda }}
            </RouterLink>
            <div class="alist__chips">
              <button
                v-for="pid in problemIds.filter((p) => cellTier(agenda.id, p) !== null)"
                :key="pid"
                type="button"
                class="alist__chip"
                :class="{ hatched: tierStyle(cellTier(agenda.id, pid)).hatched, 'alist__chip--on': isPinned(agenda.id, pid) }"
                :style="cellStyle(agenda.id, pid)"
                :aria-label="cellTitle(agenda.id, pid)"
                @mouseenter="showHover($event.currentTarget, agenda.id, pid)"
                @mouseleave="scheduleLeave"
                @focus="showHover($event.currentTarget, agenda.id, pid)"
                @blur="scheduleLeave"
                @click="togglePin($event.currentTarget, agenda.id, pid)"
              >
                <span class="tabular">{{ pid }}</span>
                <span class="alist__chiptier">{{ tierAbbrev(cellTier(agenda.id, pid)) }}</span>
              </button>
            </div>
          </li>
        </ul>
      </div>
    </section>

    <!-- Info + feedback popover, anchored to the hovered or pinned cell. -->
    <Teleport to="body">
      <div
        v-if="popInfo"
        class="pop"
        :class="{ 'pop--above': popInfo.above }"
        :style="{ left: `${popInfo.x}px`, top: `${popInfo.y}px`, width: `${POP_WIDTH}px` }"
        role="dialog"
        aria-label="Cell details"
        @mouseenter="cancelLeave"
        @mouseleave="scheduleLeave"
      >
        <p class="pop__title">
          <span class="tabular">{{ popInfo.agendaId }}</span> {{ popInfo.agendaName }}
          <span class="pop__x">×</span>
          <span class="tabular">{{ popInfo.problemId }}</span> {{ popInfo.problemName }}
        </p>
        <p class="pop__tier">
          <span
            class="pop__chip"
            :class="{ hatched: tierStyle(popInfo.tier).hatched }"
            :style="{ backgroundColor: tierStyle(popInfo.tier).fill, color: tierStyle(popInfo.tier).ink }"
            >{{ popInfo.tier }}</span
          >
          <span>{{ popInfo.problemDef }}</span>
        </p>
        <RouterLink class="pop__fb" :to="popInfo.to" @click="closeAll">
          Give feedback on {{ popInfo.agendaId }} →
        </RouterLink>
      </div>
    </Teleport>
  </div>
</template>


<style scoped>
.full__head {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.full__title {
  margin-bottom: 0.25rem;
}

.full__sub {
  color: var(--ink-secondary);
  font-size: 0.9375rem;
  margin: 0;
}

.full__controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.band {
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  background: var(--surface);
}

.band + .band {
  margin-top: 0.45rem;
}

.band__headwrap {
  margin: 0;
  font-size: inherit;
}

.band__head {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.7rem 0.9rem;
  background: none;
  border: 0;
  cursor: pointer;
  text-align: left;
}

.band__tag {
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--ink-muted);
  min-width: 1.75rem;
  flex: none;
}

.band__name {
  flex: 1 1 auto;
  font-weight: 600;
  font-size: 0.9375rem;
}

.band__count {
  font-size: 0.75rem;
  color: var(--ink-muted);
  flex: none;
}

.band__caret {
  width: 1rem;
  text-align: center;
  color: var(--ink-muted);
  flex: none;
}

.band__body {
  padding: 0 0.9rem 1rem;
  border-top: 1px solid var(--rule);
}

.band__def {
  font-size: 0.8125rem;
  color: var(--ink-muted);
  margin: 0.75rem 0;
  max-width: var(--measure);
}

.band__scroll {
  display: none;
  overflow-x: auto;
}

.agrid {
  border-collapse: separate;
  border-spacing: 2px;
  width: 100%;
  min-width: 640px;
}

.agrid__corner,
.agrid__colhead {
  font-size: 0.6875rem;
  color: var(--ink-muted);
  font-weight: 600;
  text-align: center;
  padding-bottom: 0.2rem;
}

.agrid__corner {
  text-align: left;
}

.agrid__colhead abbr {
  text-decoration: none;
  border-bottom: 1px dotted var(--rule-strong);
  cursor: help;
}

.agrid__rowhead {
  text-align: left;
  font-weight: 400;
  padding-right: 0.5rem;
  max-width: 17rem;
}

.agrid__rowhead a {
  display: flex;
  gap: 0.4rem;
  align-items: baseline;
  color: inherit;
  text-decoration: none;
  font-size: 0.8125rem;
}

.agrid__rowhead a:hover .agrid__name {
  text-decoration: underline;
}

.agrid__id {
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--ink-muted);
  flex: none;
}

.agrid__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ink-secondary);
}

.agrid__cell {
  width: 2.1rem;
  height: 1.6rem;
  padding: 0;
  text-align: center;
  border-radius: 3px;
  font-size: 0.5625rem;
  font-weight: 700;
}

/* Each rated cell is a button that opens the info popover in place, the same
   "click to open the detail" the condensed grid above has, not a hover-only hint. */
.agrid__cellbtn,
.agrid__cellblank {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border-radius: 3px;
}

.agrid__cellbtn {
  background: none;
  border: 0;
  padding: 0;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.agrid__cellbtn:hover,
.agrid__cellbtn:focus-visible,
.agrid__cellbtn--on {
  outline: 2px solid var(--ink);
  outline-offset: -2px;
}

.agrid__cell--blank {
  border: 1px dashed var(--rule);
}

.agrid__reach {
  font-size: 0.6875rem;
  color: var(--ink-muted);
  text-align: right;
  padding-left: 0.4rem;
  white-space: nowrap;
}

.alist {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.6rem;
}

.alist__item {
  border-top: 1px solid var(--rule);
  padding-top: 0.6rem;
}

.alist__link {
  color: inherit;
  font-size: 0.875rem;
  text-decoration: none;
}

.alist__link:hover {
  text-decoration: underline;
}

.alist__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-top: 0.35rem;
}

.alist__chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border: 0;
  border-radius: 3px;
  padding: 0.1rem 0.35rem;
  font-family: inherit;
  font-size: 0.625rem;
  font-weight: 700;
  cursor: pointer;
}

.alist__chip:hover,
.alist__chip:focus-visible,
.alist__chip--on {
  outline: 2px solid var(--ink);
  outline-offset: 1px;
}

.alist__chiptier {
  opacity: 0.85;
  font-size: 0.5625rem;
}

@media (min-width: 860px) {
  .band__scroll {
    display: block;
  }

  .alist {
    display: none;
  }
}

.pop {
  position: fixed;
  z-index: 61;
  transform: translate(-50%, 0);
  background: var(--surface);
  border: 1px solid var(--rule-strong);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 0.7rem 0.8rem;
  font-size: 0.8125rem;
}

.pop--above {
  transform: translate(-50%, -100%);
}

.pop__title {
  margin: 0 0 0.45rem;
  font-weight: 600;
  color: var(--ink);
  line-height: 1.35;
}

.pop__x {
  color: var(--ink-muted);
  font-weight: 400;
}

.pop__tier {
  margin: 0 0 0.55rem;
  display: flex;
  gap: 0.4rem;
  align-items: baseline;
  color: var(--ink-secondary);
}

.pop__chip {
  flex: none;
  border-radius: 3px;
  padding: 0.05rem 0.35rem;
  font-size: 0.625rem;
  font-weight: 700;
}

.pop__fb {
  font-weight: 500;
}
</style>
