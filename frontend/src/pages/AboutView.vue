<script setup>
import { FAMILIARITY_LEVELS } from '../config.js'
</script>

<template>
  <div class="container page">
    <article class="prose card">
      <h1>About this platform</h1>
      <p>
        This platform collects and visualizes self-reported familiarity across a
        framework of AI safety research areas. Its purpose is descriptive: to show
        where a community's collective expertise is concentrated and where it is
        sparse, so that gaps can be discussed openly. It is not an assessment of any
        individual and makes no claim about the quality of anyone's work.
      </p>

      <h2>The familiarity scale</h2>
      <ul class="scale">
        <li v-for="lvl in FAMILIARITY_LEVELS" :key="lvl.value">
          <strong>{{ lvl.value }}</strong> — {{ lvl.label }}
        </li>
      </ul>

      <h2>Method and honest reporting</h2>
      <p>
        Each research area shows the number of ratings behind it (<em>n</em>) and the
        mean familiarity. Where the sample is small (fewer than three ratings) the
        values are flagged as provisional. The heatmap encodes the mean familiarity
        on a single-hue scale; areas with no ratings are shown as neutral, not as
        zero, to distinguish "unknown" from "no familiarity".
      </p>

      <h2>Privacy</h2>
      <p>
        An email address is required only to deduplicate submissions and, for
        researchers who choose not to remain anonymous, to allow follow-up questions
        about their answers. Email addresses are never displayed and are never
        returned by the public API. Submitting again from the same address replaces
        the previous response. For rate limiting we store a salted, irreversible hash
        of the submitter's IP address — never the address itself.
      </p>

      <h2>The defense model</h2>
      <p>
        The <RouterLink to="/swiss-cheese">Swiss cheese defense model</RouterLink>
        page maps the research areas onto layered defenses. That conceptual mapping
        was drafted with AI assistance and has not been peer-reviewed; it is labeled
        as such and is offered as a starting point for discussion rather than a
        settled account.
      </p>

      <h2>Open by design</h2>
      <p>
        The framework, aggregation, and visualizations are intentionally simple and
        transparent. No accounts, tracking, or advertising are involved.
      </p>
    </article>
  </div>
</template>

<style scoped>
.page {
  padding-top: 32px;
}
.prose {
  max-width: 76ch;
  margin-inline: auto;
  padding: 28px 30px;
}
.prose h1 {
  margin-top: 0;
  font-size: 1.9rem;
}
.prose h2 {
  font-size: 1.2rem;
  margin: 26px 0 8px;
}
.prose p {
  color: var(--text-secondary);
}
.scale {
  padding-left: 18px;
  color: var(--text-secondary);
}
.scale li {
  margin-bottom: 4px;
}
</style>
