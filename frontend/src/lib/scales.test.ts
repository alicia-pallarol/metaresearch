import { describe, expect, it } from 'vitest'
import { describeCell, familiarityBucket, familiarityFill, tierAbbrev, tierStyle } from './scales'
import { tierOrder } from './atlas'

describe('tier styling', () => {
  it('gives every declared tier its own fill and abbreviation', () => {
    const fills = new Set<string>()
    const abbrevs = new Set<string>()
    for (const tier of tierOrder) {
      const style = tierStyle(tier)
      expect(style.kind).not.toBe('blank')
      fills.add(style.fill)
      abbrevs.add(tierAbbrev(tier))
    }
    expect(fills.size).toBe(tierOrder.length)
    expect(abbrevs.size).toBe(tierOrder.length)
  })

  it('keeps the two off-axis tiers off the evidence ramp', () => {
    // Contested and Never demonstrated are different kinds of claim, not weaker
    // rungs, so they carry the hatch that distinguishes them without colour.
    expect(tierStyle('Contested').hatched).toBe(true)
    expect(tierStyle('Never demonstrated').hatched).toBe(true)
    expect(tierStyle('Untested').hatched).toBe(false)
    expect(tierStyle('Robust (small scale)').hatched).toBe(false)
    expect(tierStyle('Contested').kind).toBe('disputed')
    expect(tierStyle('Never demonstrated').kind).toBe('negative')
  })

  it('treats blank as the absence of a tier, not as a tier', () => {
    const blank = tierStyle(null)
    expect(blank.kind).toBe('blank')
    expect(blank.fill).toBe('var(--cell-blank)')
    expect(tierAbbrev(null)).toBe('')
    expect(describeCell(null)).toMatch(/no meaningful relationship/i)
  })
})

describe('familiarity scale', () => {
  it('buckets averages to the five answer levels, and maps each integer to itself', () => {
    const cases: Array<[number, 0 | 1 | 2 | 3 | 4]> = [
      [0, 0],
      [0.49, 0],
      [0.5, 1],
      [1, 1],
      [1.5, 2],
      [2, 2],
      [2.5, 3],
      [3, 3],
      [3.5, 4],
      [4, 4],
    ]
    for (const [avg, want] of cases) {
      expect(familiarityBucket(avg)).toBe(want)
    }
  })

  it('uses a different colour family from the maturity ramp, across all five levels', () => {
    const maturityFills = tierOrder.map((t) => tierStyle(t).fill)
    for (const avg of [0, 1, 2, 3, 4]) {
      expect(maturityFills).not.toContain(familiarityFill(avg))
      expect(familiarityFill(avg)).toMatch(/^var\(--fam-[0-4]\)$/)
    }
  })
})
