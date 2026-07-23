<script setup lang="ts">
/**
 * What opens when a cell of the condensed grid is clicked: the problem, the
 * area, and every agenda in that area that reaches the problem, strongest
 * first, each expandable to its full record.
 *
 * The agenda rows always carry OUR rating, whichever way the grid above is
 * switched: this panel is the record. What the switches change is the reading
 * line, which says what the cell is claiming under them and where that came from.
 */
import { computed, ref, watch } from 'vue'
import { agendasReaching, getAreaByName, getProblem } from '@/lib/atlas'
import { gridCell, SCENARIO_HINTS, type GridSource, type Scenario } from '@/lib/community'
import { tierStyle } from '@/lib/scales'
import { useSummary } from '@/composables/useSummary'
import AgendaAccordion from '@/components/AgendaAccordion.vue'
import FeedbackForm from '@/components/FeedbackForm.vue'
import type { FeedbackSubject } from '@/types/atlas'

const props = defineProps<{
  problemId: string
  areaName: string
  source: GridSource
  scenario: Scenario
}>()
const emit = defineEmits<{ (e: 'close'): void }>()

const { summary } = useSummary()

const problem = computed(() => getProblem(props.problemId))
const area = computed(() => getAreaByName(props.areaName))
const rows = computed(() => agendasReaching(props.areaName, props.problemId))

const cell = computed(() =>
  gridCell(props.source, props.scenario, area.value?.tag ?? '', props.areaName, props.problemId, summary.value),
)

const reading = computed(() => {
  const c = cell.value
  const hint = SCENARIO_HINTS[props.scenario]
  if (props.source === 'research') {
    return `Our rating, taken as ${hint}. The cell in the grid shows this tier only.`
  }
  if (c.tier === null) {
    return 'Nobody has rated this area yet, so the community view of this cell is empty. The agendas below are our reading.'
  }
  const ratings = `${c.ratings} rating${c.ratings === 1 ? '' : 's'} of ${props.areaName}`
  return c.inherited
    ? `What researchers voted, taken as ${hint}, from ${ratings}. No agenda here has been rated on its own yet, so this is the area's own votes.`
    : `What researchers voted, taken as ${hint}, from ${ratings}, refined by the votes on individual agendas.`
})

// Feedback about the whole cell (this area's coverage of this problem), distinct
// from the per-agenda feedback reachable by opening an agenda.
const showCellForm = ref(false)
watch(
  () => [props.areaName, props.problemId],
  () => {
    showCellForm.value = false
  },
)

const cellSubject = computed<FeedbackSubject>(() => ({
  kind: 'cell',
  areaTag: area.value?.tag ?? '',
  areaName: props.areaName,
  problemId: props.problemId,
  problemName: problem.value?.name ?? '',
}))
</script>

<template>
  <section class="panel card" aria-live="polite">
    <header class="panel__head">
      <div>
        <p class="eyebrow">Drill-down</p>
        <h3 class="panel__title">
          {{ area?.name }} <span class="panel__x">×</span> {{ problem?.id }} {{ problem?.name }}
        </h3>
      </div>
      <button type="button" class="btn btn--quiet" @click="emit('close')">Close ✕</button>
    </header>

    <div class="panel__defs">
      <p class="panel__def">
        <strong>{{ problem?.id }}.</strong> {{ problem?.short_def }}
      </p>
      <p class="panel__def panel__def--muted">
        <strong>{{ area?.name }}.</strong> {{ area?.definition }}
      </p>
    </div>

    <p v-if="rows.length === 0" class="panel__empty">
      No agenda in this area is rated against this problem. A blank cell is a claim of no meaningful
      relationship, not an oversight, though a missing row is exactly the kind of thing the feedback
      form exists to catch.
    </p>

    <template v-else>
      <p class="panel__count">
        <span
          v-if="cell.tier"
          class="panel__bestchip"
          :class="{ hatched: tierStyle(cell.tier).hatched }"
          :style="{ backgroundColor: tierStyle(cell.tier).fill, color: tierStyle(cell.tier).ink }"
          >{{ cell.tier }}</span
        >
        <span v-else class="panel__bestchip panel__bestchip--none">Not rated yet</span>
        across <strong>{{ rows.length }}</strong> agenda{{ rows.length === 1 ? '' : 's' }} here.
        {{ reading }}
      </p>

      <p class="panel__feedback">
        Expand a row to read it, or <strong>open any agenda to tell us where its rating is wrong</strong>.
        To rate this area yourself, and the agendas in this cell, use the box below.
      </p>

      <AgendaAccordion
        v-for="row in rows"
        :key="row.agenda.id"
        :agenda="row.agenda"
        :tier="row.tier"
        :highlight-problem-id="problemId"
      />
    </template>

    <div class="panel__cellfb">
      <button
        v-if="!showCellForm"
        type="button"
        class="btn"
        :aria-expanded="false"
        @click="showCellForm = true"
      >
        Rate {{ area?.name }} yourself, and see what others said →
      </button>
      <FeedbackForm v-else :subject="cellSubject" />
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 1.15rem 1.25rem 1.35rem;
  margin-top: 1.25rem;
}

.panel__head {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  justify-content: space-between;
}

.panel__title {
  margin-bottom: 0.75rem;
  font-size: 1.05rem;
}

.panel__x {
  color: var(--ink-muted);
  font-weight: 400;
}

.panel__defs {
  display: grid;
  gap: 0.35rem;
  margin-bottom: 1rem;
}

.panel__def {
  font-size: 0.875rem;
  color: var(--ink-secondary);
  margin: 0;
  max-width: var(--measure);
}

.panel__def--muted {
  color: var(--ink-muted);
}

.panel__count {
  font-size: 0.875rem;
  color: var(--ink-secondary);
  margin-bottom: 0.75rem;
  max-width: var(--measure);
}

.panel__bestchip {
  display: inline-block;
  font-size: 0.6875rem;
  font-weight: 700;
  padding: 0.1rem 0.45rem;
  border-radius: 3px;
  margin-right: 0.15rem;
}

.panel__bestchip--none {
  border: 1px dotted var(--rule-strong);
  color: var(--ink-muted);
}

.panel__empty {
  font-size: 0.875rem;
  color: var(--ink-secondary);
  background: var(--surface-sunken);
  border-radius: var(--radius);
  padding: 0.85rem 1rem;
  max-width: var(--measure);
  margin: 0;
}

.panel__feedback {
  font-size: 0.8125rem;
  color: var(--ink-secondary);
  background: var(--surface-sunken);
  border-radius: var(--radius);
  padding: 0.6rem 0.85rem;
  margin: 0 0 0.9rem;
  max-width: var(--measure);
}

.panel__feedback strong {
  color: var(--ink);
}

.panel__cellfb {
  margin-top: 1.1rem;
  padding-top: 1.1rem;
  border-top: 1px solid var(--rule);
}
</style>
