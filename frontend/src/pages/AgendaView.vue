<script setup lang="ts">
/** One agenda in full, plus the community stats and the feedback form for it. */
import { computed, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { getAgenda, getAreaByName, problemsReachedBy } from '@/lib/atlas'
import { tierAbbrev, tierStyle } from '@/lib/scales'
import { useSummary } from '@/composables/useSummary'
import CommunityDistribution from '@/components/CommunityDistribution.vue'
import FeedbackForm from '@/components/FeedbackForm.vue'

const route = useRoute()
const { forArea, forAgenda, state } = useSummary()

const agendaId = computed(() => String(route.params.id ?? ''))
const agenda = computed(() => getAgenda(agendaId.value))
const area = computed(() => (agenda.value ? getAreaByName(agenda.value.area_name) : undefined))
const reached = computed(() => (agenda.value ? problemsReachedBy(agenda.value.id) : []))
const stats = computed(() => (agenda.value ? forAgenda(agenda.value.id) : undefined))
const areaStats = computed(() => (agenda.value ? forArea(agenda.value.area_tag) : undefined))

/**
 * "best_evidence_against" holds one of two very different things: a list of the
 * problems this agenda bites hardest on (e.g. "P2, P4"), or a maturity ceiling
 * when it bites nowhere in particular (e.g. "Early / partial at best"). The label
 * follows the content so the reader is not told "evidence against" a maturity tier.
 */
const bestAgainstIsProblems = computed(() =>
  /^P\d+(\s*,\s*P\d+)*$/.test((agenda.value?.best_evidence_against ?? '').trim()),
)
const bestAgainstLabel = computed(() =>
  bestAgainstIsProblems.value ? 'Most-targeted problems' : 'Strongest evidence maturity',
)

function setTitle(): void {
  document.title = agenda.value
    ? `${agenda.value.id} · ${agenda.value.agenda} · AI Safety Agendas`
    : 'Agenda not found · AI Safety Agendas'
}

onMounted(setTitle)
watch(agendaId, () => {
  setTitle()
  window.scrollTo({ top: 0 })
})
</script>

<template>
  <div class="wrap page">
    <p class="crumbs">
      <RouterLink to="/">The map</RouterLink>
      <span aria-hidden="true">/</span>
      <span v-if="agenda">{{ agenda.area_name }}</span>
    </p>

    <template v-if="agenda">
      <header class="head">
        <p class="eyebrow">{{ agenda.area_tag }} · {{ agenda.area_name }}</p>
        <h1 class="head__title"><span class="tabular head__id">{{ agenda.id }}</span> {{ agenda.agenda }}</h1>
        <p class="head__area">{{ area?.definition }}</p>
      </header>

      <div class="cols">
        <div class="cols__main stack">
          <section class="card block">
            <h2 class="block__title">The record</h2>
            <dl class="facts">
              <div>
                <dt>Mechanism</dt>
                <dd>{{ agenda.mechanism }}</dd>
              </div>
              <div>
                <dt>Assumption it rests on</dt>
                <dd>{{ agenda.assumption }}</dd>
              </div>
              <div>
                <dt>Key work</dt>
                <dd>{{ agenda.key_work }}</dd>
              </div>
              <div class="facts__pair">
                <div>
                  <dt>Overall maturity</dt>
                  <dd>
                    <span
                      class="chip"
                      :class="{ hatched: tierStyle(agenda.overall_maturity).hatched }"
                      :style="{ backgroundColor: tierStyle(agenda.overall_maturity).fill, color: tierStyle(agenda.overall_maturity).ink }"
                      >{{ agenda.overall_maturity }}</span
                    >
                  </dd>
                </div>
                <div>
                  <dt>{{ bestAgainstLabel }}</dt>
                  <dd>{{ agenda.best_evidence_against }}</dd>
                </div>
              </div>
            </dl>
            <p class="block__note">
              "Overall maturity" rates this line of work as a technique in general. The tiers below
              rate it against one problem at a time. They can differ on purpose: a technique can be
              mature in general and unproven against a particular problem.
            </p>
          </section>

          <section class="card block">
            <h2 class="block__title">Against each problem</h2>
            <ul v-if="reached.length" class="reach">
              <li v-for="row in reached" :key="row.problem.id" :id="`p-${row.problem.id}`" class="reach__item">
                <span
                  class="chip chip--sm"
                  :class="{ hatched: tierStyle(row.tier).hatched }"
                  :style="{ backgroundColor: tierStyle(row.tier).fill, color: tierStyle(row.tier).ink }"
                  :title="row.tier"
                  >{{ tierAbbrev(row.tier) }}</span
                >
                <span class="reach__body">
                  <span class="reach__name">
                    <span class="tabular">{{ row.problem.id }}</span> {{ row.problem.name }}
                  </span>
                  <span class="reach__tier">{{ row.tier }}</span>
                  <span class="reach__def">{{ row.problem.short_def }}</span>
                </span>
              </li>
            </ul>
            <p v-else class="block__note">This agenda is not rated against any problem yet.</p>
          </section>
        </div>

        <aside class="cols__side stack">
          <section class="card block">
            <h2 class="block__title">What researchers said</h2>
            <template v-if="(stats?.count ?? 0) > 0 || (areaStats?.count ?? 0) > 0">
              <CommunityDistribution v-if="stats" :subject="agenda.id" :stats="stats" />
              <p v-else class="block__note">
                Nobody has rated {{ agenda.id }} on its own yet. Its area's ratings are below.
              </p>
              <CommunityDistribution
                v-if="areaStats"
                class="stats__area"
                :subject="agenda.area_name"
                :stats="areaStats"
              />
            </template>
            <p v-else-if="state === 'disabled'" class="block__note">
              Feedback collection is switched off on this deployment.
            </p>
            <p v-else-if="state === 'unavailable'" class="block__note">
              The feedback server is asleep or unreachable, so community numbers are not showing. The
              map itself is unaffected.
            </p>
            <p v-else class="block__note">Nobody has answered for this agenda yet. You would be the first.</p>
          </section>

          <FeedbackForm :subject="{ kind: 'agenda', agenda }" />
        </aside>
      </div>
    </template>

    <section v-else class="card block">
      <h1>No agenda with that id</h1>
      <p>
        The id <code>{{ agendaId }}</code> is not in this iteration of the map.
        <RouterLink to="/">Back to the map</RouterLink>.
      </p>
    </section>
  </div>
</template>

<style scoped>
.page {
  padding-block: 2rem 4rem;
}

.crumbs {
  display: flex;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--ink-muted);
  margin-bottom: 1.25rem;
}

