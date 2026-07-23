/**
 * The focus classifier is what turned the hand-written priority list into
 * something computed, so its shape is pinned here. The assertions are of two
 * kinds: structural invariants that must hold on any dataset, and a few facts
 * anchored to iteration-0 data (each flagged) that check the algorithm agrees
 * with the curated editorial read, the strongest evidence it is doing something
 * real, not arbitrary.
 */
import { describe, expect, it } from 'vitest'
import {
  FOCUS_CATEGORIES,
  focusAreas,
  focusProblems,
  MATURITY_VALUE,
  ROBUST_FLOOR,
  type FocusBoard,
  type FocusCategoryId,
} from '@/lib/focus'
import type { Summary } from '@/types/atlas'

const CATS: FocusCategoryId[] = ['easy', 'bottleneck', 'underexplored']

function allItems(board: FocusBoard) {
  return CATS.flatMap((c) => board[c])
}

describe('structural invariants (any dataset)', () => {
  it('places each unit in at most one category', () => {
    for (const board of [focusAreas('research', null), focusProblems('research', null)]) {
      const ids = allItems(board).map((i) => i.id)
      expect(new Set(ids).size).toBe(ids.length)
    }
  })

  it('never lists a unit that is already near-Robust (closest to solved)', () => {
    for (const board of [focusAreas('research', null), focusProblems('research', null)]) {
      for (const item of allItems(board)) {
        expect(item.signals.maturity).not.toBeNull()
        expect(item.signals.maturity as number).toBeLessThan(ROBUST_FLOOR)
      }
    }
  })

  it('tags every item with the category it was filed under, and a reason', () => {
    const board = focusAreas('research', null)
    for (const cat of CATS) {
      for (const item of board[cat]) {
        expect(item.category).toBe(cat)
        expect(item.reason.length).toBeGreaterThan(0)
      }
    }
  })

  it('names three categories and gives each a per-unit blurb', () => {
    expect(FOCUS_CATEGORIES.map((c) => c.id)).toEqual(CATS)
    for (const c of FOCUS_CATEGORIES) {
      expect(c.blurb('area')).not.toBe(c.blurb('problem'))
    }
  })

  it('treats Contested and Never demonstrated as non-progress', () => {
    expect(MATURITY_VALUE.Contested).toBe(0)
    expect(MATURITY_VALUE['Never demonstrated']).toBe(0)
  })
})

describe('agrees with the curated read (iteration-0 data)', () => {
  it('puts the measurement hub P1 among the bottleneck problems', () => {
    // The dataset's own note calls P1 "the hub of the whole map"; the most
    // agendas depend on it. It should surface as leverage, not as almost-done.
    const p = focusProblems('research', null)
    expect(p.bottleneck.map((i) => i.id)).toContain('P1')
  })

  it('puts thin, low-agenda areas under underexplored', () => {
    // Security Hardening is the thinnest column (fewest problems covered).
    const a = focusAreas('research', null)
    expect(a.underexplored.map((i) => i.id)).toContain('SH')
  })

  it('ranks within a category by how strongly it fits', () => {
    const p = focusProblems('research', null)
    for (const cat of CATS) {
      const scores = p[cat].map((i) => i.score)
      expect(scores).toEqual([...scores].sort((x, y) => y - x))
    }
  })
})

describe('follows the source', () => {
  const summary: Summary = {
    iteration: 0,
    totals: { total_submissions: 4, unique_submitters: 4 },
    per_area: {
      // IN voted strongly (should NOT read as underexplored); ST voted very weak.
      IN: { count: 3, maturity: { 'Strong existence proof': 2, 'Robust (small scale)': 1 }, familiarity: {}, avg_familiarity: 0 },
      ST: { count: 2, maturity: { Untested: 2 }, familiarity: {}, avg_familiarity: 0 },
    },
    per_agenda: {},
  }

  it('classifies only the areas the community has actually voted on', () => {
    const a = focusAreas('community', summary)
    const ids = allItems(a).map((i) => i.id)
    // Only IN and ST have votes; everything else has no community signal.
    expect(ids.every((id) => id === 'IN' || id === 'ST')).toBe(true)
    // IN's votes average Strong+ (>= ROBUST_FLOOR) → excluded as near-solved.
    expect(ids).not.toContain('IN')
    // ST voted Untested → a live focus, and weak → not "easy".
    expect(ids).toContain('ST')
    expect(a.easy.map((i) => i.id)).not.toContain('ST')
  })

  it('gives an empty board when the community has voted on nothing', () => {
    const board = focusAreas('community', { ...summary, per_area: {} })
    expect(allItems(board)).toHaveLength(0)
  })

  it('research and community can disagree about the same unit', () => {
    const research = focusAreas('research', null)
    const community = focusAreas('community', summary)
    // ST is not near-solved in research (it appears somewhere)…
    expect(allItems(research).map((i) => i.id)).toContain('ST')
    // …and appears in community too, but placed from the votes, not our ratings.
    expect(allItems(community).map((i) => i.id)).toContain('ST')
  })
})
