/**
 * The two axes the grid can be read along, and the arithmetic behind them.
 *
 * SOURCE, whose judgment is being shown.
 *   "research"  the curated ratings in data/atlas.json: what the authors read and
 *               concluded, one source at a time.
 *   "community" the maturity votes researchers have submitted through the
 *               feedback form.
 *
 *   The two share a skeleton. Which agendas address which problem is a curated
 *   claim and is NOT voted on, the community substitutes its own maturity
 *   ratings into that same structure. So a cell is the same set of agendas in
 *   both views; only the colour changes hands.
 *
 * SCENARIO, which end of the spread is being shown.
 *   Every cell stands on a set of ratings: the tiers of the agendas behind it,
 *   and (in the community view) the tiers those agendas were voted. "Best case"
 *   summarises the strongest quarter of that set, "worst case" the weakest, and
 *   "typical" the whole of it.
 *
 * Two rules keep this honest on data that is ordinal, not numeric:
 *
 *   1. A band is summarised by its MEDIAN, never by a mean. The six tiers are not
 *      numbers, and two of them (Contested, Never demonstrated) are not even
 *      rungs on the evidence ladder, they are a live dispute and a negative
 *      finding. Averaging them would invent a value nobody claimed. A median is
 *      always a tier some rating actually is.
 *   2. Where a band has no single middle value, the WEAKER of the two is shown.
 *      The map does not round up.
 *
 * The band is a quarter of the ratings, at least one. Below eight ratings that
 * quarter is a single rating, so "best case" is simply the strongest and "worst
 * case" the weakest, which is why the research view in best case is exactly the
 * "best of N agendas" grid this map has always shown.
 */
import { agendasReaching, tierOrder } from '@/lib/atlas'
import type { Summary, Tier } from '@/types/atlas'

export type GridSource = 'research' | 'community'
export type Scenario = 'best' | 'typical' | 'worst'

/** A histogram over the six tiers. Missing keys are zero. */
export type TierCounts = Partial<Record<Tier, number>>

/** One cell of the grid, whichever source and scenario produced it. */
export interface GridCell {
  /** The tier to colour the cell, or null when there is nothing to show. */
  tier: Tier | null
  /** Agendas in this area that reach this problem, the same in both sources. */
  agendaCount: number
  /**
   * How many community ratings stand behind the cell (0 in the research view).
   * Counts submissions, not agendas: one reader who rated three agendas of the
   * area is one rating here.
   */
  ratings: number
  /**
   * True when the community view is showing the area's own votes because no
   * agenda in the cell has been rated individually yet.
   */
  inherited: boolean
}

export function tierCountsFrom(tiers: Tier[]): TierCounts {
  const counts: TierCounts = {}
  for (const tier of tiers) counts[tier] = (counts[tier] ?? 0) + 1
  return counts
}

export function totalOf(counts: TierCounts): number {
  let total = 0
  for (const tier of tierOrder) total += counts[tier] ?? 0
  return total
}

/**
 * The band summary: the median of the strongest quarter, the weakest quarter, or
 * the whole set. See the two rules at the top of this file.
 *
 * Works straight off the histogram, a histogram of ordinal values gives exact
 * quantiles, so nothing has to be expanded into a list first.
 */
export function bandTier(counts: TierCounts, scenario: Scenario): Tier | null {
  const n = totalOf(counts)
  if (n === 0) return null

  // Positions are indices into the ratings sorted strongest first.
  const quarter = Math.max(1, Math.floor(n / 4))
  const [start, end] =
    scenario === 'best' ? [0, quarter - 1] : scenario === 'worst' ? [n - quarter, n - 1] : [0, n - 1]

  // Rule 2: for an even band, floor(size / 2) lands on the weaker of the two
  // middles. For an odd band it is the middle itself.
  const target = start + Math.floor((end - start + 1) / 2)

  let seen = 0
  for (const tier of tierOrder) {
    seen += counts[tier] ?? 0
    if (seen > target) return tier
  }
  return null
}

/* --- Reading the two sources --------------------------------------------- */

/** The curated tiers of every agenda in this area that reaches this problem. */
export function researchTiers(areaName: string, problemId: string): Tier[] {
  return agendasReaching(areaName, problemId).map((row) => row.tier)
}

function knownTiers(raw: Record<string, number> | undefined): TierCounts | null {
  if (!raw) return null
  const counts: TierCounts = {}
  let total = 0
  for (const tier of tierOrder) {
    const n = raw[tier] ?? 0
    if (n > 0) {
      counts[tier] = n
      total += n
    }
  }
  return total > 0 ? counts : null
}

/** What the community voted for a research area as a whole. */
export function areaVotes(summary: Summary | null, areaTag: string): TierCounts | null {
  return knownTiers(summary?.per_area?.[areaTag]?.maturity)
}

/** What the community voted for one agenda specifically. */
export function agendaVotes(summary: Summary | null, agendaId: string): TierCounts | null {
  return knownTiers(summary?.per_agenda?.[agendaId]?.maturity)
}

export interface AgendaReading {
  tier: Tier | null
  /** Ratings behind it. */
  n: number
  /** True when this is the area's votes standing in for an unrated agenda. */
  inherited: boolean
}

/**
 * What the community makes of one agenda.
 *
 * An agenda nobody has rated yet inherits its area's votes: a reader who rated
 * the area did make a claim about the work inside it, they just did not break it
 * down. That is marked, so an inherited reading is never mistaken for one the
 * community actually gave agenda by agenda.
 */
