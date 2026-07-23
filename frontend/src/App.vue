<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { atlas } from '@/lib/atlas'
import { refreshSummary } from '@/composables/useSummary'

type Theme = 'system' | 'light' | 'dark'
const theme = ref<Theme>('system')

function applyTheme(next: Theme): void {
  theme.value = next
  const root = document.documentElement
  if (next === 'system') root.removeAttribute('data-theme')
  else root.setAttribute('data-theme', next)
  try {
    localStorage.setItem('atlas.theme', next)
  } catch {
    // Storage unavailable; the toggle still works for this session.
  }
}

function cycleTheme(): void {
  applyTheme(theme.value === 'system' ? 'light' : theme.value === 'light' ? 'dark' : 'system')
}

onMounted(() => {
  try {
    const stored = localStorage.getItem('atlas.theme')
    if (stored === 'light' || stored === 'dark') applyTheme(stored)
  } catch {
    // ignore
  }
  // Aggregates are optional to the page, so this is fire-and-forget: a sleeping
  // API just leaves the counters showing their empty state.
  void refreshSummary()
})
</script>

<template>
  <a class="skip-link" href="#main">Skip to content</a>

  <header class="site">
    <div class="wrap site__inner">
      <RouterLink class="site__brand" to="/">
        <span class="site__name">Agendas × Problems</span>
        <span class="site__iter">{{ atlas.meta.iteration_label }}</span>
      </RouterLink>

      <nav class="site__nav" aria-label="Main">
        <RouterLink to="/">Map</RouterLink>
        <RouterLink to="/methodology">Methodology</RouterLink>
        <RouterLink to="/privacy">Privacy</RouterLink>
        <button
          type="button"
          class="site__theme"
          :aria-label="`Colour theme: ${theme}. Click to change.`"
          @click="cycleTheme"
        >
          {{ theme === 'system' ? '◐' : theme === 'light' ? '☀' : '☾' }}
        </button>
      </nav>
    </div>
  </header>

  <main id="main">
    <RouterView />
  </main>

  <footer class="foot">
    <div class="wrap foot__inner">
      <p class="foot__claim">
        <strong>Manually curated.</strong> Every rating and every note on this site was written by
        hand from published work. Nothing here is model-generated, and the site calls no model at
        runtime. Sources last checked {{ atlas.meta.source_check_date }}.
      </p>
      <p class="foot__links">
        <RouterLink to="/methodology">Methodology</RouterLink>
        <a href="/data/atlas.json" download>Source data (JSON)</a>
        <RouterLink to="/privacy">Privacy</RouterLink>
      </p>
      <p class="foot__note">
        {{ atlas.meta.iteration_label }}, a draft put out for correction, not a finished result.
        Ratings are judgments about evidence maturity, not about how important a research area is.
      </p>
    </div>
  </footer>
</template>

<style scoped>
.site {
  border-bottom: 1px solid var(--rule);
  background: var(--surface);
  position: sticky;
  top: 0;
  z-index: 10;
}

.site__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding-block: 0.7rem;
  flex-wrap: wrap;
}

.site__brand {
  display: flex;
  align-items: baseline;
  gap: 0.6rem;
  color: inherit;
  text-decoration: none;
}

.site__name {
  font-weight: 600;
  letter-spacing: -0.01em;
}

.site__iter {
  font-size: 0.6875rem;
  color: var(--ink-muted);
  text-transform: uppercase;
  letter-spacing: 0.07em;
}

.site__nav {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.875rem;
}

.site__nav a {
  color: var(--ink-secondary);
  text-decoration: none;
}

.site__nav a:hover,
.site__nav a.router-link-active {
  color: var(--ink);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.site__theme {
  border: 1px solid var(--rule-strong);
  background: var(--surface);
  border-radius: var(--radius);
  width: 2rem;
  height: 1.85rem;
  cursor: pointer;
  line-height: 1;
  color: var(--ink-secondary);
}

.foot {
  border-top: 1px solid var(--rule);
  background: var(--surface);
  margin-top: 2rem;
}

.foot__inner {
  padding-block: 2rem 2.75rem;
}

.foot__claim {
  font-size: 0.875rem;
  color: var(--ink-secondary);
  max-width: var(--measure);
}

.foot__links {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  font-size: 0.875rem;
  margin: 1rem 0;
}

.foot__note {
  font-size: 0.8125rem;
  color: var(--ink-muted);
  max-width: var(--measure);
  margin: 0;
}
</style>
