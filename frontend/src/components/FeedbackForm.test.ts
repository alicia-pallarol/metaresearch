/**
 * The feedback form is the only thing on the site that sends anything anywhere,
 * so its defaults are tested rather than trusted: anonymous unless asked
 * otherwise, no identity fields until anonymity is turned off, a mandatory
 * maturity vote, and a honeypot that stays empty.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import FeedbackForm from './FeedbackForm.vue'
import { getAgenda } from '@/lib/atlas'
import { useSubmitter } from '@/composables/useSubmitter'
import type { Agenda } from '@/types/atlas'

const agenda = getAgenda('IN6') as Agenda

function mountForm() {
  return mount(FeedbackForm, {
    props: { subject: { kind: 'agenda', agenda } },
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

function mountCellForm() {
  return mount(FeedbackForm, {
    props: {
      subject: { kind: 'cell', areaTag: 'IN', areaName: 'Interpretability', problemId: 'P4', problemName: 'Deceptive alignment and scheming' },
    },
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

beforeEach(() => {
  // The profile is a module-level singleton (one identity per browser tab), so
  // clearing storage is not enough, reset the live object between tests.
  useSubmitter().reset()
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('FeedbackForm', () => {
  it('defaults to anonymous and asks for nothing identifying', () => {
    const wrapper = mountForm()

    expect(wrapper.find('#fb-name').exists()).toBe(false)
    expect(wrapper.find('#fb-email').exists()).toBe(false)
    expect(wrapper.text()).toContain('Submit anonymously')
  })

  it('only shows name, email and contact consent once anonymity is switched off', async () => {
    const wrapper = mountForm()

    const anonymousToggle = wrapper.findAll('input[type="checkbox"]')[0]
    expect(anonymousToggle).toBeDefined()
    await anonymousToggle!.setValue(false)

    expect(wrapper.find('#fb-name').exists()).toBe(true)
    expect(wrapper.find('#fb-email').exists()).toBe(true)
    expect(wrapper.text()).toContain('you may contact me with follow-up questions')
  })

  it('needs both a maturity vote and a familiarity level before it can be sent', async () => {
    const wrapper = mountForm()
    const submit = () => wrapper.find('button[type="submit"]')

    // Nothing chosen yet.
    expect(submit().attributes('disabled')).toBeDefined()

    // Familiarity alone is not enough, the maturity vote is the point of the form.
    await wrapper.find('input[name="familiarity"]').setValue()
    expect(submit().attributes('disabled')).toBeDefined()

    // Maturity as well: now it sends.
    await wrapper.find('input[name="area-maturity"]').setValue()
    expect(submit().attributes('disabled')).toBeUndefined()
  })

  it('asks for the maturity of the research area, in the map’s own tiers', () => {
    const wrapper = mountForm()
    // IN6 belongs to Interpretability, so the vote is about the area.
    expect(wrapper.text()).toContain('How mature is the research in Interpretability as a whole?')
    const maturity = wrapper.findAll('input[name="area-maturity"]')
    expect(maturity).toHaveLength(6) // the six tiers of tier_order
  })

  it('no longer asks whether you agree with our rating', () => {
    const wrapper = mountForm()
    expect(wrapper.text()).not.toContain('Does our rating look right?')
    expect(wrapper.text()).not.toContain('agree')
  })

  it('carries a honeypot field that a human never fills', () => {
    const wrapper = mountForm()
    const honeypot = wrapper.find('#fb-website')

    expect(honeypot.exists()).toBe(true)
    expect(honeypot.attributes('tabindex')).toBe('-1')
    expect((honeypot.element as HTMLInputElement).value).toBe('')
  })

  it('offers a reuse consent that anonymous submitters can give too', () => {
    const wrapper = mountForm()
    expect(wrapper.text()).toContain('quote or summarise this response')
  })

  it('welcomes a nickname, not just a real name', async () => {
    const wrapper = mountForm()
    const anonymousToggle = wrapper.findAll('input[type="checkbox"]')[0]
    await anonymousToggle!.setValue(false)
    expect(wrapper.text()).toContain('Name or nickname')
  })
})

describe('FeedbackForm targeting a cell', () => {
  it('votes on the whole research area, and lets you rate the agendas in the cell', () => {
    const wrapper = mountCellForm()

    // The vote is about the area, however the form was opened.
    expect(wrapper.text()).toContain('How mature is the research in Interpretability as a whole?')
    // Familiarity and the free-text comments stay.
    expect(wrapper.text()).toContain('How familiar are you')
    expect(wrapper.find('#fb-notes').exists()).toBe(true)
    // Titled by the area whose maturity is being rated.
    expect(wrapper.text()).toContain('Your reading of Interpretability')
  })

  it('offers the agendas of the cell as optional per-agenda ratings once the area is rated', async () => {
    const wrapper = mountCellForm()

    // The optional subarea block only appears after the required area vote.
    expect(wrapper.find('#sub-IN6').exists()).toBe(false)
    await wrapper.find('input[name="area-maturity"]').setValue()
    // IN6 reaches P4, so it is offered as a per-agenda rating.
    expect(wrapper.find('#sub-IN6').exists()).toBe(true)
  })
})
