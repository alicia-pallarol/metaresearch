<script setup lang="ts" generic="T extends string">
/**
 * A segmented switch: two or three mutually exclusive readings of the same grid.
 *
 * Radio inputs rather than buttons, because that is what this is, one choice out
 * of a named set, and it gives arrow-key movement and a group label for free.
 */
const props = defineProps<{
  label: string
  options: ReadonlyArray<{ value: T; label: string; hint?: string }>
  /** Unique within the page: it is the radio group's name attribute. */
  name: string
}>()

const model = defineModel<T>({ required: true })

const groupId = `switch-${props.name}`
</script>

<template>
  <div class="sw" role="group" :aria-labelledby="groupId">
    <span :id="groupId" class="sw__label">{{ label }}</span>
    <div class="sw__opts">
      <label v-for="opt in options" :key="opt.value" class="sw__opt" :class="{ 'sw__opt--on': model === opt.value }">
        <input v-model="model" type="radio" :name="name" :value="opt.value" class="visually-hidden" />
        <span>{{ opt.label }}</span>
      </label>
    </div>
  </div>
</template>

<style scoped>
.sw {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  flex-wrap: wrap;
}

.sw__label {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  font-weight: 600;
  color: var(--ink-muted);
}

.sw__opts {
  display: flex;
  border: 1px solid var(--rule-strong);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--surface);
}

.sw__opt {
  font-size: 0.8125rem;
  font-weight: 500;
  padding: 0.34rem 0.7rem;
  cursor: pointer;
  color: var(--ink-secondary);
  white-space: nowrap;
  transition: background-color 120ms ease, color 120ms ease;
}

.sw__opt + .sw__opt {
  border-left: 1px solid var(--rule);
}

.sw__opt:hover {
  background: var(--surface-sunken);
}

.sw__opt--on {
  background: var(--ink);
  color: var(--page);
}

.sw__opt--on:hover {
  background: var(--ink);
}

/* The input is visually hidden, so the focus ring has to live on the label. */
.sw__opt:has(input:focus-visible) {
  outline: 2px solid var(--focus);
  outline-offset: -2px;
}
</style>
