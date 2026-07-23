<script setup lang="ts">
/**
 * The one write path in the whole system.
 *
 * What it asks for, and why:
 *  · A MATURITY RATING for the research area, required, in the same six tiers the
 *    map itself uses. This replaced "do you agree with our rating?", which bought
 *    one bit and anchored the reader on ours. A rating in the map's own vocabulary
 *    can be aggregated, disagreed with, and drawn, the community view of the grid
 *    is built from exactly this field.
 *  · Optionally, a rating for the individual AGENDAS inside that area, for readers
 *    who want to be more precise than "the area as a whole".
 *  · Familiarity, so the ratings can be read with the right weight.
 *  · Open comments, because the reasoning is worth more than the number.
 *
 * Maturity here is a property of the research, how much work exists, how well it
 * was done, what it has actually shown, not of how much of a problem it solves.
 * That is why the vote is scoped to an area even when the form was opened from one
 * cell of the grid.
 *
 * Design rules, in order of importance:
 *  · Anonymous is the default. Name, email and contact consent only appear once
 *    the reader turns anonymity off, and the server drops them anyway if the
 *    anonymous flag is set.
 *  · Consent is opt-in, in plain words, and visible before the submit button.
 *  · The identity is remembered locally so answering a second area asks only for
 *    the ratings.
 *  · A sleeping backend is a state to explain, not an error to blame the reader
 *    for.
 */
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { agendasReaching, atlas, tierOrder } from '@/lib/atlas'
import { ApiError, apiConfigured, submitFeedback, type FeedbackPayload } from '@/lib/api'
import { tierAbbrev, tierStyle } from '@/lib/scales'
import { useSubmitter } from '@/composables/useSubmitter'
import { noteLocalSubmission, refreshSummary, useSummary } from '@/composables/useSummary'
import CommunityDistribution from '@/components/CommunityDistribution.vue'
import type { FeedbackSubject, Tier } from '@/types/atlas'

const props = defineProps<{ subject: FeedbackSubject }>()

const { profile, displayName, hasSubmitted, markSubmitted, reset } = useSubmitter()
const { forArea, forAgenda } = useSummary()

/** The research area is the subject of the vote, whichever page asked for it. */
const area = computed(() => {
  const s = props.subject
  return s.kind === 'agenda'
    ? { tag: s.agenda.area_tag, name: s.agenda.area_name }
    : { tag: s.areaTag, name: s.areaName }
})

/**
 * The agendas the reader may also rate. From an agenda page that is the one
 * agenda; from a cell it is the agendas of that area which reach that problem.
 */
const subareas = computed(() => {
  const s = props.subject
  if (s.kind === 'agenda') return [{ id: s.agenda.id, name: s.agenda.agenda }]
  return agendasReaching(s.areaName, s.problemId).map((row) => ({ id: row.agenda.id, name: row.agenda.agenda }))
})

/** Stable local key for "have I submitted this?", agenda id, or "AREA:PROBLEM". */
const subjectId = computed(() =>
  props.subject.kind === 'agenda' ? props.subject.agenda.id : `${props.subject.areaTag}:${props.subject.problemId}`,
)

const context = computed(() => {
  const s = props.subject
  return s.kind === 'agenda'
    ? `You opened this from ${s.agenda.id}, so you can rate that agenda specifically too.`
    : `You opened this from ${s.areaTag} × ${s.problemId}, so you can rate the agendas in that cell specifically too.`
})

/** Just the target fields of the payload; the rest is shared. */
function payloadTarget(): Pick<FeedbackPayload, 'agenda_id' | 'area_tag' | 'problem_id'> {
  const s = props.subject
  // The area is not sent for an agenda: the server derives it from the id, so
  // which area a vote lands in is never the client's choice.
  return s.kind === 'agenda' ? { agenda_id: s.agenda.id } : { area_tag: s.areaTag, problem_id: s.problemId }
}

const areaMaturity = ref<Tier | ''>('')
const subareaMaturity = ref<Record<string, Tier | ''>>({})
const familiarity = ref<number | null>(null)
const notes = ref('')
const website = ref('') // honeypot, must stay empty
const editingIdentity = ref(false)

