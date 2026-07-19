<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { submitFamiliarity, ApiError } from '../api.js'
import { FAMILIARITY_LEVELS } from '../config.js'

const props = defineProps({
  areas: { type: Array, required: true },
})
const emit = defineEmits(['close', 'submitted'])

const form = reactive({
  name: '',
  email: '',
  anonymous: false,
  website: '', // honeypot — must stay empty
})

// ratings: tag -> number | null (null = skipped/blank)
const ratings = reactive({})
for (const a of props.areas) ratings[a.tag] = null

const submitting = ref(false)
const generalError = ref('')
const fieldErrors = reactive({ name: '', email: '', ratings: '' })
const dialogEl = ref(null)
const firstFieldEl = ref(null)

const answeredCount = computed(() => Object.values(ratings).filter((v) => v !== null).length)

// Group for a tidy matrix layout.
const groups = computed(() => {
  const out = []
  let current = null
  for (const a of props.areas) {
    if (!current || current.name !== a.problem_group) {
      current = { name: a.problem_group, areas: [] }
      out.push(current)
    }
    current.areas.push(a)
  }
  return out
})

function clearErrors() {
  generalError.value = ''
  fieldErrors.name = ''
  fieldErrors.email = ''
  fieldErrors.ratings = ''
}

async function onSubmit() {
  if (submitting.value) return
  clearErrors()
  submitting.value = true
  try {
    const res = await submitFamiliarity({
      name: form.name,
      email: form.email,
      anonymous: form.anonymous,
      // Non-anonymous submission carries consent to be contacted (see label).
      contact_consent: !form.anonymous,
      website: form.website,
      ratings: { ...ratings },
    })
    emit('submitted', { replaced: !!res.replaced })
  } catch (err) {
    if (err instanceof ApiError && err.status === 422 && err.fields?.length) {
      for (const f of err.fields) {
        if (f.field in fieldErrors) fieldErrors[f.field] = f.message
        else generalError.value = f.message
      }
      if (!Object.values(fieldErrors).some(Boolean)) generalError.value = err.message
    } else if (err instanceof ApiError && err.status === 429) {
      generalError.value = err.message
    } else {
      generalError.value =
        'We could not submit your response. The server may be waking from sleep — please try again in a moment.'
    }
  } finally {
    submitting.value = false
  }
}

function onKeydown(e) {
  if (e.key === 'Escape') emit('close')
}

