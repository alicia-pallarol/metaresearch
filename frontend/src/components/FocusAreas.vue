<script setup lang="ts">
/**
 * "Focus areas", where the next unit of effort pays off most, computed from the
 * grid rather than hand-written. Two independent passes (research areas, then
 * problems), each sorted into the three focus categories.
 *
 * The Research / Community switch here is the SAME state as the grid's, lifted to
 * HomeView, so this section and the grid always show the same source. The whole
 * classification lives in lib/focus.ts, this file only renders it.
 */
import { computed } from 'vue'
import { FOCUS_CATEGORIES, focusAreas, focusProblems, type FocusBoard, type FocusUnit } from '@/lib/focus'
import { useSummary } from '@/composables/useSummary'
import ModeSwitch from '@/components/ModeSwitch.vue'
import type { GridSource } from '@/lib/community'

const emit = defineEmits<{
  (e: 'focus-problem', problemId: string): void
  (e: 'focus-area', areaTag: string): void
}>()

const source = defineModel<GridSource>('source', { required: true })

const { summary, state } = useSummary()

const blocks = computed<Array<{ unit: FocusUnit; heading: string; board: FocusBoard }>>(() => [
  { unit: 'area', heading: 'Research areas', board: focusAreas(source.value, summary.value) },
  { unit: 'problem', heading: 'Problems', board: focusProblems(source.value, summary.value) },
])

function isEmpty(board: FocusBoard): boolean {
  return FOCUS_CATEGORIES.every((c) => board[c.id].length === 0)
}

/** Why a block is empty, said plainly rather than shown blank. */
const emptyNote = computed(() => {
  if (source.value !== 'community') return 'Nothing stands out here.'
  if (state.value === 'disabled') return 'Feedback is switched off on this deployment, so there are no community votes to rank.'
  if (state.value === 'unavailable') return 'The feedback server is asleep, so the community ranking is empty. Switch to “Our research”.'
  return 'No community votes yet. This fills in as researchers rate the maturity of areas.'
})

const SOURCE_OPTIONS = [
  { value: 'research' as const, label: 'Our research' },
  { value: 'community' as const, label: 'The community' },
]

function onShow(unit: FocusUnit, id: string): void {
  if (unit === 'problem') emit('focus-problem', id)
  else emit('focus-area', id)
}
</script>

<template>
  <div class="focus">
    <div class="focus__head">
      <p class="intro">
        Not a ranking of importance, a reading of where effort goes furthest right now, sorted from
        the {{ source === 'research' ? 'ratings we gave' : 'maturity researchers voted' }}. Areas and
        problems already rated near-Robust are left out: they are the closest to solved.
      </p>
      <ModeSwitch v-model="source" name="focus-source" label="Rated by" :options="SOURCE_OPTIONS" />
    </div>

    <section v-for="block in blocks" :key="block.unit" class="block">
      <h3 class="block__title">{{ block.heading }}</h3>

      <p v-if="isEmpty(block.board)" class="block__empty">{{ emptyNote }}</p>

      <template v-else>
        <div v-for="cat in FOCUS_CATEGORIES" :key="cat.id" class="cat" :class="`cat--${cat.id}`">
          <template v-if="block.board[cat.id].length">
            <div class="cat__head">
              <h4 class="cat__title">
                <span class="cat__dot" aria-hidden="true"></span>{{ cat.label }}
              </h4>
              <p class="cat__blurb">{{ cat.blurb(block.unit) }}</p>
            </div>

            <ul class="items">
              <li v-for="item in block.board[cat.id]" :key="item.id" class="item">
                <div class="item__row">
                  <span class="item__id tabular">{{ item.id }}</span>
                  <span class="item__name">{{ item.name }}</span>
                  <button type="button" class="item__show" @click="onShow(block.unit, item.id)">
                    Show in grid ↑
                  </button>
                </div>
                <p class="item__reason">{{ item.reason }}</p>
              </li>
            </ul>
          </template>
        </div>
      </template>
    </section>

    <p class="focus__note">
      How each item is placed, the signals, the thresholds, the exclusions, is spelled out and
      tunable in <code>frontend/src/lib/focus.ts</code>.
    </p>
  </div>
</template>

<style scoped>
.focus__head {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem 1.5rem;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1.75rem;
}

.intro {
  color: var(--ink-secondary);
  max-width: var(--measure);
  margin: 0;
  flex: 1 1 24rem;
}

.block + .block {
  margin-top: 2.25rem;
}

.block__title {
  font-size: 1.05rem;
  margin-bottom: 1rem;
  padding-bottom: 0.4rem;
  border-bottom: 1px solid var(--rule);
}

.block__empty {
  color: var(--ink-muted);
  font-size: 0.9375rem;
  background: var(--surface-sunken);
  border-radius: var(--radius);
  padding: 0.85rem 1rem;
  max-width: var(--measure);
}

.cat + .cat {
  margin-top: 1.5rem;
}

.cat__head {
  margin-bottom: 0.7rem;
}

.cat__title {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.9375rem;
  margin-bottom: 0.2rem;
}

.cat__dot {
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 2px;
  flex: none;
  background: var(--ink-muted);
}

/* A sober cue per category, not the maturity ramp, so the two are never confused. */
.cat--easy .cat__dot {
  background: var(--tier-early);
}
.cat--bottleneck .cat__dot {
  background: var(--tier-contested);
}
.cat--underexplored .cat__dot {
  border: 1px dotted var(--rule-strong);
  background: transparent;
}

.cat__blurb {
  color: var(--ink-muted);
  font-size: 0.8125rem;
  margin: 0;
  max-width: var(--measure);
}

.items {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.4rem;
}

.item {
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  background: var(--surface);
  padding: 0.6rem 0.8rem 0.7rem;
}

.item__row {
  display: flex;
  align-items: baseline;
  gap: 0.6rem;
}

.item__id {
  font-weight: 700;
  font-size: 0.8125rem;
  color: var(--ink-muted);
  flex: none;
  min-width: 2rem;
}

.item__name {
  flex: 1 1 auto;
  font-weight: 600;
  font-size: 0.9375rem;
  min-width: 0;
}

.item__show {
  flex: none;
  background: none;
  border: 0;
  padding: 0;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--link);
  text-decoration: underline;
  text-underline-offset: 2px;
  white-space: nowrap;
}

.item__reason {
  margin: 0.3rem 0 0;
  font-size: 0.8125rem;
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.focus__note {
  margin-top: 2rem;
  font-size: 0.8125rem;
  color: var(--ink-muted);
  max-width: var(--measure);
}
</style>