type Status = 'idle' | 'submitting' | 'waking' | 'done' | 'error'
const status = ref<Status>('idle')
const errorMessage = ref('')
const updating = ref(false)

const turnstileKey = (import.meta.env.VITE_TURNSTILE_SITE_KEY ?? '').trim()
const turnstileHost = ref<HTMLDivElement | null>(null)

const alreadySubmitted = computed(() => hasSubmitted(subjectId.value) && !updating.value && status.value !== 'done')
const notesLeft = computed(() => 2000 - notes.value.length)
const canSubmit = computed(
  () =>
    areaMaturity.value !== '' &&
    familiarity.value !== null &&
    status.value !== 'submitting' &&
    status.value !== 'waking',
)

const tierOptions = computed(() =>
  tierOrder.map((tier) => ({ tier, definition: atlas.legend.tiers[tier] ?? '' })),
)

const familiarityOptions = computed(() =>
  (['0', '1', '2', '3', '4'] as const).map((level) => ({
    value: Number(level),
    label: atlas.legend.familiarity_scale[level] ?? level,
  })),
)

const areaStats = computed(() => forArea(area.value.tag))
const agendaStats = computed(() =>
  props.subject.kind === 'agenda' ? forAgenda(props.subject.agenda.id) : undefined,
)

// A different cell or agenda is a different question: never carry an answer over.
watch(subjectId, () => {
  areaMaturity.value = ''
  subareaMaturity.value = {}
  familiarity.value = null
  notes.value = ''
  status.value = 'idle'
  updating.value = false
})

onMounted(() => {
  if (!turnstileKey) return
  // Loaded only when a site key is configured, so the default build talks to
  // nobody but the API.
  const existing = document.querySelector('script[data-turnstile]')
  if (existing) return
  const script = document.createElement('script')
  script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js'
  script.async = true
  script.defer = true
  script.dataset.turnstile = 'true'
  document.head.appendChild(script)
})

function turnstileToken(): string {
  const input = turnstileHost.value?.querySelector<HTMLInputElement>('input[name="cf-turnstile-response"]')
  return input?.value ?? ''
}

/**
 * Server phrasing is written for the operator; readers get a sentence that says
 * what happened to their answer. Validation messages are passed through, because
 * those name a field the reader can actually fix.
 */
function readableError(err: unknown): string {
  if (!(err instanceof ApiError)) {
    return 'Something went wrong sending your answer. Nothing was saved, please try again.'
  }
  switch (err.status) {
    case 400:
      return err.field ? `${err.field}: ${err.message}` : err.message
    case 429:
      return 'That is a lot of feedback in a short time, please wait a minute and send this one again.'
    case 503:
      return 'The feedback service is not accepting submissions right now. Your answer was not saved; please try again later.'
    default:
      return err.message
  }
}

/** Agendas the reader left alone are not votes, so they are not sent as any. */
function answeredSubareas(): Record<string, string> | undefined {
  const answered = Object.entries(subareaMaturity.value).filter(([, tier]) => tier !== '')
  return answered.length ? Object.fromEntries(answered) : undefined
}

async function onSubmit(): Promise<void> {
  if (areaMaturity.value === '' || familiarity.value === null) return

  status.value = 'submitting'
  errorMessage.value = ''

  const anonymous = profile.isAnonymous

  try {
    await submitFeedback(
      {
        ...payloadTarget(),
        area_maturity: areaMaturity.value,
        subarea_maturity: answeredSubareas(),
        familiarity: familiarity.value,
        notes: notes.value.trim() || undefined,
        name: anonymous ? undefined : profile.name.trim() || undefined,
        email: anonymous ? undefined : profile.email.trim() || undefined,
        is_anonymous: anonymous,
        contact_consent: anonymous ? false : profile.contactConsent,
        reuse_consent: profile.reuseConsent,
        submitter_token: profile.token,
        turnstile_token: turnstileKey ? turnstileToken() : undefined,
        website: website.value,
      },
      () => {
        status.value = 'waking'
      },
    )

    markSubmitted(subjectId.value)
    noteLocalSubmission(subjectId.value)
    status.value = 'done'
    updating.value = false
    void refreshSummary()
  } catch (err) {
    status.value = 'error'
    errorMessage.value = readableError(err)
  }
}

