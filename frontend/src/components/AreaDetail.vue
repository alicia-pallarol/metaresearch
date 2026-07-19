<script setup>
import { computed } from 'vue'
import DistributionBar from './DistributionBar.vue'

// Vue 3.5 reactive props destructure: `area` and `lowSample` stay reactive.
const { area, lowSample } = defineProps({
  area: { type: Object, required: true },
  lowSample: { type: Number, default: 3 },
})

const avgText = computed(() =>
  area.average === null || area.average === undefined ? '—' : area.average.toFixed(2)
)
const isLow = computed(() => area.n > 0 && area.n < lowSample)
</script>

<template>
  <div class="detail">
    <p class="def">{{ area.definition }}</p>

    <div class="stats">
      <div class="stat">
        <span class="stat-num">{{ avgText }}</span>
        <span class="small muted">Average familiarity</span>
      </div>
      <div class="stat">
        <span class="stat-num">{{ area.n }}</span>
        <span class="small muted">Ratings (n)</span>
      </div>
    </div>

    <p v-if="area.n === 0" class="small muted no-data">No ratings yet for this area.</p>
    <template v-else>
      <p v-if="isLow" class="small low-flag">
        Low sample (n &lt; {{ lowSample }}) — interpret with caution.
      </p>
      <DistributionBar :distribution="area.distribution" :n="area.n" />
    </template>
  </div>
</template>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.def {
  margin: 0;
  color: var(--text-secondary);
  max-width: 70ch;
}
.stats {
  display: flex;
  gap: 28px;
}
.stat {
  display: flex;
  flex-direction: column;
}
.stat-num {
  font-family: var(--font-serif);
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.1;
}
.low-flag {
  color: #9a6a00;
  margin: 0;
}
@media (prefers-color-scheme: dark) {
  .low-flag {
    color: #e0b050;
  }
}
.no-data {
  margin: 0;
}
</style>