onMounted(async () => {
  document.body.style.overflow = 'hidden'
  document.addEventListener('keydown', onKeydown)
  await nextTick()
  firstFieldEl.value?.focus()
})
onBeforeUnmount(() => {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="overlay" @click.self="emit('close')">
    <div
      ref="dialogEl"
      class="dialog card"
      role="dialog"
      aria-modal="true"
      aria-labelledby="sf-title"
    >
      <header class="dialog-head">
        <div>
          <h2 id="sf-title">Submit your familiarity</h2>
          <p class="small muted">
            Rate each research area from 0 to 3. You may leave any area blank.
          </p>
        </div>
        <button class="btn close" @click="emit('close')" aria-label="Close form">Close</button>
      </header>

      <form class="dialog-body" @submit.prevent="onSubmit" novalidate>
        <div v-if="generalError" class="notice notice-warn" role="alert">{{ generalError }}</div>

        <div class="identity">
          <div class="field">
            <label for="sf-name">Name <span class="muted">(optional if anonymous)</span></label>
            <input
              id="sf-name"
              ref="firstFieldEl"
              v-model="form.name"
              type="text"
              autocomplete="name"
              :aria-invalid="!!fieldErrors.name"
              maxlength="200"
            />
            <p v-if="fieldErrors.name" class="err small">{{ fieldErrors.name }}</p>
          </div>

          <div class="field">
            <label for="sf-email">Email <span class="req">required</span></label>
            <input
              id="sf-email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              :aria-invalid="!!fieldErrors.email"
              maxlength="254"
            />
            <p v-if="fieldErrors.email" class="err small">{{ fieldErrors.email }}</p>
          </div>
        </div>

        <div class="anon">
          <label class="checkbox">
            <input v-model="form.anonymous" type="checkbox" />
            <span>Submit anonymously</span>
          </label>
          <p class="small muted anon-info">
            If you choose not to remain anonymous, you agree that we may contact you
            by email with questions about your answers. If you remain anonymous,
            your name will not be displayed and your email will only be used for
            deduplication, never for contact.
          </p>
        </div>

        <!-- Honeypot: visually hidden and off the tab order; bots fill it, humans don't. -->
        <div class="hp" aria-hidden="true">
          <label>Website<input v-model="form.website" type="text" tabindex="-1" autocomplete="off" /></label>
        </div>

        <fieldset class="matrix">
          <legend>
            Familiarity by research area
            <span class="small muted">({{ answeredCount }} of {{ areas.length }} answered)</span>
          </legend>
          <p v-if="fieldErrors.ratings" class="err small">{{ fieldErrors.ratings }}</p>

          <div class="scale-key small muted" aria-hidden="true">
            <span v-for="lvl in FAMILIARITY_LEVELS" :key="lvl.value">
              <strong>{{ lvl.value }}</strong> {{ lvl.short }}
            </span>
            <span><strong>—</strong> Skip</span>
          </div>

          <div v-for="group in groups" :key="group.name" class="matrix-group">
            <p class="matrix-group-title">{{ group.name }}</p>
            <div v-for="a in group.areas" :key="a.tag" class="matrix-row">
              <span class="matrix-label">
                <span class="tag">{{ a.tag }}</span>{{ a.name }}
              </span>
              <div class="radios" role="radiogroup" :aria-label="`Familiarity with ${a.name}`">
                <label
                  v-for="opt in [null, 0, 1, 2, 3]"
                  :key="String(opt)"
                  class="radio"
                  :class="{ active: ratings[a.tag] === opt }"
                >
                  <input
                    type="radio"
                    :name="`r-${a.tag}`"
                    :checked="ratings[a.tag] === opt"
                    @change="ratings[a.tag] = opt"
                  />
                  <span>{{ opt === null ? '—' : opt }}</span>
                </label>
              </div>
            </div>
          </div>
        </fieldset>

        <div class="dedup notice small">
          Submitting again with the same email address will replace your previous
          response.
        </div>

        <div class="actions">
          <button type="button" class="btn" @click="emit('close')">Cancel</button>
          <button type="submit" class="btn btn-primary" :disabled="submitting">
            {{ submitting ? 'Submitting…' : 'Submit response' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(10, 10, 12, 0.55);
  display: grid;
  place-items: start center;
  padding: 24px 16px;
  overflow-y: auto;
  z-index: 50;
}
.dialog {
  width: 100%;
  max-width: 760px;
  padding: 0;
}
.dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 22px 14px;
  border-bottom: 1px solid var(--border);
}
.dialog-head h2 {
  margin: 0 0 2px;
  font-size: 1.3rem;
}
.dialog-body {
  padding: 18px 22px 22px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.identity {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.field label {
  font-weight: 500;
  font-size: 0.9rem;
}
.req {
  color: var(--danger);
  font-size: 0.78rem;
  font-weight: 600;
}
input[type='text'],
input[type='email'] {
  padding: 9px 11px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  background: var(--surface-0);
  color: var(--text-primary);
  font-size: 0.95rem;
}
input[aria-invalid='true'] {
  border-color: var(--danger);
}
.err {
  color: var(--danger);
  margin: 0;
}
.anon {
  background: var(--surface-2);
  border-radius: 7px;
  padding: 12px 14px;
}
.checkbox {
  display: flex;
  align-items: center;
  gap: 9px;
  font-weight: 500;
  cursor: pointer;
}
.anon-info {
  margin: 8px 0 0;
  max-width: 68ch;
}
.hp {
  position: absolute;
  left: -9999px;
  width: 1px;
  height: 1px;
  overflow: hidden;
}
.matrix {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 14px 16px 6px;
  margin: 0;
}
.matrix legend {
  font-weight: 600;
  padding: 0 6px;
}
.scale-key {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  margin-bottom: 12px;
}
.matrix-group-title {
  margin: 14px 0 6px;
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}
.matrix-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 7px 0;
  border-top: 1px solid var(--border);
}
.matrix-label {
  font-size: 0.92rem;
}
.tag {
  display: inline-grid;
  place-items: center;
  min-width: 28px;
  height: 20px;
  padding: 0 5px;
  margin-right: 9px;
  border-radius: 4px;
  background: var(--surface-0);
  border: 1px solid var(--border-strong);
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-secondary);
  vertical-align: middle;
}
.radios {
  display: flex;
  gap: 5px;
  flex: none;
}
.radio {
  position: relative;
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border-radius: 6px;
  border: 1px solid var(--border-strong);
  background: var(--surface-1);
  cursor: pointer;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: var(--text-secondary);
}
.radio input {
  position: absolute;
  opacity: 0;
  inset: 0;
  margin: 0;
  cursor: pointer;
}
.radio.active {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--accent-ink);
}
.radio:has(input:focus-visible) {
  outline: 2px solid var(--focus);
  outline-offset: 2px;
}
.dedup {
  color: var(--text-secondary);
}
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
@media (max-width: 560px) {
  .identity {
    grid-template-columns: 1fr;
  }
  .matrix-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
}
</style>
