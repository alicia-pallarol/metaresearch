<script setup>
import { ref, computed } from 'vue'
import { colorForAverage, textOn } from '../scale.js'
import AreaDetail from './AreaDetail.vue'

const props = defineProps({
  areas: { type: Array, required: true },
  lowSample: { type: Number, default: 3 },
})

const selectedId = ref(null)
const selected = computed(() => props.areas.find((a) => a.id === selectedId.value) || null)

function select(a) {
  selectedId.value = selectedId.value === a.id ? null : a.id
}

const groups = computed(() => {
  const out = []
  let current = null
  for (const a of props.areas) {
    if (!current || current.name !== a.problem_group) {
      current = { name: a.problem_group, areas: [] }
      out.push(current)
    }
    current.areas.push(a)
  }
  return out
})

function avgText(a) {
  return a.average === null || a.average === undefined ? '—' : a.average.toFixed(1)
}
</script>

<template>
  <div class="heatmap">
    <div v-for="group in groups" :key="group.name" class="hm-group">
      <h3 class="hm-group-title">{{ group.name }}</h3>
      <div class="hm-cells">
        <button
          v-for="a in group.areas"
          :key="a.id"
          class="hm-cell"
          :class="{ selected: selectedId === a.id }"
          :style="{ background: colorForAverage(a.average), color: textOn(a.average) }"
          :aria-pressed="selectedId === a.id"
          :title="a.n ? `${a.name}: avg ${avgText(a)} (n=${a.n})` : `${a.name}: no ratings yet`"
          @click="select(a)"
        >
          <span class="hm-tag">{{ a.tag }}</span>
          <span class="hm-avg">{{ avgText(a) }}</span>
          <span class="hm-n">n={{ a.n }}</span>
          <span
            v-if="a.n > 0 && a.n < lowSample"
            class="hm-low"
            aria-label="low sample"
          >!</span>
        </button>
      </div>
    </div>

    <div v-if="selected" class="hm-detail card" role="region" :aria-label="`Details for ${selected.name}`">
      <div class="hm-detail-head">
        <div>
          <span class="tag">{{ selected.tag }}</span>
          <strong>{{ selected.name }}</strong>
        </div>
        <button class="btn small close" @click="selectedId = null" aria-label="Close details">Close</button>
      </div>
      <AreaDetail :area="selected" :low-sample="lowSample" />
    </div>
    <p v-else class="small muted select-hint">Select a cell to see its definition and distribution.</p>
  </div>
</template>

<style scoped>
.heatmap {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.hm-group-title {
  margin: 0 0 8px;
  font-size: 0.95rem;
  color: var(--text-secondary);
}
.hm-cells {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 8px;
}
.hm-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  padding: 10px 12px;
  min-height: 78px;
  border-radius: 7px;
  border: 1px solid rgba(0, 0, 0, 0.12);
  cursor: pointer;
  text-align: left;
  transition: transform 0.1s ease, box-shadow 0.1s ease;
}
.hm-cell:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow);
}
.hm-cell.selected {
  outline: 3px solid var(--focus);
  outline-offset: 1px;
}
.hm-tag {
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  opacity: 0.85;
}
.hm-avg {
  font-family: var(--font-serif);
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1;
}
.hm-n {
  font-size: 0.72rem;
  opacity: 0.85;
  font-variant-numeric: tabular-nums;
}
.hm-low {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 16px;
  height: 16px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: #f0c674;
  color: #4a3400;
  font-size: 0.7rem;
  font-weight: 800;
}
.hm-detail {
  padding: 18px;
}
.hm-detail-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}
.tag {
  display: inline-grid;
  place-items: center;
  min-width: 30px;
  height: 22px;
  padding: 0 6px;
  margin-right: 10px;
  border-radius: 4px;
  background: var(--surface-0);
  border: 1px solid var(--border-strong);
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-secondary);
  vertical-align: middle;
}
.select-hint {
  margin: 0;
}
</style>
