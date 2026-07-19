<script setup>
import { computed } from 'vue'
import { LEGEND_SWATCHES } from '../scale.js'

// Vue 3.5 reactive props destructure so `distribution` and `n` stay reactive.
const { distribution, n } = defineProps({
  distribution: { type: Object, required: true }, // { "0":n, "1":n, "2":n, "3":n }
  n: { type: Number, required: true },
})

const segments = computed(() =>
  [0, 1, 2, 3].map((v) => {
    const count = distribution[String(v)] || 0
    return {
      value: v,
      count,
      pct: n > 0 ? (count / n) * 100 : 0,
      color: LEGEND_SWATCHES[v].color,
    }
  })
)
</script>

<template>
  <div class="dist">
    <div
      class="bar"
      role="img"
      :aria-label="`Distribution of ${n} ratings: ` +
        segments.map((s) => `${s.count} rated ${s.value}`).join(', ')"
    >
      <template v-for="s in segments" :key="s.value">
        <span
          v-if="s.count > 0"
          class="seg"
          :style="{ width: s.pct + '%', background: s.color }"
          :title="`${s.count} rated ${s.value}`"
        ></span>
      </template>
    </div>
    <ul class="counts small muted" aria-hidden="true">
      <li v-for="s in segments" :key="s.value">
        <span class="chip" :style="{ background: s.color }"></span>{{ s.value }}:
        <strong>{{ s.count }}</strong>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.dist {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.bar {
  display: flex;
  height: 14px;
  border-radius: 4px;
  overflow: hidden;
  background: var(--nodata);
  border: 1px solid var(--border);
  gap: 2px;
}
.seg {
  min-width: 2px;
  height: 100%;
}
.counts {
  list-style: none;
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  margin: 0;
  padding: 0;
}
.counts li {
  display: flex;
  align-items: center;
  gap: 5px;
}
.chip {
  width: 11px;
  height: 11px;
  border-radius: 3px;
  border: 1px solid var(--border);
}
</style>