function startUpdate(): void {
  updating.value = true
  status.value = 'idle'
}

function submitAnother(): void {
  status.value = 'idle'
  areaMaturity.value = ''
  subareaMaturity.value = {}
  familiarity.value = null
  notes.value = ''
}
</script>

<template>
  <section class="fb card" aria-labelledby="fb-title">
    <h3 id="fb-title" class="fb__title">Your reading of {{ area.name }}</h3>
    <p class="fb__lead">
      This map is a draft put out for correction. Rate how mature you think this body of research is:
      how much work exists, how well it has been done, and what it has actually shown. That is a
      judgment about the research itself, not about how much of any one problem it solves.
    </p>

    <p v-if="!apiConfigured" class="fb__notice">
      Feedback collection is switched off on this deployment (no API is configured). Everything else
      on the page works normally.
    </p>

    <div v-else-if="status === 'done'" class="fb__done" role="status">
      <p><strong>Recorded, thank you.</strong> Your rating is in the community view of the grid now.</p>
      <div class="fb__doneactions">
        <button type="button" class="btn" @click="submitAnother">Change my answer</button>
        <RouterLink class="btn btn--quiet" to="/">Back to the map</RouterLink>
      </div>
    </div>

    <div v-else-if="alreadySubmitted" class="fb__done">
      <p>You have already answered this from this browser.</p>
      <button type="button" class="btn" @click="startUpdate">Submit an update</button>
    </div>

    <template v-else>
      <CommunityDistribution
        class="fb__dist"
        :subject="area.name"
        :stats="areaStats"
        empty-note="Nobody has rated this area yet. Yours would be the first. The histograms appear here once there are answers."
      />
      <CommunityDistribution
        v-if="subject.kind === 'agenda' && agendaStats"
        class="fb__dist"
        :subject="subject.agenda.id"
        :stats="agendaStats"
      />

      <form class="fb__form" novalidate @submit.prevent="onSubmit">
        <fieldset class="fs">
          <legend class="fs__legend">
            How mature is the research in {{ area.name }} as a whole?
            <span class="req">required</span>
          </legend>
          <div class="opts">
            <label
              v-for="opt in tierOptions"
              :key="opt.tier"
              class="opt"
              :class="{ 'opt--on': areaMaturity === opt.tier }"
            >
              <input v-model="areaMaturity" type="radio" name="area-maturity" :value="opt.tier" />
              <span
                class="opt__chip"
                :class="{ hatched: tierStyle(opt.tier).hatched }"
                :style="{ backgroundColor: tierStyle(opt.tier).fill, color: tierStyle(opt.tier).ink }"
                >{{ tierAbbrev(opt.tier) }}</span
              >
              <span class="opt__body">
                <span class="opt__label">{{ opt.tier }}</span>
                <span class="opt__def">{{ opt.definition }}</span>
              </span>
            </label>
          </div>
        </fieldset>

        <fieldset v-if="areaMaturity" class="fs">
          <legend class="fs__legend">
            And the individual agendas? <span class="opt-note">optional</span>
          </legend>
          <p class="fs__hint">
            {{ context }} Leave any of them blank. A blank is not a rating of "no opinion"; it is
            simply not counted.
          </p>
          <div class="subs">
            <div v-for="sub in subareas" :key="sub.id" class="sub">
              <label :for="`sub-${sub.id}`" class="sub__label">
                <span class="sub__id tabular">{{ sub.id }}</span>
                <span class="sub__name">{{ sub.name }}</span>
              </label>
              <select :id="`sub-${sub.id}`" v-model="subareaMaturity[sub.id]" class="sub__select">
                <option value="">No rating</option>
                <option v-for="tier in tierOrder" :key="tier" :value="tier">{{ tier }}</option>
              </select>
            </div>
          </div>
        </fieldset>

        <fieldset class="fs">
          <legend class="fs__legend">
            How familiar are you with {{ area.name }}?
            <span class="req">required</span>
          </legend>
          <div class="opts">
            <label
              v-for="opt in familiarityOptions"
              :key="opt.value"
              class="opt"
              :class="{ 'opt--on': familiarity === opt.value }"
            >
              <input v-model="familiarity" type="radio" name="familiarity" :value="opt.value" />
              <span class="opt__num tabular">{{ opt.value }}</span>
              <span class="opt__label">{{ opt.label }}</span>
            </label>
          </div>
        </fieldset>

        <div class="fs">
          <label for="fb-notes">Further comments <span class="opt-note">optional</span></label>
          <p class="fs__hint">
            The reasoning behind your ratings, work we have missed, a rating you think is wrong, or
            anything else worth putting on the record.
          </p>
          <textarea
            id="fb-notes"
            v-model="notes"
            maxlength="2000"
            placeholder="e.g. 'The Contested rating rests on one paper, and the replication attempt cuts the other way.'"
          ></textarea>
          <p class="field-hint tabular">{{ notesLeft }} characters left</p>
        </div>

        <div class="fs identity">
          <label class="checkline">
            <input v-model="profile.isAnonymous" type="checkbox" />
            <span>
              <strong>Submit anonymously.</strong> No name, no email, nothing that identifies you, just
              the answer itself.
            </span>
          </label>

          <div v-if="!profile.isAnonymous" class="identity__fields">
            <div class="grid2">
              <div>
                <label for="fb-name">Name or nickname <span class="opt-note">optional</span></label>
                <input id="fb-name" v-model="profile.name" type="text" autocomplete="nickname" placeholder="a pseudonym is fine" />
                <p class="field-hint">A handle is welcome; it's only how we'd credit or refer to you, never shown publicly unless you allow reuse.</p>
              </div>
              <div>
                <label for="fb-email">Email <span class="opt-note">optional</span></label>
                <input id="fb-email" v-model="profile.email" type="email" autocomplete="email" />
                <p class="field-hint">Only used to follow up if you tick the box below. Never published.</p>
              </div>
            </div>

            <label class="checkline">
              <input
                v-model="profile.contactConsent"
                type="checkbox"
                @change="profile.reuseConsent = profile.contactConsent || profile.reuseConsent"
              />
              <span>
                I agree that you may contact me with follow-up questions about my answers, and may use
                my responses (attributed to my name/nickname, or not) in future iterations of this map.
              </span>
            </label>
          </div>

          <label v-else class="checkline">
            <input v-model="profile.reuseConsent" type="checkbox" />
            <span>You may quote or summarise this response in a future iteration of the map.</span>
          </label>

          <p class="fb__privacy">
            Submitting as <strong>{{ displayName }}</strong
            >.
            <button v-if="!editingIdentity" type="button" class="btn btn--quiet" @click="editingIdentity = true">
              change
            </button>
            <button v-else type="button" class="btn btn--quiet" @click="reset(); editingIdentity = false">
              forget me on this device
            </button>
            <br />
            We store your answer, your consent choices and a random id for this browser. No IP address
            is kept, only a salted hash. <RouterLink to="/privacy">What we keep and why</RouterLink>.
          </p>
        </div>

        <!-- Honeypot: off-screen, not focusable, never filled by a human. -->
        <div class="hp" aria-hidden="true">
          <label for="fb-website">Leave this field empty</label>
          <input id="fb-website" v-model="website" type="text" tabindex="-1" autocomplete="off" />
        </div>

        <div v-if="turnstileKey" ref="turnstileHost" class="cf-turnstile" :data-sitekey="turnstileKey"></div>

        <div class="fb__actions">
          <button type="submit" class="btn btn--primary" :disabled="!canSubmit">
            {{ status === 'submitting' ? 'Sending…' : status === 'waking' ? 'Waking the server…' : 'Send feedback' }}
          </button>
          <p v-if="status === 'waking'" class="fb__waking">
            The API sleeps when nobody is using it and takes up to a minute to wake. Hold on, your
            answer is still queued.
          </p>
          <p v-if="!canSubmit && status === 'idle'" class="field-hint">
            A maturity rating and a familiarity level are needed before this can be sent.
          </p>
        </div>

        <p v-if="status === 'error'" class="fb__error" role="alert">{{ errorMessage }}</p>
      </form>
    </template>
  </section>
