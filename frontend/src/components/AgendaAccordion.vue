<script setup lang="ts">
/**
 * One agenda as a collapsible row: the header carries its tier against the
 * problem in view, the body carries the full record and the agenda's whole row
 * of per-problem tiers as chips.
 */
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { problemsReachedBy } from '@/lib/atlas'
import { tierAbbrev, tierStyle } from '@/lib/scales'
import type { Agenda, Tier } from '@/types/atlas'

const props = defineProps<{
  agenda: Agenda
  /** Tier against the problem currently in view, when there is one. */
  tier?: Tier | null
  /** Highlight this problem's chip in the expanded body. */
  highlightProblemId?: string
}>()

const open = ref(false)
const bodyId = computed(() => `agenda-body-${props.agenda.id}`)
const reached = computed(() => problemsReachedBy(props.agenda.id))
</script>

<template>
  <div class="row" :class="{ 'row--open': open }">
    <h4 class="row__headwrap">
      <button
        type="button"
        class="row__head"
        :aria-expanded="open"
        :aria-controls="bodyId"
        @click="open = !open"
      >
        <span
          v-if="tier"
          class="chip"
          :class="{ hatched: tierStyle(tier).hatched }"
          :style="{ backgroundColor: tierStyle(tier).fill, color: tierStyle(tier).ink }"
          :title="tier"
          >{{ tierAbbrev(tier) }}</span
        >
        <span class="row__id tabular">{{ agenda.id }}</span>
        <span class="row__name">{{ agenda.agenda }}</span>
        <span class="row__tier">{{ tier ?? agenda.overall_maturity }}</span>
        <span class="row__caret" aria-hidden="true">{{ open ? '−' : '+' }}</span>
      </button>
      <RouterLink
        class="row__open"
        :to="`/agenda/${agenda.id}`"
        :title="`Open ${agenda.id} to read it and give feedback`"
      >
        <span class="row__openfull">Open &amp; give feedback →</span>
        <span class="row__openshort">Open →</span>
      </RouterLink>
    </h4>

    <Transition name="drop">
      <div v-show="open" :id="bodyId" class="row__body">
        <dl class="facts">
          <div class="fact">
            <dt>Mechanism</dt>
            <dd>{{ agenda.mechanism }}</dd>
          </div>
          <div class="fact">
            <dt>Assumption it rests on</dt>
            <dd>{{ agenda.assumption }}</dd>
          </div>
          <div class="fact">
            <dt>Overall maturity</dt>
            <dd>{{ agenda.overall_maturity }}</dd>
          </div>
          <div class="fact">
            <dt>Key work</dt>
            <dd>{{ agenda.key_work }}</dd>
          </div>
        </dl>

        <div class="chips">
          <span class="chips__label">Across all problems</span>
          <span
            v-for="row in reached"
            :key="row.problem.id"
            class="chip chip--wide"
            :class="{
              hatched: tierStyle(row.tier).hatched,
              'chip--current': row.problem.id === highlightProblemId,
            }"
            :style="{ backgroundColor: tierStyle(row.tier).fill, color: tierStyle(row.tier).ink }"
            :title="`${row.problem.id} · ${row.problem.name}, ${row.tier}`"
          >
            <span class="tabular">{{ row.problem.id }}</span>
            <span class="chip__tier">{{ tierAbbrev(row.tier) }}</span>
          </span>
        </div>

        <div class="row__foot">
          <RouterLink class="btn btn--primary btn--sm" :to="`/agenda/${agenda.id}`">
            Rate {{ agenda.id }} or flag this rating →
          </RouterLink>
          <span class="row__hint">Opens the agenda, where the feedback form is.</span>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.row {
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  background: var(--surface);
}

.row + .row {
  margin-top: 0.4rem;
}

.row--open {
  border-color: var(--rule-strong);
}

.row__headwrap {
  margin: 0;
  font-size: inherit;
  font-weight: inherit;
  display: flex;
  align-items: stretch;
}

.row__head {
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 0.75rem;
  background: none;
  border: 0;
  cursor: pointer;
  text-align: left;
}

.row__open {
  flex: none;
  display: flex;
  align-items: center;
  padding: 0.6rem 0.85rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--link);
  text-decoration: none;
  border-left: 1px solid var(--rule);
  white-space: nowrap;
}

.row__open:hover {
  background: var(--surface-sunken);
  text-decoration: underline;
}

.row__openshort {
  display: none;
}

@media (max-width: 560px) {
  .row__openfull {
    display: none;
  }
  .row__openshort {
    display: inline;
  }
}

.chip {
  flex: none;
  display: inline-grid;
  place-items: center;
  min-width: 1.9rem;
  height: 1.25rem;
  border-radius: 3px;
  font-size: 0.625rem;
  font-weight: 700;
  padding-inline: 0.3rem;
}

.chip--wide {
  display: inline-flex;
  gap: 0.3rem;
  align-items: center;
  min-width: auto;
  font-size: 0.6875rem;
}

.chip--current {
  outline: 2px solid var(--ink);
  outline-offset: 1px;
}

.chip__tier {
  opacity: 0.85;
  font-size: 0.5625rem;
}

.row__id {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--ink-muted);
  flex: none;
}

.row__name {
  flex: 1 1 auto;
  font-size: 0.9375rem;
  font-weight: 500;
  min-width: 0;
}

.row__tier {
  font-size: 0.75rem;
  color: var(--ink-secondary);
  flex: none;
  display: none;
}

.row__caret {
  flex: none;
  color: var(--ink-muted);
  font-size: 1rem;
  width: 1rem;
  text-align: center;
}

.row__body {
  padding: 0 0.85rem 0.9rem;
  border-top: 1px solid var(--rule);
  overflow: hidden;
}

.facts {
  margin: 0.85rem 0 0;
  display: grid;
  gap: 0.7rem;
}

.fact dt {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--ink-muted);
  font-weight: 600;
}

.fact dd {
  margin: 0.15rem 0 0;
  font-size: 0.875rem;
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.fact--split {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.7rem;
}

.chips {
  margin-top: 0.9rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  align-items: center;
}

.chips__label {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--ink-muted);
  font-weight: 600;
  margin-right: 0.25rem;
}

.row__foot {
  margin-top: 0.9rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.75rem;
  align-items: center;
  justify-content: space-between;
}

.row__hint {
  font-size: 0.8125rem;
  color: var(--ink-muted);
}

@media (min-width: 720px) {
  .row__tier {
    display: block;
  }
}

.drop-enter-active,
.drop-leave-active {
  transition: opacity 160ms ease, transform 160ms ease;
}

.drop-enter-from,
.drop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
