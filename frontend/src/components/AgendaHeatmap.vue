<script setup lang="ts">
/**
 * The full agenda-level grid: 58 agendas × 12 problems, grouped into collapsible
 * bands by research area.
 *
 * On wide screens each band is a grid. On narrow screens the same data is a list
 * of agendas with their tiers as wrapped chips, a 12-column table on a phone is
 * a scroll maze, and the chips carry the same information.
 */
import { ref } from 'vue'
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
                  :title="cellTitle(agenda.id, pid)"
                >
                  <span class="tabular">{{ cellText(agenda.id, pid) }}</span>
                  <span class="visually-hidden">{{ cellTitle(agenda.id, pid) }}</span>
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
              <span
                v-for="pid in problemIds.filter((p) => cellTier(agenda.id, p) !== null)"
                :key="pid"
                class="alist__chip"
                :class="{ hatched: tierStyle(cellTier(agenda.id, pid)).hatched }"
                :style="cellStyle(agenda.id, pid)"
                :title="cellTitle(agenda.id, pid)"
              >
                <span class="tabular">{{ pid }}</span>
                <span class="alist__chiptier">{{ tierAbbrev(cellTier(agenda.id, pid)) }}</span>
              </span>
            </div>
          </li>
        </ul>
      </div>
    </section>
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
  text-align: center;
  border-radius: 3px;
  font-size: 0.5625rem;
  font-weight: 700;
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
  border-radius: 3px;
  padding: 0.1rem 0.35rem;
  font-size: 0.625rem;
  font-weight: 700;
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
</style>
