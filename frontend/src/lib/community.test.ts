/**
 * The band arithmetic is the load-bearing new idea, so it is pinned here.
 *
 * The two rules it must obey (see lib/community.ts):
 *   1. a band is summarised by its MEDIAN, never a mean;
 *   2. where a band has no single middle, the WEAKER of the two shows.
 * And the invariant that matters for continuity: research + best case must
 * reproduce the "best of N agendas" grid this map has always shown.
 */
import { describe, expect, it } from 'vitest'
import {
  agendaReading,
  bandTier,
  gridCell,
  numericStats,
  tierCountsFrom,
  tierStats,
} from '@/lib/community'
import { agendasReaching, atlas, getAreaByName, tierRank } from '@/lib/atlas'
import type { Summary, Tier } from '@/types/atlas'

function counts(list: Tier[]) {
  return tierCountsFrom(list)
}

describe('bandTier', () => {
  it('is null with nothing to summarise', () => {
    expect(bandTier({}, 'typical')).toBeNull()
  })

  it('best case takes the strongest quarter, worst case the weakest', () => {
    // Eight ratings, strongest→weakest: R S S E E U U C
    const c = counts([
      'Robust (small scale)',
      'Strong existence proof',
      'Strong existence proof',
      'Early / partial',
      'Early / partial',
      'Untested',
      'Untested',
      'Contested',
    ])
    // Quarter of 8 = 2. Top two are R,S → median lands on the weaker, S.
    expect(bandTier(c, 'best')).toBe('Strong existence proof')
    // Bottom two are U,C → weaker is C.
    expect(bandTier(c, 'worst')).toBe('Contested')
  })

  it('takes the weaker of two middles for an even typical band', () => {
    // Two ratings: Strong, then Early. No single middle → the weaker, Early.
    expect(bandTier(counts(['Strong existence proof', 'Early / partial']), 'typical')).toBe('Early / partial')
  })

  it('below the threshold a band is a single rating, so best = strongest', () => {
    // Three ratings, a quarter is one, so best case is exactly the strongest.
    const c = counts(['Early / partial', 'Untested', 'Contested'])
    expect(bandTier(c, 'best')).toBe('Early / partial')
    expect(bandTier(c, 'worst')).toBe('Contested')
  })

  it('never invents a tier: an off-axis-heavy cell reports an off-axis tier', () => {
    // Three Contested and one Robust must not average into the middle of the
    // ladder, the typical reading is Contested, a tier that was actually cast.
    const c = counts(['Robust (small scale)', 'Contested', 'Contested', 'Contested'])
    expect(bandTier(c, 'typical')).toBe('Contested')
  })
})

describe('research + best case reproduces the curated grid', () => {
  it('matches the best of N agendas for every cell', () => {
    for (const problem of atlas.problems) {
      for (const areaName of atlas.area_matrix.area_order) {
        const tag = getAreaByName(areaName)?.tag ?? ''
        const reaching = agendasReaching(areaName, problem.id)
        const cell = gridCell('research', 'best', tag, areaName, problem.id, null)

        if (reaching.length === 0) {
          expect(cell.tier).toBeNull()
          continue
        }
        // agendasReaching is sorted strongest first, so [0] is the best tier.
        const best = reaching.reduce((a, b) => (tierRank(a.tier) <= tierRank(b.tier) ? a : b))
        expect(cell.tier).toBe(best.tier)
        expect(cell.agendaCount).toBe(reaching.length)
      }
    }
  })
})

describe('community readings', () => {
  const summary: Summary = {
    iteration: 0,
    totals: { total_submissions: 3, unique_submitters: 3 },
    per_area: {
      IN: { count: 3, maturity: { 'Early / partial': 2, Untested: 1 }, familiarity: {}, avg_familiarity: 0 },
    },
    per_agenda: {
      IN6: { count: 2, maturity: { Contested: 2 }, familiarity: {}, avg_familiarity: 0 },
    },
  }

  it('uses an agenda’s own votes when it has them', () => {
    const r = agendaReading(summary, 'IN', 'IN6', 'typical')
    expect(r.tier).toBe('Contested')
    expect(r.inherited).toBe(false)
    expect(r.n).toBe(2)
  })

  it('inherits the area’s votes for an unrated agenda, and marks it', () => {
    const r = agendaReading(summary, 'IN', 'IN2', 'typical')
    // IN votes: E,E,U → weaker middle of the two-E-one-U is E.
    expect(r.tier).toBe('Early / partial')
    expect(r.inherited).toBe(true)
  })

  it('a community cell is empty when the area has no votes at all', () => {
    const cell = gridCell('community', 'typical', 'DS', 'Detecting deception', 'P4', summary)
    expect(cell.tier).toBeNull()
  })
})

describe('distribution statistics', () => {
  it('reports a median, a mode, and flags two peaks', () => {
    const s = tierStats(counts(['Robust (small scale)', 'Robust (small scale)', 'Contested', 'Contested']))
    expect(s.n).toBe(4)
    expect(s.modes).toContain('Robust (small scale)')
    expect(s.modes).toContain('Contested')
    expect(s.bimodal).toBe(true)
  })

  it('does not call a single spread bimodal', () => {
    const s = tierStats(counts(['Early / partial', 'Early / partial', 'Untested']))
    expect(s.bimodal).toBe(false)
  })

  it('gives familiarity a mean, because its steps are equal', () => {
    const s = numericStats({ '0': 1, '1': 0, '2': 0, '3': 1 }, [0, 1, 2, 3])
    expect(s.mean).toBe(1.5)
    expect(s.bimodal).toBe(true) // peaks at 0 and 3, a gap between
  })
})