export function agendaReading(
  summary: Summary | null,
  areaTag: string,
  agendaId: string,
  scenario: Scenario,
): AgendaReading {
  const own = agendaVotes(summary, agendaId)
  if (own) return { tier: bandTier(own, scenario), n: totalOf(own), inherited: false }

  const area = areaVotes(summary, areaTag)
  if (area) return { tier: bandTier(area, scenario), n: totalOf(area), inherited: true }

  return { tier: null, n: 0, inherited: false }
}

/**
 * The community reading of one cell: two levels, both summarised by the same band
 * rule, so the scenario means one thing throughout. Best case is the optimistic
 * quarter of readers on the strongest quarter of agendas; worst case the
 * pessimistic quarter on the weakest.
 */
function communityTier(
  summary: Summary | null,
  areaTag: string,
  areaName: string,
  problemId: string,
  scenario: Scenario,
): { tier: Tier | null; inherited: boolean } {
  const perAgenda: Tier[] = []
  let inherited = true

  for (const row of agendasReaching(areaName, problemId)) {
    const reading = agendaReading(summary, areaTag, row.agenda.id, scenario)
    if (!reading.inherited && reading.tier) inherited = false
    if (reading.tier) perAgenda.push(reading.tier)
  }
  if (perAgenda.length === 0) return { tier: null, inherited: false }

  return { tier: bandTier(tierCountsFrom(perAgenda), scenario), inherited }
}

/** One cell of the grid, for whichever source and scenario is switched on. */
export function gridCell(
  source: GridSource,
  scenario: Scenario,
  areaTag: string,
  areaName: string,
  problemId: string,
  summary: Summary | null,
): GridCell {
  const tiers = researchTiers(areaName, problemId)
  const base = { agendaCount: tiers.length, ratings: 0, inherited: false }

  if (source === 'research') {
    return { ...base, tier: bandTier(tierCountsFrom(tiers), scenario) }
  }
  return {
    ...base,
    ratings: summary?.per_area?.[areaTag]?.count ?? 0,
    ...communityTier(summary, areaTag, areaName, problemId, scenario),
  }
}

/* --- Distributions, for the histograms in the feedback panel -------------- */

export interface TierStats {
  n: number
  /** The middle rating, weaker of the two middles where there is no single one. */
  median: Tier | null
  /** Every tier tied at the highest count. */
  modes: Tier[]
  /** Two or more separated peaks: the room does not agree. */
  bimodal: boolean
}

export function tierStats(counts: TierCounts): TierStats {
  const n = totalOf(counts)
  if (n === 0) return { n: 0, median: null, modes: [], bimodal: false }

  const values = tierOrder.map((tier) => counts[tier] ?? 0)
  const highest = Math.max(...values)
  const modes = tierOrder.filter((_, i) => values[i] === highest)

  return {
    n,
    median: bandTier(counts, 'typical'),
    modes,
    bimodal: peakCount(values) >= 2,
  }
}

export interface NumericStats {
  n: number
  mean: number
  median: number
  modes: number[]
  bimodal: boolean
}

/** Familiarity is a 0–3 scale with equal steps, so it does have a mean. */
export function numericStats(counts: Record<string, number> | undefined, levels: number[]): NumericStats {
  const values = levels.map((level) => counts?.[String(level)] ?? 0)
  const n = values.reduce((sum, v) => sum + v, 0)
  if (n === 0) return { n: 0, mean: 0, median: 0, modes: [], bimodal: false }

  const total = levels.reduce((sum, level, i) => sum + level * (values[i] ?? 0), 0)
  const highest = Math.max(...values)

  return {
    n,
    mean: Math.round((total / n) * 10) / 10,
    // For an even count the two middle answers are averaged, which can land
    // between two levels, legitimate here, where the steps are equal.
    median: (valueAt(levels, values, Math.floor((n - 1) / 2)) + valueAt(levels, values, Math.floor(n / 2))) / 2,
    modes: levels.filter((_, i) => values[i] === highest),
    bimodal: peakCount(values) >= 2,
  }
}

function valueAt(levels: number[], counts: number[], index: number): number {
  let seen = 0
  for (let i = 0; i < levels.length; i++) {
    seen += counts[i] ?? 0
    if (seen > index) return levels[i] ?? 0
  }
  return levels[levels.length - 1] ?? 0
}

/**
 * How many separated peaks the histogram has, reading it in display order. A
 * plateau counts once; a bar with a lower bar on both sides is a peak. Two peaks
 * is the shape that matters: it says the responses are split rather than spread.
 */
function peakCount(values: number[]): number {
  let peaks = 0
  let i = 0
  while (i < values.length) {
    const height = values[i] ?? 0
    if (height === 0) {
      i++
      continue
    }
    let end = i
    while (end + 1 < values.length && values[end + 1] === height) end++

    const before = i === 0 ? -1 : (values[i - 1] ?? -1)
    const after = end === values.length - 1 ? -1 : (values[end + 1] ?? -1)
    if (height > before && height > after) peaks++

    i = end + 1
  }
  return peaks
}

/* --- Labels --------------------------------------------------------------- */

export const SCENARIO_LABELS: Record<Scenario, string> = {
  best: 'Best case',
  typical: 'Typical',
  worst: 'Worst case',
}

export const SCENARIO_HINTS: Record<Scenario, string> = {
  best: 'the strongest quarter of the ratings behind each cell',
  typical: 'the middle of all the ratings behind each cell',
  worst: 'the weakest quarter of the ratings behind each cell',
}

export const SOURCE_LABELS: Record<GridSource, string> = {
  research: 'Our research',
  community: 'The community',
}

/** How many ratings a cell needs before its band is more than one rating. */
export const BAND_THRESHOLD = 8
