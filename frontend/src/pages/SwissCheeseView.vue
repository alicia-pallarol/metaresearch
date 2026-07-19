<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAggregates } from '../composables/useAggregates.js'
import { LAYERS, AI_DISCLAIMER, FURTHER_READING } from '../data/swissCheese.js'
import { colorForAverage, textOn } from '../scale.js'
import LoadingState from '../components/LoadingState.vue'

const { state, load } = useAggregates()
onMounted(load)

const areaByTag = computed(() => {
  const m = {}
  for (const a of state.data?.areas || []) m[a.tag] = a
  return m
})

// Per-layer aggregate familiarity = mean of member areas' averages that have
// at least one rating. n = total ratings across the layer's areas.
const layers = computed(() =>
  LAYERS.map((l) => {
    const members = l.tags.map((t) => areaByTag.value[t]).filter(Boolean)
    const rated = members.filter((a) => a.average !== null && a.average !== undefined)
    const avg = rated.length
      ? rated.reduce((s, a) => s + a.average, 0) / rated.length
      : null
    const n = members.reduce((s, a) => s + (a.n || 0), 0)
    return { ...l, members, avg, n }
  })
)

const selectedId = ref(LAYERS[0].id)
const selected = computed(() => layers.value.find((l) => l.id === selectedId.value) || layers.value[0])

// --- SVG geometry ---
// Each layer is a labelled caption ABOVE a "cheese" bar (labels are never drawn
// over the holes). Hole size grows as familiarity falls, so a well-known layer
// reads as solid cheese and a thinly-known one as porous — the same signal the
// colour carries, doubled for accessibility.
const W = 720
const captionH = 22
const barH = 46
const rowGap = 22
const padTop = 12
const rowH = captionH + barH + rowGap
const H = padTop + LAYERS.length * rowH
const alignedX = 300 // the column the threat descends through
const holeXs = [120, alignedX, 470, 600]

function barY(i) {
  return padTop + i * rowH + captionH
}
function barCenter(i) {
  return barY(i) + barH / 2
}
function captionY(i) {
  return padTop + i * rowH + 15
}
// Radius grows with the "porosity" of a layer: solid (avg 3) -> small holes,
// unknown (no data) -> largest holes.
function holeR(avg) {
  const p = avg === null || avg === undefined ? 1 : Math.max(0, Math.min(1, (3 - avg) / 3))
  return 5 + p * 12
}

const threat = ref(0) // 0..1 progress down the stack
const threatY = computed(() => padTop + threat.value * (H - padTop))
function avgText(v) {
  return v === null || v === undefined ? 'no data' : v.toFixed(2)
}
</script>

