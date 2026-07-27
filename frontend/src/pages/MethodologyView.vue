<script setup lang="ts">
/** About / Methodology: the legend notes verbatim, plus the problem and area definitions. */
import { onMounted, ref } from 'vue'
import { atlas } from '@/lib/atlas'
import TierLegend from '@/components/TierLegend.vue'
import AreaProfileDialog from '@/components/AreaProfileDialog.vue'
import type { Area } from '@/types/atlas'

onMounted(() => {
  document.title = 'Methodology · AI Safety Agendas'
})

const familiarityLevels = ['0', '1', '2', '3', '4'] as const

// The research area whose agenda profile is open, or null.
const openedArea = ref<Area | null>(null)
</script>

<template>
  <div class="wrap page">
    <header class="head">
      <p class="eyebrow">About &amp; methodology</p>
      <h1>How to read this map</h1>
      <p class="head__lead">
        Everything on this site is a manually curated judgment about published work, last checked in
        {{ atlas.meta.source_check_date }}. No part of the analysis is model-generated, and the
        running site never calls a model. The dataset is
        <a href="/data/atlas.json" download>downloadable in full</a>.
      </p>
    </header>

    <section class="section-block">
      <h2>The tiers</h2>
      <TierLegend />
    </section>

    <section class="section-block">
      <h2>Reading the grid: two switches</h2>
      <dl class="notes">
        <div class="note">
          <dt>Our research, or the community</dt>
          <dd>
            The first switch changes whose judgment the colours show: our curated ratings, or the
            maturity ratings researchers have submitted. Which agendas address which problem is our
            claim in both views; the community does not vote on that structure, only on how mature
            the work is. So a cell is the same set of agendas either way; only the colour changes
            hands. An area nobody has rated yet stays uncoloured in the community view rather than
            guessing.
          </dd>
        </div>
        <div class="note">
          <dt>Best case, average, worst case</dt>
          <dd>
            Every cell stands on a set of ratings, the tiers of the agendas behind it, and in the
            community view the tiers those agendas were voted. The second switch reads the strongest
            quarter of that set, the middle of it, or the weakest quarter. The middle band, labelled
            <em>average</em> on the switch, is summarised by its <em>median</em>, never a numeric
            mean: the tiers are names, not numbers, and two of them
            (<em>Contested</em>, <em>Never demonstrated</em>) are not rungs on the evidence ladder at
            all. Where a band has no single middle, the weaker tier shows; the map does not round up.
            Below eight ratings the "quarter" is a single rating, so <em>Our research · best case</em>
            is exactly the best-of-its-agendas grid this map has always led with.
          </dd>
        </div>
        <div class="note">
          <dt>Maturity is about the research, not the problem</dt>
          <dd>
            When you rate an area, you are rating the body of work itself, how much of it exists, how
            well it has been done, what it has actually shown, not how much of any one problem it
            solves. An agenda can be mature research and still only chip at a hard problem. That is why
            a vote is scoped to a research area (and optionally its agendas), and never to a single
            Area × Problem cell.
          </dd>
        </div>
      </dl>
    </section>

    <section class="section-block">
      <h2>Focus areas: how the ranking is computed</h2>
      <p class="lead">
        The "focus areas" section is not a hand-picked list. It reads the grid, under whichever
        source the grid is set to, and sorts the research areas, and separately the problems, into
        three categories. Anything already rated near-Robust is left out, as the closest to solved.
      </p>
      <dl class="notes">
        <div class="note">
          <dt>Easy to intervene</dt>
          <dd>
            The best result already on the board is strong, but the work is not finished. Keyed off
            the best tier reached (tempered by the average, so a single fluke cell can't crown a weak
            unit). A push has something concrete to build on.
          </dd>
        </div>
        <div class="note">
          <dt>Bottlenecks</dt>
          <dd>
            Much depends on it, and it is still weak. For a problem, that is how many agendas are
            betting on it; for an area, how much its problems are shared with other areas. Solving one
            unblocks the rest.
          </dd>
        </div>
        <div class="note">
          <dt>Underexplored</dt>
          <dd>
            Thin: few agendas, many blank cells, low maturity, and <em>not</em> merely contested.
            Room the map has barely entered, as distinct from room that was tried and disputed.
          </dd>
        </div>
        <div class="note">
          <dt>Where it lives</dt>
          <dd>
            Every signal, threshold and weight is a named, commented constant in one file,
            <code>frontend/src/lib/focus.ts</code>, so the rule can be read and changed directly
            rather than reverse-engineered. No prose is generated; each item's one-line reason is
            assembled from its own numbers.
          </dd>
        </div>
      </dl>
    </section>

    <section class="section-block">
      <h2>Notes on the method</h2>
      <dl class="notes">
        <div v-for="(text, title) in atlas.legend.notes" :key="title" class="note">
          <dt>{{ title }}</dt>
          <dd>{{ text }}</dd>
        </div>
      </dl>
    </section>

    <section class="section-block">
      <h2>The familiarity scale</h2>
      <p class="lead">
        When researchers answer, they rate their own familiarity on this scale alongside their
        maturity rating. It is what lets the ratings be read with the right weight: a rating from
        someone who works in the area is not the same input as one from a bystander.
      </p>
      <dl class="notes">
        <div v-for="level in familiarityLevels" :key="level" class="note">
          <dt class="tabular">{{ level }}</dt>
          <dd>{{ atlas.legend.familiarity_scale[level] }}</dd>
        </div>
      </dl>
    </section>

    <section class="section-block">
      <h2>The {{ atlas.problems.length }} problems</h2>
      <dl class="notes">
        <div v-for="problem in atlas.problems" :key="problem.id" class="note">
          <dt><span class="tabular">{{ problem.id }}</span> {{ problem.name }}</dt>
          <dd>{{ problem.short_def }}</dd>
        </div>
      </dl>
    </section>

    <section class="section-block">
      <h2>The {{ atlas.areas.length }} research areas</h2>
      <p class="lead">
        Open any area to see the agendas inside it, each with its mechanism, the assumption it rests
        on and its key work.
      </p>
      <dl class="notes">
        <div v-for="area in atlas.areas" :key="area.tag" class="note">
          <dt>
            <button type="button" class="area-open" @click="openedArea = area">
              <span class="tabular">{{ area.tag }}</span> {{ area.name }}
              <span class="note__count">{{ area.agenda_ids.length }} agendas &rarr;</span>
            </button>
          </dt>
          <dd>{{ area.definition }}</dd>
        </div>
      </dl>
    </section>

    <AreaProfileDialog :area="openedArea" @close="openedArea = null" />
  </div>
</template>

<style scoped>
.page {
  padding-block: 2rem 4rem;
}

.head {
  margin-bottom: 2rem;
}

.head__lead,
.lead {
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.section-block {
  padding-top: 2rem;
  margin-top: 2rem;
  border-top: 1px solid var(--rule);
}

.notes {
  display: grid;
  gap: 1rem;
  margin: 1rem 0 0;
}

.note dt {
  font-weight: 600;
  font-size: 0.9375rem;
}

.note dd {
  margin: 0.2rem 0 0;
  color: var(--ink-secondary);
  font-size: 0.9375rem;
  max-width: var(--measure);
}

.note__count {
  font-size: 0.75rem;
  color: var(--ink-muted);
  font-weight: 400;
  margin-left: 0.35rem;
}

/* The area header is a button so the whole row opens its agenda profile. */
.area-open {
  display: inline;
  background: none;
  border: 0;
  padding: 0;
  margin: 0;
  font: inherit;
  font-weight: 600;
  color: var(--link);
  cursor: pointer;
  text-align: left;
}

.area-open:hover {
  text-decoration: underline;
  text-underline-offset: 2px;
}

.area-open:hover .note__count {
  color: var(--link);
}
</style>