</template>

<style scoped>
.fb {
  padding: 1.25rem 1.35rem 1.4rem;
}

.fb__title {
  margin-bottom: 0.3rem;
}

.fb__lead {
  color: var(--ink-secondary);
  font-size: 0.9375rem;
}

.fb__dist {
  margin-bottom: 1rem;
}

.fb__notice,
.fb__done {
  background: var(--surface-sunken);
  border-radius: var(--radius);
  padding: 0.9rem 1rem;
  font-size: 0.9375rem;
  color: var(--ink-secondary);
}

.fb__done p {
  margin: 0 0 0.6rem;
}

.fb__doneactions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.fb__form {
  display: grid;
  gap: 1.35rem;
}

.fs {
  border: 0;
  padding: 0;
  margin: 0;
}

.fs__legend {
  font-size: 0.9375rem;
  font-weight: 500;
  padding: 0;
  margin-bottom: 0.5rem;
}

.fs__hint {
  font-size: 0.8125rem;
  color: var(--ink-muted);
  margin: -0.25rem 0 0.6rem;
  max-width: var(--measure);
}

.req {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--ink-muted);
  font-weight: 600;
  margin-left: 0.3rem;
}

.opt-note {
  font-size: 0.6875rem;
  color: var(--ink-muted);
  font-weight: 400;
  margin-left: 0.25rem;
}