<template>
  <div class="container page">
    <section class="intro">
      <h1>A Swiss cheese defense model</h1>
      <p class="lede">
        Safety as layered defenses: each layer catches failures the previous one
        missed, and harm requires holes in every layer to align. Here each layer is
        shaded by the community's aggregate familiarity with the research areas that
        reinforce it — a paler layer signals thinner collective expertise.
      </p>
      <div class="notice notice-warn disclaimer" role="note">
        <strong>AI-assisted content.</strong> {{ AI_DISCLAIMER }}
      </div>
    </section>

    <LoadingState
      v-if="state.loading && !state.loaded"
      :waking="state.waking"
      :attempt="state.wakeAttempt"
      :total="state.wakeTotal"
    />

    <div v-else class="model">
      <div class="stage card">
        <svg
          :viewBox="`0 0 ${W} ${H}`"
          class="cheese"
          role="group"
          aria-label="Layered Swiss cheese defense model. Select a layer for details."
        >
          <defs>
            <mask v-for="(l, i) in layers" :key="'m' + l.id" :id="`holes-${l.id}`">
              <rect x="0" :y="barY(i)" :width="W" :height="barH" rx="10" fill="white" />
              <circle
                v-for="(x, k) in holeXs"
                :key="k"
                :cx="x"
                :cy="barCenter(i)"
                :r="holeR(l.avg)"
                fill="black"
              />
            </mask>
          </defs>

          <g
            v-for="(l, i) in layers"
            :key="l.id"
            class="slice"
            :class="{ selected: selectedId === l.id, empty: l.avg === null }"
            role="button"
            tabindex="0"
            :aria-pressed="selectedId === l.id"
            :aria-label="`Layer ${l.id}: ${l.title}. Familiarity ${avgText(l.avg)}, ${l.n} ratings. Select for details.`"
            @click="selectedId = l.id"
            @keydown.enter.prevent="selectedId = l.id"
            @keydown.space.prevent="selectedId = l.id"
          >
            <!-- caption above the cheese, never over the holes -->
            <text x="2" :y="captionY(i)" class="slice-label">{{ l.id }}. {{ l.title }}</text>
            <text :x="W - 2" :y="captionY(i)" text-anchor="end" class="slice-meta">
              familiarity {{ avgText(l.avg) }} · n={{ l.n }}
            </text>
            <!-- transparent hit area so the whole row is clickable -->
            <rect x="0" :y="barY(i) - captionH" :width="W" :height="barH + captionH" fill="transparent" />
            <rect
              x="0"
              :y="barY(i)"
              :width="W"
              :height="barH"
              rx="10"
              :fill="l.avg === null ? 'var(--nodata)' : colorForAverage(l.avg)"
              :mask="`url(#holes-${l.id})`"
              class="slice-rect"
            />
            <rect x="0" :y="barY(i)" :width="W" :height="barH" rx="10" fill="none" class="slice-outline" />
          </g>

          <!-- threat trajectory drawn on top of the stack -->
          <g aria-hidden="true" class="threat">
            <text :x="alignedX + 10" y="10" class="threat-label">Threat</text>
            <line :x1="alignedX" y1="0" :x2="alignedX" :y2="threatY" class="threat-line" />
            <polygon
              :points="`${alignedX - 6},${threatY - 9} ${alignedX + 6},${threatY - 9} ${alignedX},${threatY + 3}`"
              class="threat-head"
            />
          </g>
        </svg>

        <div class="reading-key small muted">
          <span><span class="key-swatch solid"></span>Darker &amp; more solid = more community familiarity</span>
          <span><span class="key-swatch porous"></span>Paler &amp; more holes = thinner expertise (bigger gaps)</span>
        </div>

        <div class="threat-control">
          <label for="threat" class="small muted">Threat trajectory — drag to descend</label>
          <input id="threat" v-model.number="threat" type="range" min="0" max="1" step="0.01" />
          <p class="small muted threat-note">
            A hazard only causes harm if it finds a hole in <em>every</em> layer.
            Layers with little collective expertise (paler, more porous) are the ones
            most likely to let it through.
          </p>
        </div>
      </div>

      <aside class="panel card" aria-live="polite">
        <span class="layer-stage small muted">{{ selected.stage }}</span>
        <h2>{{ selected.id }}. {{ selected.title }}</h2>
        <p class="hazard">{{ selected.hazard }}</p>

        <div class="layer-fam">
          <span
            class="fam-chip"
            :style="{
              background: selected.avg === null ? 'var(--nodata)' : colorForAverage(selected.avg),
              color: selected.avg === null ? 'var(--text-secondary)' : textOn(selected.avg),
            }"
          >{{ avgText(selected.avg) }}</span>
          <span class="small muted">aggregate familiarity · {{ selected.n }} ratings</span>
        </div>

        <h3 class="sub">Research areas reinforcing this layer</h3>
        <ul class="areas">
          <li v-for="a in selected.members" :key="a.tag">
            <span class="tag">{{ a.tag }}</span>
            <span class="a-name">{{ a.name }}</span>
            <span
              class="a-avg"
              :style="{
                background: a.average === null ? 'var(--nodata)' : colorForAverage(a.average),
                color: a.average === null ? 'var(--text-secondary)' : textOn(a.average),
              }"
            >{{ a.average === null ? '—' : a.average.toFixed(1) }}</span>
          </li>
        </ul>

        <h3 class="sub">Typical holes</h3>
        <ul class="holes-list">
          <li v-for="(h, i) in selected.holes" :key="i">{{ h }}</li>
        </ul>
        <p v-if="selected.note" class="small muted note">{{ selected.note }}</p>
      </aside>
    </div>

    <section class="reading card">
      <h3>Further reading</h3>
      <ul>
        <li v-for="(r, i) in FURTHER_READING" :key="i">{{ r }}</li>
      </ul>
      <p class="small muted caveat">
        Familiarity is a proxy for attention, not for a layer's actual effectiveness:
        a layer can be well-studied and still weak, or vice versa. The layer groupings
        are this model's own construction and are intended to be revised by the
        research team.
      </p>
    </section>
  </div>
</template>

