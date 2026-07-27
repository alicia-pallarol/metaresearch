<script setup lang="ts">
/** Where this project comes from, what it is trying to do, and who is behind it. */
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'

onMounted(() => {
  document.title = 'About · AI Safety Agendas'
})


const fellows = [
  { name: 'Alicia Pallarol Isábal', linkedin: 'https://www.linkedin.com/in/alicia-pallarol-isabal/' },
  { name: 'Shaun Paul', linkedin: 'https://www.linkedin.com/in/shaunpaul2002/' },
]
const mentor = { name: 'Ihor Kendiukhov', linkedin: 'https://www.linkedin.com/in/ikendiukhov/' }
</script>

<template>
  <div class="wrap wrap--narrow page">
    <p class="eyebrow">About</p>
    <h1>About this project</h1>

    <p class="lead">
      This is a metaresearch project on the field of AI safety, built during the SAIGE fellowship.
    </p>

    <h2>The goal</h2>
    <p>
      As investment in AI capabilities accelerates, it is worth asking whether safety research is
      keeping pace and pointed at the risks that matter most. The aim here is concrete: to build a
      clear, empirical picture of the AI safety research landscape: what is being worked on, how
      mature the evidence behind it is, and where the gaps are, so the field can be seen as it stands
      rather than as it is assumed to be.
    </p>

    <h2>The part that matters most</h2>
    <p>
      The map is only a starting point, and our own reading of the field is the least interesting
      thing about it. What makes it worth building is the correction: every rating here is a draft,
      and the people who know an area best are the ones who can sharpen it. The feedback
      left on these pages builds the community view of the grid and keeps the map honest as the field
      moves. If you work in one of these areas, your reading is the point.
    </p>

    <h2>Where we hope it goes</h2>
    <p>
      Over time, we would like this to work as a small feedback mechanism for the field itself:
      showing funders where work is neglected, helping researchers find promising directions that are
      being overlooked, and giving the people who set standards a clearer view of where today's
      benchmarks fall short. That is the direction we are aiming in, not a claim we have arrived.
    </p>

    <h2>Who is behind it</h2>
    <p>The project is carried out by two SAIGE fellows, with mentorship.</p>
    <ul class="people">
      <li v-for="(person, i) in fellows" :key="i" class="person">
        <span class="person__role">Fellow</span>
        <span class="person__name">{{ person.name }}</span>
        <a class="person__link" :href="person.linkedin" target="_blank" rel="noopener">LinkedIn →</a>
      </li>
      <li class="person">
        <span class="person__role">Mentor</span>
        <span class="person__name">{{ mentor.name }}</span>
        <a class="person__link" :href="mentor.linkedin" target="_blank" rel="noopener">LinkedIn →</a>
      </li>
    </ul>

    <p class="back"><RouterLink to="/">← Back to the map</RouterLink></p>
  </div>
</template>

<style scoped>
.page {
  padding-block: 2rem 4rem;
}

.lead {
  font-size: 1.0625rem;
  color: var(--ink-secondary);
}

h2 {
  margin-top: 2rem;
  font-size: 1.05rem;
}

p {
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.people {
  list-style: none;
  margin: 1rem 0 0;
  padding: 0;
  display: grid;
  gap: 0.5rem;
  max-width: var(--measure);
}

.person {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.4rem 0.75rem;
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  padding: 0.6rem 0.8rem;
  background: var(--surface);
}

.person__role {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--ink-muted);
  font-weight: 600;
  flex: none;
}

.person__name {
  font-weight: 500;
  color: var(--ink);
}

.person__link {
  margin-left: auto;
  font-size: 0.875rem;
  white-space: nowrap;
}

.back {
  margin-top: 2.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--rule);
}
</style>
