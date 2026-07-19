<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <a class="skip-link" href="#main">Skip to main content</a>
  <header class="site-header">
    <div class="container header-inner">
      <RouterLink to="/" class="brand" aria-label="Home">
        <span class="brand-mark" aria-hidden="true"></span>
        <span class="brand-text">
          <strong>AI Safety Research Familiarity</strong>
          <span class="brand-sub muted small">A community mapping of expertise</span>
        </span>
      </RouterLink>
      <nav class="site-nav" aria-label="Primary">
        <RouterLink to="/">Framework</RouterLink>
        <RouterLink to="/swiss-cheese">Defense model</RouterLink>
        <RouterLink to="/about">About</RouterLink>
      </nav>
    </div>
  </header>

  <main id="main">
    <RouterView />
  </main>

  <footer class="site-footer">
    <div class="container footer-inner">
      <p class="small muted">
        This platform records self-reported familiarity across AI safety research
        areas and visualizes it in aggregate. It is a research instrument, not an
        assessment of any individual. Submissions are aggregated; individual
        responses and contact details are never published.
      </p>
      <p class="small muted">
        <RouterLink to="/about">About &amp; methodology</RouterLink>
      </p>
    </div>
  </footer>
</template>

<style scoped>
.site-header {
  position: sticky;
  top: 0;
  z-index: 20;
  background: color-mix(in srgb, var(--surface-1) 92%, transparent);
  backdrop-filter: saturate(1.1) blur(6px);
  border-bottom: 1px solid var(--border);
}
.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 64px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  color: var(--text-primary);
}
.brand-mark {
  width: 22px;
  height: 22px;
  border-radius: 5px;
  background: linear-gradient(150deg, #184f95, #3987e5);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.15);
  flex: none;
}
.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.brand-sub {
  font-weight: 400;
}
.site-nav {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.site-nav a {
  padding: 7px 12px;
  border-radius: 6px;
  text-decoration: none;
  color: var(--text-secondary);
  font-weight: 500;
}
.site-nav a:hover {
  background: var(--surface-2);
  color: var(--text-primary);
}
.site-nav a.router-link-exact-active {
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
}
main {
  min-height: 60vh;
}
.site-footer {
  margin-top: 56px;
  border-top: 1px solid var(--border);
  background: var(--surface-1);
}
.footer-inner {
  padding-block: 24px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 760px;
}
@media (max-width: 560px) {
  .brand-sub {
    display: none;
  }
}
</style>
