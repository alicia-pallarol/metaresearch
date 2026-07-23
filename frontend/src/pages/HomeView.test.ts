/**
 * Graceful degradation is a requirement, not a nicety: free backends sleep, and
 * the map has to be fully readable while that happens. This mounts the landing
 * page with every network call failing.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import HomeView from './HomeView.vue'
import { refreshSummary } from '@/composables/useSummary'

function mountHome() {
  return mount(HomeView, {
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('HomeView with the backend unreachable', () => {
  it('still renders the whole map', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.reject(new Error('backend asleep'))),
    )
    await refreshSummary()

    const wrapper = mountHome()

    // The condensed grid: 12 problems × 12 areas of real cells.
    expect(wrapper.findAll('[role="gridcell"]')).toHaveLength(144)
    // The focus-areas section (computed, both units) and the full grid are there too.
    expect(wrapper.text()).toContain('Focus areas')
    expect(wrapper.text()).toContain('Research areas')
    expect(wrapper.text()).toContain('Every agenda, every problem')
    // And the hero degrades to a stated absence rather than a broken number.
    expect(wrapper.text()).toContain('not available right now')
  })

  it('opens a drill-down when a cell is activated', async () => {
    const wrapper = mountHome()

    const cells = wrapper.findAll('[role="gridcell"]')
    // P1 × Interpretability is the second row's Interpretability column; any
    // non-blank cell will do, so find the first one with agendas behind it.
    const populated = cells.find((cell) => (cell.text() ?? '').trim().length > 0)
    expect(populated).toBeDefined()

    await populated!.trigger('click')

    expect(wrapper.text()).toContain('Drill-down')
  })

  it('hides the Swiss-cheese overlay while no layers are authored', () => {
    const wrapper = mountHome()
    expect(wrapper.text()).not.toContain('Defence in depth')
  })
})
