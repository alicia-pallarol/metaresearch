<script setup>
defineProps({
  waking: { type: Boolean, default: false },
  attempt: { type: Number, default: 0 },
  total: { type: Number, default: 0 },
})
</script>

<template>
  <div class="loading card" role="status" aria-live="polite">
    <span class="spinner" aria-hidden="true"></span>
    <div>
      <p class="line1">
        {{ waking ? 'Waking the server…' : 'Loading research areas…' }}
      </p>
      <p class="small muted" v-if="waking">
        The backend runs on a free tier that sleeps after inactivity. The first
        request can take up to a minute.
        <template v-if="total"> (checking {{ attempt }} of {{ total }})</template>
      </p>
    </div>
  </div>
</template>

<style scoped>
.loading {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 22px;
}
.line1 {
  margin: 0 0 2px;
  font-weight: 600;
}
.spinner {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 3px solid var(--border-strong);
  border-top-color: var(--accent);
  animation: spin 0.8s linear infinite;
  flex: none;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