.opts {
  display: grid;
  gap: 0.4rem;
}

.opt {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  padding: 0.5rem 0.7rem;
  cursor: pointer;
  font-weight: 400;
  font-size: 0.875rem;
  background: var(--surface);
}

.opt:hover {
  border-color: var(--rule-strong);
}

.opt--on {
  border-color: var(--ink);
  box-shadow: inset 0 0 0 1px var(--ink);
}

.opt input {
  margin-top: 0.15rem;
  accent-color: var(--focus);
}

.opt__num {
  font-weight: 700;
  color: var(--ink-muted);
  flex: none;
}

.opt__chip {
  flex: none;
  align-self: center;
  min-width: 1.6rem;
  text-align: center;
  border-radius: 3px;
  padding: 0.05rem 0.3rem;
  font-size: 0.5625rem;
  font-weight: 700;
}

.opt__body {
  min-width: 0;
}

.opt__label {
  display: block;
  font-weight: 500;
}

.opt__def {
  display: block;
  font-size: 0.8125rem;
  color: var(--ink-muted);
}

.subs {
  display: grid;
  gap: 0.4rem;
}

.sub {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.75rem;
  align-items: center;
  justify-content: space-between;
  border: 1px solid var(--rule);
  border-radius: var(--radius);
  padding: 0.45rem 0.7rem;
  background: var(--surface);
}

.sub__label {
  display: flex;
  align-items: baseline;
  gap: 0.45rem;
  min-width: 0;
  font-weight: 400;
  font-size: 0.875rem;
  flex: 1 1 12rem;
}

.sub__id {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--ink-muted);
  flex: none;
}

.sub__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sub__select {
  flex: none;
  max-width: 100%;
  padding: 0.3rem 0.4rem;
  border: 1px solid var(--rule-strong);
  border-radius: var(--radius);
  background: var(--surface);
  color: var(--ink);
  font: inherit;
  font-size: 0.8125rem;
}

.identity {
  border-top: 1px solid var(--rule);
  padding-top: 1.1rem;
  display: grid;
  gap: 0.85rem;
}

.identity__fields {
  display: grid;
  gap: 0.85rem;
}

.grid2 {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
}

.fb__privacy {
  font-size: 0.8125rem;
  color: var(--ink-muted);
  margin: 0;
  max-width: var(--measure);
}

.hp {
  position: absolute;
  left: -9999px;
  width: 1px;
  height: 1px;
  overflow: hidden;
}

.fb__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}

.fb__waking,
.fb__error {
  font-size: 0.8125rem;
  margin: 0;
  max-width: var(--measure);
}

.fb__waking {
  color: var(--ink-muted);
}

.fb__error {
  color: var(--ink);
  background: var(--surface-sunken);
  border-left: 3px solid var(--tier-contested);
  padding: 0.6rem 0.8rem;
  border-radius: 0 var(--radius) var(--radius) 0;
}
</style>
