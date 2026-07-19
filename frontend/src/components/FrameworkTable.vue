<script setup>
import { ref, computed } from 'vue'
import { colorForAverage, textOn } from '../scale.js'
import AreaDetail from './AreaDetail.vue'

const props = defineProps({
  areas: { type: Array, required: true },
  lowSample: { type: Number, default: 3 },
})

const expanded = ref(new Set())

function toggle(id) {
  const next = new Set(expanded.value)
  next.has(id) ? next.delete(id) : next.add(id)
  expanded.value = next
}

// Group areas by problem group, preserving incoming (sorted) order.
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
  return a.average === null || a.average === undefined ? '—' : a.average.toFixed(2)
}
function cellStyle(a) {
  return { background: colorForAverage(a.average), color: textOn(a.average) }
}
</script>

<template>
  <div class="table-wrap card">
    <table class="framework">
      <caption class="visually-hidden">
        AI safety research areas grouped by problem, with average community
        familiarity and number of ratings. Activate a row to expand its full
        definition and distribution.
      </caption>
      <thead>
        <tr>
          <th scope="col" class="col-area">Research area</th>
          <th scope="col" class="col-avg">Avg</th>
          <th scope="col" class="col-n">n</th>
          <th scope="col" class="col-toggle"><span class="visually-hidden">Details</span></th>
        </tr>
      </thead>
      <tbody v-for="group in groups" :key="group.name" class="group">
        <tr class="group-row">
          <th :colspan="4" scope="colgroup" class="group-head">{{ group.name }}</th>
        </tr>
        <template v-for="a in group.areas" :key="a.id">
          <tr
            class="area-row"
            :class="{ open: expanded.has(a.id) }"
            tabindex="0"
            role="button"
            :aria-expanded="expanded.has(a.id)"
            :aria-controls="`detail-${a.id}`"
            @click="toggle(a.id)"
            @keydown.enter.prevent="toggle(a.id)"
            @keydown.space.prevent="toggle(a.id)"
          >
            <td class="col-area">
              <span class="tag" aria-hidden="true">{{ a.tag }}</span>
              <span class="area-name">{{ a.name }}</span>
            </td>
            <td class="col-avg">
              <span
                class="avg-cell"
                :style="cellStyle(a)"
                :title="a.n ? `Average ${avgText(a)} from ${a.n} ratings` : 'No ratings yet'"
              >{{ avgText(a) }}</span>
            </td>
            <td class="col-n">
              <span :class="{ low: a.n > 0 && a.n < lowSample }">{{ a.n }}</span>
              <span
                v-if="a.n > 0 && a.n < lowSample"
                class="low-dot"
                :title="`Low sample (n < ${lowSample})`"
                aria-label="low sample"
              >!</span>
            </td>
            <td class="col-toggle">
              <span class="chev" :class="{ open: expanded.has(a.id) }" aria-hidden="true"></span>
            </td>
          </tr>
          <tr v-if="expanded.has(a.id)" class="detail-row">
            <td :colspan="4" :id="`detail-${a.id}`">
              <AreaDetail :area="a" :low-sample="lowSample" />
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap {
  overflow-x: auto;
}
.framework {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
}
thead th {
  text-align: left;
  padding: 12px 14px;
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 64px;
  background: var(--surface-1);
}
.col-avg,
.col-n {
  width: 1%;
  white-space: nowrap;
}
.col-toggle {
  width: 32px;
}
.group-head {
  text-align: left;
  padding: 14px 14px 6px;
  font-family: var(--font-serif);
  font-size: 0.98rem;
  font-weight: 600;
  color: var(--text-primary);
  background: var(--surface-2);
  border-top: 1px solid var(--border);
}
.area-row {
  cursor: pointer;
  border-top: 1px solid var(--border);
}
.area-row:hover {
  background: var(--surface-2);
}
.area-row.open {
  background: var(--surface-2);
}
.area-row td {
  padding: 11px 14px;
  vertical-align: middle;
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
  letter-spacing: 0.03em;
  color: var(--text-secondary);
  vertical-align: middle;
}
.area-name {
  font-weight: 500;
}
.avg-cell {
  display: inline-grid;
  place-items: center;
  min-width: 48px;
  height: 30px;
  padding: 0 8px;
  border-radius: 5px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  border: 1px solid rgba(0, 0, 0, 0.12);
}
.col-n .low {
  color: #9a6a00;
  font-weight: 600;
}
.low-dot {
  display: inline-grid;
  place-items: center;
  width: 16px;
  height: 16px;
  margin-left: 6px;
  border-radius: 50%;
  background: #f0c674;
  color: #4a3400;
  font-size: 0.7rem;
  font-weight: 800;
  vertical-align: middle;
}
.chev {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-right: 2px solid var(--text-muted);
  border-bottom: 2px solid var(--text-muted);
  transform: rotate(-45deg);
  transition: transform 0.15s ease;
}
.chev.open {
  transform: rotate(45deg);
}
.detail-row td {
  padding: 4px 16px 20px;
  background: var(--surface-2);
}
</style>
