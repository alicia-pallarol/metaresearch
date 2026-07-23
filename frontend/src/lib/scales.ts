/**
 * Colour scales for the two heatmaps.
 *
 * Two scales, deliberately kept apart and separately labelled:
 *
 *   1. MATURITY, how mature the evidence is that an agenda addresses a problem.
 *      Four of the six tiers form an ordered evidence axis (Untested → Early →
 *      Strong existence proof → Robust) and take a single-hue blue ramp,
 *      light → dark. Two tiers are NOT points on that axis and must not be
 *      coloured as if they were:
 *        · "Never demonstrated" is a negative finding (looked for, not found)
 *        · "Contested" is a live dispute, a stronger claim than Untested
 *      Both take an off-axis hue plus a diagonal hatch, so they read as a
 *      different kind of statement rather than a rung on the ladder.
 *      A blank cell is not a tier at all and gets no fill.
 *
 *   2. FAMILIARITY, the community's average self-rated familiarity, 0–3.
 *      A separate single-hue green ramp so it can never be confused with the
 *      maturity scale when the overlay is toggled on.
 *
 * Every step below was checked with the data-viz palette validator: both ramps
 * pass the ordinal checks (monotone lightness, adjacent ΔL ≥ 0.06, light end
 * ≥ 2:1 against the surface) in light and dark mode, and the two off-axis hues
 * clear the CVD separation floor against the ramp and against each other. The
 * hatch and the always-present cell labels are the required secondary encoding.
 */
import type { CellTier, Tier } from '@/types/atlas'

export interface TierStyle {
  /** CSS custom property name holding the fill for this tier. */
  fill: string
  /** Ink that stays readable on that fill. */
  ink: string
  /** Whether the cell carries the off-axis hatch. */
  hatched: boolean
  /** One-word shape of the claim, used in the legend and in aria labels. */
  kind: 'evidence' | 'negative' | 'disputed' | 'blank'
}

const STYLES: Record<Tier, TierStyle> = {
  'Robust (small scale)': { fill: 'var(--tier-robust)', ink: 'var(--ink-on-strong)', hatched: false, kind: 'evidence' },
  'Strong existence proof': { fill: 'var(--tier-strong)', ink: 'var(--ink-on-strong)', hatched: false, kind: 'evidence' },
  'Early / partial': { fill: 'var(--tier-early)', ink: 'var(--ink-on-mid)', hatched: false, kind: 'evidence' },
  Untested: { fill: 'var(--tier-untested)', ink: 'var(--ink-on-light)', hatched: false, kind: 'evidence' },
  'Never demonstrated': { fill: 'var(--tier-never)', ink: 'var(--ink-on-strong)', hatched: true, kind: 'negative' },
  Contested: { fill: 'var(--tier-contested)', ink: 'var(--ink-on-contested)', hatched: true, kind: 'disputed' },
}

const BLANK: TierStyle = { fill: 'var(--cell-blank)', ink: 'var(--ink-muted)', hatched: false, kind: 'blank' }

export function tierStyle(tier: CellTier): TierStyle {
  return tier === null ? BLANK : (STYLES[tier] ?? BLANK)
}

/**
 * One-letter tier code, the text fallback so tier identity never rests on colour
 * alone. Deliberately a SINGLE letter: two-letter codes collided with the
 * two-letter research-area tags (the old "EA" for Early/partial read as the
 * Empirical Alignment area). One letter can't be mistaken for an area or agenda
 * code. R > S > E > U in strength; N and C are the two off-axis tiers.
 */
export function tierAbbrev(tier: CellTier): string {
  switch (tier) {
    case 'Robust (small scale)':
      return 'R'
    case 'Strong existence proof':
      return 'S'
    case 'Early / partial':
      return 'E'
    case 'Untested':
      return 'U'
    case 'Never demonstrated':
      return 'N'
    case 'Contested':
      return 'C'
    default:
      return ''
  }
}

/**
 * Familiarity colour helpers.
 *
 * The green `--fam-*` ramp is separate from the maturity ramp so the two scales
 * can never be confused. It backs the familiarity histogram in the feedback
 * panel; the exact average is always printed next to the colour.
 *
 * Buckets rather than a continuous interpolation: the underlying answers are five
 * discrete options (0 = never heard of it .. 4 = expert).
 */
export function familiarityBucket(avg: number): 0 | 1 | 2 | 3 | 4 {
  if (avg < 0.5) return 0
  if (avg < 1.5) return 1
  if (avg < 2.5) return 2
  if (avg < 3.5) return 3
  return 4
}

export function familiarityFill(avg: number): string {
  return `var(--fam-${familiarityBucket(avg)})`
}

/**
 * Ink that stays readable on the fill. The `--fam-*` ramp keeps its lightness
 * order in both themes (0 is the palest, 4 the deepest), unlike the maturity ink
 * tokens which flip with the theme, so the ink is chosen by the fill's own
 * lightness with fixed colours, not by a theme-flipping variable.
 */
export function familiarityInk(avg: number): string {
  return familiarityBucket(avg) <= 1 ? '#0b0b0b' : '#ffffff'
}

/** Human sentence for a cell, used by tooltips and screen readers. */
export function describeCell(tier: CellTier, agendaCount?: number): string {
  if (tier === null) return 'Blank: no meaningful relationship between this area and this problem.'
  const suffix =
    agendaCount === undefined
      ? ''
      : ` Best of ${agendaCount} agenda${agendaCount === 1 ? '' : 's'} in this area that reach the problem.`
  return `${tier}.${suffix}`
}