.head {
  margin-bottom: 1.75rem;
}

.head__title {
  display: flex;
  gap: 0.6rem;
  align-items: baseline;
  flex-wrap: wrap;
  font-size: clamp(1.5rem, 1.2rem + 1.4vw, 2.1rem);
}

.head__id {
  color: var(--ink-muted);
  font-size: 0.7em;
  font-weight: 700;
}

.head__area {
  color: var(--ink-secondary);
  font-size: 0.9375rem;
}

.cols {
  display: grid;
  gap: 1.25rem;
  align-items: start;
}

.block {
  padding: 1.15rem 1.3rem 1.3rem;
}

.block__title {
  font-size: 1rem;
  margin-bottom: 0.9rem;
}

.block__note {
  font-size: 0.8125rem;
  color: var(--ink-muted);
  margin: 0.9rem 0 0;
  max-width: var(--measure);
}

.facts {
  display: grid;
  gap: 0.9rem;
  margin: 0;
}

.facts dt {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--ink-muted);
  font-weight: 600;
}

.facts dd {
  margin: 0.2rem 0 0;
  font-size: 0.9375rem;
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.facts__pair {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.9rem;
}

.chip {
  display: inline-block;
  border-radius: 3px;
  padding: 0.12rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.chip--sm {
  font-size: 0.625rem;
  font-weight: 700;
  min-width: 1.9rem;
  text-align: center;
  flex: none;
}

.reach {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.7rem;
}

.reach__item {
  display: flex;
  gap: 0.6rem;
  align-items: flex-start;
  /* Clear the sticky header when arrived at from an agenda-grid cell link. */
  scroll-margin-top: 5rem;
  border-radius: var(--radius);
}

/* Landed on from an "Every agenda, every problem" cell: mark which problem. */
.reach__item:target {
  outline: 2px solid var(--focus);
  outline-offset: 5px;
}

.reach__body {
  min-width: 0;
}

.reach__name {
  display: block;
  font-size: 0.875rem;
  font-weight: 600;
}

.reach__tier {
  display: block;
  font-size: 0.75rem;
  color: var(--ink-muted);
}

.reach__def {
  display: block;
  font-size: 0.8125rem;
  color: var(--ink-secondary);
  margin-top: 0.15rem;
  max-width: var(--measure);
}

.stats__area {
  margin-top: 0.6rem;
}

@media (min-width: 940px) {
  .cols {
    grid-template-columns: minmax(0, 1.15fr) minmax(340px, 0.85fr);
    gap: 1.5rem;
  }
}
</style>