<style scoped>
.page {
  padding-top: 32px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.intro {
  max-width: 78ch;
}
.intro h1 {
  font-size: clamp(1.5rem, 3.2vw, 2.1rem);
  margin: 0 0 10px;
}
.lede {
  font-size: 1.05rem;
  color: var(--text-secondary);
  margin: 0 0 14px;
}
.disclaimer {
  max-width: 80ch;
}
.model {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(280px, 1fr);
  gap: 18px;
  align-items: start;
}
.stage {
  padding: 16px;
}
.cheese {
  width: 100%;
  height: auto;
  display: block;
}
.slice {
  cursor: pointer;
}
.slice-rect {
  transition: opacity 0.15s ease;
}
.slice-outline {
  stroke: rgba(0, 0, 0, 0.18);
  stroke-width: 1.5;
}
.slice.selected .slice-outline {
  stroke: var(--focus);
  stroke-width: 3;
}
.slice:focus-visible {
  outline: none;
}
.slice:focus-visible .slice-outline {
  stroke: var(--focus);
  stroke-width: 3;
}
.slice:hover .slice-rect {
  opacity: 0.9;
}
.slice-label {
  font-family: var(--font-serif);
  font-size: 14px;
  font-weight: 600;
  fill: var(--text-primary);
}
.slice-meta {
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  fill: var(--text-muted);
}
.threat-label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  fill: var(--danger);
}
.threat-line {
  stroke: var(--danger);
  stroke-width: 2.5;
  stroke-dasharray: 5 4;
}
.threat-head {
  fill: var(--danger);
}
.reading-key {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 20px;
  margin-top: 14px;
}
.reading-key span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.key-swatch {
  width: 26px;
  height: 15px;
  border-radius: 4px;
  border: 1px solid var(--border-strong);
  flex: none;
}
.key-swatch.solid {
  background: #184f95;
}
.key-swatch.porous {
  background: #cde2fb;
  /* two "holes" to suggest porosity */
  background-image: radial-gradient(circle at 30% 50%, var(--surface-1) 2.5px, transparent 3px),
    radial-gradient(circle at 70% 50%, var(--surface-1) 2.5px, transparent 3px);
}
.threat-control {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.threat-control input[type='range'] {
  width: 100%;
  accent-color: var(--danger);
}
.threat-note {
  margin: 0;
}
.panel {
  padding: 18px;
  position: sticky;
  top: 80px;
}
.layer-stage {
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.panel h2 {
  margin: 4px 0 8px;
  font-size: 1.2rem;
}
.hazard {
  margin: 0 0 14px;
  color: var(--text-secondary);
}
.layer-fam {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.fam-chip {
  display: inline-grid;
  place-items: center;
  min-width: 52px;
  height: 34px;
  padding: 0 10px;
  border-radius: 6px;
  font-family: var(--font-serif);
  font-weight: 700;
  border: 1px solid rgba(0, 0, 0, 0.12);
}
.sub {
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
  margin: 18px 0 8px;
}
.areas {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.areas li {
  display: flex;
  align-items: center;
  gap: 9px;
}
.a-name {
  flex: 1;
  font-size: 0.92rem;
}
.a-avg {
  display: inline-grid;
  place-items: center;
  min-width: 34px;
  height: 24px;
  border-radius: 5px;
  font-size: 0.82rem;
  font-weight: 700;
  border: 1px solid rgba(0, 0, 0, 0.12);
}
.tag {
  display: inline-grid;
  place-items: center;
  min-width: 28px;
  height: 20px;
  padding: 0 5px;
  border-radius: 4px;
  background: var(--surface-0);
  border: 1px solid var(--border-strong);
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-secondary);
}
.holes-list {
  margin: 0;
  padding-left: 18px;
  color: var(--text-secondary);
  font-size: 0.92rem;
}
.holes-list li {
  margin-bottom: 3px;
}
.note {
  margin: 12px 0 0;
}
.reading {
  padding: 18px;
}
.reading h3 {
  margin: 0 0 8px;
  font-size: 1.05rem;
}
.reading ul {
  margin: 0 0 12px;
  padding-left: 18px;
  color: var(--text-secondary);
  font-size: 0.92rem;
}
.caveat {
  margin: 0;
  max-width: 80ch;
}
@media (max-width: 860px) {
  .model {
    grid-template-columns: 1fr;
  }
  .panel {
    position: static;
  }
}
</style>
