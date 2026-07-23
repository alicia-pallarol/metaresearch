/**
 * Dataset integrity.
 *
 * data/atlas.json is hand-edited when a new iteration is cut, so these are the
 * checks that catch a slip: wrong counts, a tier that is not in tier_order, a
 * condensed cell that no longer matches the agenda-level grid it summarises.
 */
import { describe, expect, it } from 'vitest'
import { atlas, agendasOfArea, agendasReaching, cellTier, coverageOf, problemIds, tierRank } from './atlas'
import type { Tier } from '@/types/atlas'

describe('atlas dataset', () => {
  it('has the expected shape', () => {
    expect(atlas.areas).toHaveLength(12)
    expect(atlas.problems).toHaveLength(12)
    expect(atlas.agendas).toHaveLength(58)
    expect(atlas.priorities.items).toHaveLength(6)
    expect(problemIds).toHaveLength(12)
  })

  it('has a full 58 × 12 matrix', () => {
    expect(Object.keys(atlas.matrix)).toHaveLength(58)
    for (const [agendaId, row] of Object.entries(atlas.matrix)) {
      expect(Object.keys(row).sort()).toEqual([...problemIds].sort())
      expect(atlas.agendas.some((a) => a.id === agendaId)).toBe(true)
    }
  })

  it('only uses tiers declared in tier_order', () => {
    const allowed = new Set<string>(atlas.legend.tier_order)
    for (const row of Object.values(atlas.matrix)) {
      for (const tier of Object.values(row)) {
        if (tier !== null) expect(allowed.has(tier)).toBe(true)
      }
    }
    for (const agenda of atlas.agendas) {
      expect(allowed.has(agenda.overall_maturity)).toBe(true)
    }
  })

  it('lists every agenda under exactly one area', () => {
    const fromAreas = atlas.areas.flatMap((a) => a.agenda_ids).sort()
    const fromAgendas = atlas.agendas.map((a) => a.id).sort()
    expect(fromAreas).toEqual(fromAgendas)
    expect(new Set(fromAreas).size).toBe(fromAgendas.length)
  })

  it('derives every condensed cell from the agenda-level grid', () => {
    // The condensed view claims "best tier, and how many agendas reach it".
    // If that stops being true the hero visual is quietly lying.
    for (const problem of atlas.problems) {
      for (const areaName of atlas.area_matrix.area_order) {
        const tiers = agendasOfArea(areaName)
          .map((a) => cellTier(a.id, problem.id))
          .filter((t): t is Tier => t !== null)

        const cell = atlas.area_matrix.cells[problem.id]?.[areaName] ?? null

        if (tiers.length === 0) {
          expect(cell).toBeNull()
          continue
        }
        const best = tiers.reduce((a, b) => (tierRank(a) <= tierRank(b) ? a : b))
        expect(cell).toEqual({ tier: best, count: tiers.length })
      }
    }
  })

  it('points every priority item at a real problem', () => {
    for (const item of atlas.priorities.items) {
      expect(atlas.problems.some((p) => p.id === item.problem_id)).toBe(true)
    }
  })
})

describe('derived lookups', () => {
  it('sorts drill-down rows strongest first', () => {
    const rows = agendasReaching('Interpretability', 'P1')
    expect(rows.length).toBeGreaterThan(0)
    const ranks = rows.map((r) => tierRank(r.tier))
    expect([...ranks]).toEqual([...ranks].sort((a, b) => a - b))
  })

  it('counts coverage as the number of non-blank cells in a row', () => {
    const manual = problemIds.filter((p) => cellTier('IN6', p) !== null).length
    expect(coverageOf('IN6')).toBe(manual)
    expect(coverageOf('does-not-exist')).toBe(0)
  })
})
