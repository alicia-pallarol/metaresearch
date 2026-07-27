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
        <span class="site__name">AI Safety Agendas</span>
      </RouterLink>

      <nav class="site__nav" aria-label="Main">
        <RouterLink to="/">Map</RouterLink>
        <RouterLink to="/methodology">Methodology</RouterLink>
        <RouterLink to="/about">About</RouterLink>
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
      <p class="foot__links">
        <RouterLink to="/methodology">Methodology</RouterLink>
        <RouterLink to="/about">About</RouterLink>
        <a href="/data/atlas.json" download>Source data (JSON)</a>
        <RouterLink to="/privacy">Privacy</RouterLink>
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

/* A serif wordmark on an otherwise sans page: reads as academic and credible,
   distinct from the UI, and needs no webfont (system serifs only, so the strict
   font-src 'self' CSP is untouched). One line by design, a toolbar mark should
   stay compact rather than stack. */
.site__name {
  font-family: 'Iowan Old Style', 'Palatino Linotype', Palatino, Georgia, 'Times New Roman', serif;
  font-weight: 600;
  font-size: 1.2rem;
  letter-spacing: 0;
  color: var(--ink);
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

.foot__links {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  font-size: 0.875rem;
  margin: 1rem 0 0;
}
</style>
