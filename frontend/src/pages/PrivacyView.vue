<script setup lang="ts">
/**
 * Short privacy notice. Written to be read, not to be survived: what is stored,
 * why, for how long, and how to get rid of it.
 */
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useSubmitter } from '@/composables/useSubmitter'

const { reset } = useSubmitter()
const contact = (import.meta.env.VITE_CONTACT_EMAIL ?? '').trim()

onMounted(() => {
  document.title = 'Privacy · AI Safety Agendas × Problems Map'
})
</script>

<template>
  <div class="wrap wrap--narrow page">
    <p class="eyebrow">Privacy</p>
    <h1>What this site stores</h1>

    <p class="lead">
      Reading the map stores nothing. The only data we ever collect is what you type into a feedback
      form and choose to send.
    </p>

    <h2>If you never submit feedback</h2>
    <p>
      Nothing is stored about you. There is no analytics script, no tracking pixel, no advertising
      network, and no cookie. The map itself is a static file served from a CDN.
    </p>

    <h2>If you submit feedback</h2>
    <p>We store, in a Postgres database in the EU:</p>
    <ul>
      <li>the research area (and any agendas) you answered about, and the maturity and familiarity ratings you gave;</li>
      <li>your free-text comments, if you wrote any;</li>
      <li>your name and email <strong>only</strong> if you turned off anonymous submission and typed them;</li>
      <li>your consent choices;</li>
      <li>a random identifier generated in your browser, so repeat answers can be counted as one person;</li>
      <li>a salted SHA-256 hash of your IP address and your browser's user-agent string, to make abuse of the form detectable.</li>
    </ul>

    <p>
      <strong>Your IP address itself is never written down.</strong> It is hashed with a secret salt
      before storage and cannot be read back out. Anonymous submissions carry no name, no email, and
      no contact consent; the server discards those fields before the row is written, even if a
      client sends them.
    </p>

    <h2>Why we keep it</h2>
    <p>
      To correct the map. Maturity ratings build the community view of the grid, familiarity ratings
      let them be weighted, notes point at evidence we missed, and contact details let us ask a
      follow-up question, but only if you ticked the box saying we may. Consent is opt-in and you can
      submit fully anonymously without losing any of the site's function.
    </p>

    <h2>What we never do</h2>
    <p>
      We do not sell or share the data, do not publish raw submissions, and do not attribute a quote
      to you unless you consented to reuse. The public API returns aggregate counts only, no single
      submission is ever readable from the web.
    </p>

    <h2>Your data on your own device</h2>
    <p>
      Your name, email, consent choices, random identifier and the list of agendas you have answered
      are kept in your browser's local storage so the form can prefill. Clearing site data removes
      them, or use this button:
    </p>
    <p><button type="button" class="btn" @click="reset()">Forget me on this device</button></p>

    <h2>Your rights</h2>
    <p>
      Under the GDPR you can ask for a copy of what we hold about you, ask for it to be corrected, or
      ask for it to be deleted. Anonymous submissions cannot be traced back to you, so they cannot be
      individually deleted on request; that is the trade-off anonymity buys.
      <template v-if="contact">
        Write to <a :href="`mailto:${contact}`">{{ contact }}</a>.
      </template>
      <template v-else>
        The contact address for these requests is set by the site operator in
        <code>VITE_CONTACT_EMAIL</code>.
      </template>
    </p>

    <h2>Retention</h2>
    <p>
      Feedback is kept while this project is active, because later iterations of the map are built
      from it. Contact details are removed on request, and anything you asked us not to reuse is not
      reused.
    </p>

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

ul {
  color: var(--ink-secondary);
  max-width: var(--measure);
  padding-left: 1.2rem;
  display: grid;
  gap: 0.3rem;
}

.back {
  margin-top: 2.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--rule);
}
</style>
