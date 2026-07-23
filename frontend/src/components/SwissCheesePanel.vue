<script setup lang="ts">
/**
 * Swiss-cheese defence-in-depth view, an INTERPRETIVE OVERLAY, not a result.
 *
 * Read swiss_layers.json before touching this. In short: only the hole sizes come
 * from the verified matrix (weak tier → big hole). The layer ordering and the
 * independence the picture implies are editorial judgments that the dataset does
 * not contain, and the dataset's own reading note argues these defences are
 * correlated rather than independent. That is why the caveat is permanent, and
 * why the whole section stays hidden until an operator deliberately authors
 * layers.
 */
import { computed, ref } from 'vue'
import { atlas, cellTier, getAreaByTag, tierRank } from '@/lib/atlas'
import { tierStyle } from '@/lib/scales'
import { swissLayers } from '@/lib/atlas'
import type { CellTier } from '@/types/atlas'

const enabled = computed(() => swissLayers.enabled && swissLayers.layers.length > 0)

const hazard = ref(
  atlas.problems.some((p) => p.id === swissLayers.default_hazard) ? swissLayers.default_hazard : (atlas.problems[0]?.id ?? 'P1'),
)

/** A layer's members are area tags and/or agenda ids; expand to agenda ids. */
function agendaIdsOf(members: string[]): string[] {
  const ids = new Set<string>()
  for (const member of members) {
    const area = getAreaByTag(member)
    if (area) {
      for (const id of area.agenda_ids) ids.add(id)
    } else if (atlas.matrix[member]) {
      ids.add(member)
    }
  }
  return [...ids]
}

/**
 * Hole size is the only thing derived deterministically: the best tier the layer
 * reaches against the hazard maps to how much of the layer is open.
 */
const HOLE_BY_TIER: Record<string, number> = {
  'Robust (small scale)': 0.14,
  'Strong existence proof': 0.32,
  'Early / partial': 0.55,
  Untested: 0.78,
  'Never demonstrated': 0.88,
  Contested: 0.72,
}

const rows = computed(() =>
  swissLayers.layers.map((layer) => {
    const ids = agendaIdsOf(layer.members)
    const tiers = ids.map((id) => cellTier(id, hazard.value)).filter((t): t is Exclude<CellTier, null> => t !== null)
    const best = tiers.length ? tiers.reduce((a, b) => (tierRank(a) <= tierRank(b) ? a : b)) : null
    const hole = best ? (HOLE_BY_TIER[best] ?? 0.8) : 1

    return {
      layer,
      best,
      hole,
      reaching: tiers.length,
      total: ids.length,
      style: tierStyle(best),
    }
  }),
)

/** Residual risk if you (wrongly) treated the layers as independent. */
const residual = computed(() => rows.value.reduce((product, row) => product * row.hole, 1))
</script>

<template>
  <section v-if="enabled" class="swiss" aria-labelledby="swiss-title">
    <p class="eyebrow">Interpretive overlay · not a result from the matrix</p>
    <h2 id="swiss-title">Defence in depth, read as Swiss cheese</h2>

    <p class="swiss__caveat" role="note">
      <strong>Read this first.</strong> {{ swissLayers.caveat }}
    </p>

    <div class="swiss__controls">
      <label for="swiss-hazard">Hazard</label>
      <select id="swiss-hazard" v-model="hazard">
        <option v-for="problem in atlas.problems" :key="problem.id" :value="problem.id">
          {{ problem.id }} · {{ problem.name }}
        </option>
      </select>
    </div>

    <ol class="layers">
      <li v-for="row in rows" :key="row.layer.id" class="layer">
        <div class="layer__meta">
          <h3 class="layer__name">{{ row.layer.name }}</h3>
          <p class="layer__rationale">{{ row.layer.rationale }}</p>
          <p class="layer__stat tabular">
            {{ row.best ?? 'nothing reaches this hazard' }} ·
            {{ row.reaching }}/{{ row.total }} agendas reach it
          </p>
        </div>

        <div
          class="layer__slice"
          :class="{ hatched: row.style.hatched }"
          :style="{ backgroundColor: row.best ? row.style.fill : 'var(--surface-sunken)' }"
          role="img"
          :aria-label="`${row.layer.name}: ${row.best ?? 'no coverage'}, roughly ${Math.round(row.hole * 100)} percent open against ${hazard}`"
        >
          <span class="layer__hole" :style="{ inlineSize: `${row.hole * 100}%` }"></span>
        </div>
      </li>
    </ol>

    <p class="swiss__residual">
      Multiplying the holes as if the layers were independent leaves
      <strong class="tabular">{{ (residual * 100).toFixed(1) }}%</strong> of paths open. Treat that
      number as an illustration of the assumption, not as an estimate: if the layers are correlated,
      and the dataset argues they are, the true figure is worse, and no multiplication of
      independent holes can tell you by how much.
    </p>
  </section>
</template>

<style scoped>
.swiss {
  margin-top: 1rem;
}

.swiss__caveat {
  background: var(--surface-sunken);
  border-left: 3px solid var(--tier-contested);
  padding: 0.85rem 1rem;
  border-radius: 0 var(--radius) var(--radius) 0;
  font-size: 0.875rem;
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.swiss__controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin: 1.25rem 0;
  font-size: 0.875rem;
}

.swiss__controls select {
  font: inherit;
  padding: 0.35rem 0.5rem;
  border-radius: var(--radius);
  border: 1px solid var(--rule-strong);
  background: var(--surface);
  color: var(--ink);
  max-width: 100%;
}

.layers {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}

.layer {
  display: grid;
  gap: 0.5rem 1.25rem;
  align-items: center;
  grid-template-columns: 1fr;
}

.layer__name {
  font-size: 0.9375rem;
  margin: 0;
}

.layer__rationale,
.layer__stat {
  margin: 0.15rem 0 0;
  font-size: 0.8125rem;
  color: var(--ink-secondary);
}

.layer__stat {
  color: var(--ink-muted);
}

.layer__slice {
  height: 2.5rem;
  border-radius: var(--radius);
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.layer__hole {
  display: block;
  block-size: 100%;
  background: var(--page);
  box-shadow: inset 2px 0 0 var(--rule);
}

.swiss__residual {
  margin-top: 1.25rem;
  font-size: 0.875rem;
  color: var(--ink-secondary);
  max-width: var(--measure);
}

@media (min-width: 720px) {
  .layer {
    grid-template-columns: minmax(0, 1fr) minmax(200px, 40%);
  }
}
</style>
