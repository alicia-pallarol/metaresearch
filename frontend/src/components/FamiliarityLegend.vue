<script setup>
import { FAMILIARITY_LEVELS } from '../config.js'
import { LEGEND_SWATCHES } from '../scale.js'

// showScale: render the heatmap color swatches alongside the definitions.
defineProps({ showScale: { type: Boolean, default: true } })
</script>

<template>
  <div class="legend" role="group" aria-label="Familiarity scale">
    <span class="legend-title small muted">Familiarity</span>
    <ul class="legend-items">
      <li v-for="(lvl, i) in FAMILIARITY_LEVELS" :key="lvl.value" class="legend-item">
        <span
          v-if="showScale"
          class="swatch"
          :style="{ background: LEGEND_SWATCHES[i].color }"
          aria-hidden="true"
        >{{ lvl.value }}</span>
        <span v-else class="badge" aria-hidden="true">{{ lvl.value }}</span>
        <span class="small">{{ lvl.label }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.legend {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px 18px;
}
.legend-title {
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 600;
}
.legend-items {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 8px 18px;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 7px;
}
.swatch,
.badge {
  width: 22px;
  height: 22px;
  border-radius: 5px;
  display: grid;
  place-items: center;
  font-size: 0.72rem;
  font-weight: 700;
  color: #0b1b33;
  border: 1px solid var(--border);
  flex: none;
}
.swatch:nth-child(1) {
  color: #0b1b33;
}
.badge {
  background: var(--surface-2);
  color: var(--text-secondary);
}
</style>
