<script setup lang="ts">
/**
 * A modal profile of one research area and the agendas inside it, opened from the
 * methodology page. It reads straight from the curated dataset, so it works with
 * the API asleep like everything else.
 *
 * Behaviour of a well-mannered dialog: opens with focus on the close button,
 * closes on Escape or a backdrop click, and hands focus back to whatever opened
 * it. Teleported to the body so the page's layout can never clip it.
 */
import { nextTick, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { agendasOfArea } from '@/lib/atlas'
import { tierStyle } from '@/lib/scales'
import type { Area } from '@/types/atlas'

const props = defineProps<{ area: Area | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const closeButton = ref<HTMLButtonElement | null>(null)
let opener: HTMLElement | null = null

watch(
  () => props.area,
  async (area) => {
    if (area) {
      opener = document.activeElement as HTMLElement | null
      await nextTick()
      closeButton.value?.focus()
    } else {
      opener?.focus?.()
      opener = null
    }
  },
)

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="area" class="ov" @click.self="emit('close')" @keydown="onKeydown">
      <div
        class="dlg"
        role="dialog"
        aria-modal="true"
        :aria-label="`${area.name}: agendas`"
      >
        <header class="dlg__head">
          <div>
            <p class="dlg__eyebrow"><span class="tabular">{{ area.tag }}</span> · research area</p>
            <h2 class="dlg__title">{{ area.name }}</h2>
          </div>
          <button ref="closeButton" type="button" class="btn btn--quiet dlg__close" @click="emit('close')">
            Close ✕
          </button>
        </header>

        <p class="dlg__def">{{ area.definition }}</p>

        <p class="dlg__count">
          <strong class="tabular">{{ agendasOfArea(area.name).length }}</strong>
          agenda{{ agendasOfArea(area.name).length === 1 ? '' : 's' }} in this area.
        </p>

        <ul class="agendas">
          <li v-for="ag in agendasOfArea(area.name)" :key="ag.id" class="agenda">
            <div class="agenda__head">
              <span class="agenda__id tabular">{{ ag.id }}</span>
              <span class="agenda__name">{{ ag.agenda }}</span>
              <span
                class="chip"
                :class="{ hatched: tierStyle(ag.overall_maturity).hatched }"
                :style="{ backgroundColor: tierStyle(ag.overall_maturity).fill, color: tierStyle(ag.overall_maturity).ink }"
                :title="`Overall maturity: ${ag.overall_maturity}`"
                >{{ ag.overall_maturity }}</span
              >
            </div>
            <dl class="agenda__facts">
              <div>
                <dt>Mechanism</dt>
                <dd>{{ ag.mechanism }}</dd>
              </div>
              <div>
                <dt>Assumption it rests on</dt>
                <dd>{{ ag.assumption }}</dd>
              </div>
              <div>
                <dt>Key work</dt>
                <dd>{{ ag.key_work }}</dd>
              </div>
            </dl>
            <RouterLink class="agenda__open" :to="`/agenda/${ag.id}`" @click="emit('close')">
              Open {{ ag.id }} in full, and rate it &rarr;
            </RouterLink>
          </li>
        </ul>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.ov {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: clamp(1rem, 5vh, 4rem) 1rem;
  overflow-y: auto;
  background: color-mix(in srgb, var(--ink) 45%, transparent);
}

.dlg {
  width: min(720px, 100%);
  background: var(--surface);
  border: 1px solid var(--rule-strong);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  padding: 1.25rem 1.35rem 1.4rem;
}

.dlg__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.dlg__eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--ink-muted);
  margin: 0 0 0.2rem;
}

.dlg__title {
  margin: 0;
  font-size: 1.25rem;
}

.dlg__close {
  flex: none;
}

.dlg__def {
  color: var(--ink-secondary);
  font-size: 0.9375rem;
  margin: 0.75rem 0 0;
  max-width: none;
}

.dlg__count {
  font-size: 0.8125rem;
  color: var(--ink-muted);
  margin: 0.6rem 0 1rem;
}

.agendas {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}

.agenda {
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  padding: 0.8rem 0.9rem 0.85rem;
  background: var(--surface-sunken);
}

.agenda__head {
  display: flex;
  align-items: baseline;
  gap: 0.55rem;
  flex-wrap: wrap;
}

.agenda__id {
  font-weight: 700;
  font-size: 0.8125rem;
  color: var(--ink-muted);
  flex: none;
}

.agenda__name {
  font-weight: 600;
  font-size: 0.9375rem;
  flex: 1 1 auto;
  min-width: 0;
}

.chip {
  flex: none;
  display: inline-block;
  border-radius: 3px;
  padding: 0.12rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 700;
}

.agenda__facts {
  display: grid;
  gap: 0.5rem;
  margin: 0.7rem 0 0;
}

.agenda__facts dt {
  font-size: 0.625rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--ink-muted);
  font-weight: 600;
}

.agenda__facts dd {
  margin: 0.1rem 0 0;
  font-size: 0.8125rem;
  color: var(--ink-secondary);
  max-width: none;
}

.agenda__open {
  display: inline-block;
  margin-top: 0.7rem;
  font-size: 0.8125rem;
  font-weight: 500;
}
</style>
