/**
 * FOCUS AREAS, where the next unit of effort pays off most.
 * =========================================================================
 *
 * This module replaces the old hand-written "priority tiers". Nothing here is
 * hardcoded: it reads the grid and sorts the map into three focus categories,
 * for research areas and for problems independently. It follows the same
 * Research / Community source the grid is set to, so the two always agree.
 *
 * It is written to be read and tuned. Every threshold and weight is a named
 * constant in the TUNABLES block below, with a comment on what moving it does.
 * If you want to change what counts as a bottleneck, or how "already solved" is
 * decided, this is the one file to edit.
 *
 *
 * THE THREE CATEGORIES
 * --------------------
 *   Easy to intervene  Real traction already, mid-to-strong evidence and enough
 *                      active work, but not yet finished. A marginal push plausibly
 *                      moves it up. (High maturity, not Robust, some breadth.)
 *   Bottleneck         Many other parts of the map depend on it, so progress here
 *                      unblocks them. For a problem: how many research areas reach
 *                      it. For an area: how much its problems are shared with other
 *                      areas. (High connectivity, weighted toward the still-weak.)
 *   Underexplored      Thin: few agendas, many blank cells, low maturity. Explored-
 *                      and-contested does NOT count as underexplored. (Low coverage,
 *                      low maturity, low dispute.)
 *
 *
 * HOW A UNIT IS PLACED
 * --------------------
 *   1. Read the unit's SIGNALS (below). Structure, coverage, blanks, agenda
 *      counts, connectivity, comes from the curated grid and is the same in both
 *      sources. MATURITY is what the source changes: the average of our curated
 *      tiers, or the average of the community's votes.
 *   2. EXCLUDE it if it is already essentially solved (average maturity at or above
 *      ROBUST_FLOOR), because "closest to solved" is not a focus. In the community
 *      source, also exclude a unit with no votes at all, we have no signal to place
 *      it, and would rather say nothing than guess.
 *   3. Score it on all three lenses (each in 0..1) and put it in its STRONGEST one.
 *      If even that score is below MIN_SCORE the unit is unremarkable and is left
 *      out, "focus" means a short list, not all twelve.
 *   4. Within each category, rank by score, strongest first.
 *
 * A unit can genuinely fit two lenses (a neglected linchpin is both underexplored
 * and a bottleneck); it is shown under whichever scores highest, and its signals
 * are on the card so the other reading is still visible.
 */
import { atlas, getAreaByName, getAreaByTag, tierOrder } from '@/lib/atlas'
import type { GridSource } from '@/lib/community'
import type { Summary, Tier } from '@/types/atlas'

/* ======================================================================== *
 *  TUNABLES, change these to change what counts as a focus area.          *
 * ======================================================================== */

/**
 * The maturity ladder as numbers, so cells can be averaged. Only the four
 * on-axis tiers are progress. Contested and Never demonstrated are deliberately
 * 0: they are a live dispute and a negative finding, not steps toward "solved",
 * and a cell that reached them should not lift a unit's maturity. They are
 * tracked separately as `disputed` (see below) so an explored-but-contested unit
 * is not mistaken for an unexplored one.
 */
export const MATURITY_VALUE: Record<Tier, number> = {
  'Robust (small scale)': 4,
  'Strong existence proof': 3,
  'Early / partial': 2,
  Untested: 1,
  Contested: 0,
  'Never demonstrated': 0,
}

/** Average maturity at or above this = "closest to solved", excluded entirely. */
export const ROBUST_FLOOR = 3.3

/** Below this maturity, work is too green to call "easy to intervene". */
export const EASY_MIN_MATURITY = 1.5

/**
 * "Easy to intervene" blends the BEST result already on the board with the
 * AVERAGE, at this weight on the best. Pure-best would let one fluke Robust cell
 * (e.g. the systems-security sandboxing outlier) crown a mostly-weak unit; the
 * average tempers it. 1 = best only, 0 = average only.
 */
export const EASY_BEST_WEIGHT = 0.6

/**
 * A unit whose strongest lens scores below this (after the normalisation
 * described under "HOW A UNIT IS PLACED", step 3) is too unremarkable to be a
 * focus, and is left out. Raise it to shorten every list.
 */
export const MIN_SCORE = 0.15

/** At most this many units per category, strongest first. Keeps it a shortlist. */
export const MAX_PER_CATEGORY = 6

/** Weights inside each lens, before the cross-unit normalisation. */
const WEIGHTS = {
  /** Easy: how much breadth of coverage adds on top of raw maturity. */
  easyCoverage: 0.5,
  /** Bottleneck: how much being still-weak adds on top of raw connectivity. */
  bottleneckWeakness: 0.5,
  /** Underexplored: split between "few problems/areas touched" and "few agendas". */
  underBlankness: 0.55,
  underThinness: 0.45,
  /** Underexplored: how much an explored-but-contested unit is discounted. */
  underDisputePenalty: 0.6,
}

/* ======================================================================== *
 *  Categories                                                              *
 * ======================================================================== */

export type FocusCategoryId = 'easy' | 'bottleneck' | 'underexplored'
export type FocusUnit = 'area' | 'problem'

export interface FocusCategory {
  id: FocusCategoryId
  label: string
  /** Rendered under the category heading. */
  blurb: (unit: FocusUnit) => string
}

export const FOCUS_CATEGORIES: FocusCategory[] = [
  {
    id: 'easy',
    label: 'Easy to intervene',
    blurb: (u) =>
      u === 'area'
        ? 'Areas with real traction, mid-to-strong evidence and active work, that are not yet finished. A marginal push should move them.'
        : 'Problems already partly solved across several areas. Close enough that a push could get an existence proof over the line.',
  },
  {
    id: 'bottleneck',
    label: 'Bottlenecks',
    blurb: (u) =>
      u === 'area'
        ? 'Areas whose problems many OTHER areas also depend on. Strengthening them unblocks work elsewhere on the map.'
        : 'Problems many research areas depend on. Solving one unblocks every area that is currently betting on it.',
  },
  {
    id: 'underexplored',
    label: 'Underexplored',
    blurb: (u) =>
      u === 'area'
        ? 'Thin areas: few agendas, many blank cells, low maturity. Room that has barely been entered, not room that was tried and contested.'
        : 'Problems almost nobody works on: reached by few areas, mostly blank, largely untested.',
  },
]

/* ======================================================================== *
 *  Output                                                                  *
 * ======================================================================== */

export interface FocusSignals {
  /** Fraction of this unit's cells that are not blank, 0..1. */
  coverage: number
  /** How many of this unit's cells are blank. */
  blanks: number
  /** Total agendas behind the unit (distinct, each agenda sits in one area). */
  agendas: number
  /** Average on-axis maturity 0..4 for the current source, or null with no signal. */
  maturity: number | null
  /**
   * Best on-axis maturity any cell reaches, 0..4. This is closeness-to-a-result,
   * and it drives "easy to intervene": for a problem, coverage and connectivity
   * are the same count of areas, so the average would make "easy" and
   * "bottleneck" the same signal, the best tier is what actually separates
   * "there is already a strong result here" from "many areas depend on it".
   * Only from the curated grid (research); the community does not vote per cell.
   */
  bestMaturity: number
  /** Raw connectivity: for a problem, areas reaching it; for an area, the average
   *  number of OTHER areas its problems are shared with. */
  connectivity: number
  /** Fraction of non-blank cells that are Contested / Never demonstrated, 0..1. */
  disputed: number
}

export interface FocusItem {
  /** Area tag ("IN") or problem id ("P1"). */
  id: string
  name: string
  unit: FocusUnit
  category: FocusCategoryId
  /** The winning lens score, for ranking. */
  score: number
  signals: FocusSignals
  /** A factual one-liner built from the signals, no curated prose, no fabrication. */
  reason: string
}

export type FocusBoard = Record<FocusCategoryId, FocusItem[]>

/* ======================================================================== *
 *  Structure (source-independent): coverage, blanks, agendas, connectivity *
 * ======================================================================== */

const areaOrder = atlas.area_matrix.area_order
const problems = atlas.problems
const AREA_N = areaOrder.length
const PROBLEM_N = problems.length

function cell(problemId: string, areaName: string) {
  return atlas.area_matrix.cells[problemId]?.[areaName] ?? null
}

/** How many research areas have a non-blank cell against a problem (its leverage). */
function reachOf(problemId: string): number {
  return areaOrder.reduce((n, area) => n + (cell(problemId, area) ? 1 : 0), 0)
}

const problemReach = new Map<string, number>(problems.map((p) => [p.id, reachOf(p.id)]))

/* ======================================================================== *
 *  Maturity (source-dependent): our curated tiers, or the community's votes *
 * ======================================================================== */

function meanMaturity(values: number[]): number | null {
  if (values.length === 0) return null
  return values.reduce((s, v) => s + v, 0) / values.length
}

/** The community's average vote for one area, or null if nobody has voted on it. */
function communityAreaMaturity(summary: Summary | null, areaTag: string): number | null {
  const hist = summary?.per_area?.[areaTag]?.maturity
  if (!hist) return null
  let sum = 0
  let n = 0
  for (const tier of tierOrder) {
    const c = hist[tier] ?? 0
    sum += MATURITY_VALUE[tier] * c
    n += c
  }
  return n > 0 ? sum / n : null
}

/* ======================================================================== *
 *  Per-unit signal builders                                                *
 * ======================================================================== */

function areaSignals(areaTag: string, source: GridSource, summary: Summary | null): FocusSignals {
  const area = getAreaByTag(areaTag)
  const areaName = area?.name ?? ''

  const tiers: Tier[] = []
  let disputedCount = 0
  const otherAreasShared: number[] = []

  for (const problem of problems) {
    const c = cell(problem.id, areaName)
    if (!c) continue
    tiers.push(c.tier)
    if (MATURITY_VALUE[c.tier] === 0) disputedCount++
    // Every other area that also reaches this problem, the shared-ness that
    // makes strengthening this area spill over.
    otherAreasShared.push((problemReach.get(problem.id) ?? 1) - 1)
  }

  const nonBlank = tiers.length
  const mvs = tiers.map((t) => MATURITY_VALUE[t])
  const maturity = source === 'research' ? meanMaturity(mvs) : communityAreaMaturity(summary, areaTag)

  return {
    coverage: nonBlank / PROBLEM_N,
    blanks: PROBLEM_N - nonBlank,
    agendas: area?.agenda_ids.length ?? 0,
    maturity,
    // Research knows the best cell; the community only votes an area average, so
    // there is no finer "best" than that average.
    bestMaturity: source === 'research' ? (mvs.length ? Math.max(...mvs) : 0) : maturity ?? 0,
    connectivity: meanMaturity(otherAreasShared) ?? 0,
    disputed: nonBlank > 0 ? disputedCount / nonBlank : 0,
  }
}

function problemSignals(problemId: string, source: GridSource, summary: Summary | null): FocusSignals {
  const tiers: Tier[] = []
  let agendas = 0
  let disputedCount = 0
  const communityPerArea: number[] = []

  for (const areaName of areaOrder) {
    const c = cell(problemId, areaName)
    if (!c) continue
    tiers.push(c.tier)
    agendas += c.count
    if (MATURITY_VALUE[c.tier] === 0) disputedCount++
    // areaOrder holds names; the votes are keyed by tag.
    const areaTag = getAreaByName(areaName)?.tag
    const m = areaTag ? communityAreaMaturity(summary, areaTag) : null
    if (m !== null) communityPerArea.push(m)
  }

  const nonBlank = tiers.length
  const mvs = tiers.map((t) => MATURITY_VALUE[t])
  // A problem is not voted on directly; its community maturity is the average of
  // the votes of the areas that reach it.
  const maturity = source === 'research' ? meanMaturity(mvs) : meanMaturity(communityPerArea)

  return {
    coverage: nonBlank / AREA_N,
    blanks: AREA_N - nonBlank,
    agendas,
    maturity,
    bestMaturity: source === 'research' ? (mvs.length ? Math.max(...mvs) : 0) : maturity ?? 0,
    // A problem's leverage is how much work rides on it. Agenda count captures
    // that better than area count: P1 has 16 agendas betting on it across a
    // handful of areas, deep dependence a plain area tally would miss. Swap this
    // for `nonBlank` to weigh by breadth of areas instead.
    connectivity: agendas,
    disputed: nonBlank > 0 ? disputedCount / nonBlank : 0,
  }
}

/* ======================================================================== *
 *  Scoring                                                                 *
 *                                                                          *
 *  Each unit gets a raw AFFINITY for each lens. The three lenses live on   *
 *  different natural scales, maturity is 0..4, connectivity is a count,  *
 *  so a raw argmax would always pick the same one. Each affinity is placed  *
 *  on a 0..1 ruler by min-max against a fixed REFERENCE, the spread of     *
 *  that lens across the whole map in our own ratings, so the three lenses  *
 *  get comparable ranges and a unit lands in the lens it is most extreme    *
 *  on. (Dividing by the max instead would let the flattest lens win for     *
 *  everyone; stretching each lens to its own range is what separates them.) *
 *                                                                           *
 *  The reference is the full 12-unit research population, not the handful   *
 *  that survive the filters, so the ruler does not move when only a few     *
 *  units have votes, a single voted area is still measured against the     *
 *  same scale as everything else, and the two sources are directly          *
 *  comparable.                                                              *
 * ======================================================================== */

const clamp01 = (x: number) => Math.max(0, Math.min(1, x))

/** How weak a unit still is, 0 (solved) .. 1 (untested). Used by two lenses. */
function weakness(maturity: number): number {
  return clamp01((ROBUST_FLOOR - maturity) / ROBUST_FLOOR)
}

interface Affinity {
  easy: number
  bottleneck: number
  underexplored: number
}

/** Raw, un-normalised affinities. Maturity is real here (guarded by the caller). */
function affinities(s: FocusSignals, agendasNorm: number): Affinity {
  const mat = s.maturity as number
  const close = EASY_BEST_WEIGHT * s.bestMaturity + (1 - EASY_BEST_WEIGHT) * mat
  return {
    // Closer to done: the best result on the board, tempered by the average so a
    // lone fluke cell can't crown a weak unit; lightly favouring broader units.
    easy: Math.max(0, close - EASY_MIN_MATURITY) * (1 - WEIGHTS.easyCoverage + WEIGHTS.easyCoverage * s.coverage),
    // Central, weighted toward the parts that are still weak.
    bottleneck: s.connectivity * (1 - WEIGHTS.bottleneckWeakness + WEIGHTS.bottleneckWeakness * weakness(mat)),
    // Thin and untested, discounted where the thinness is really live dispute.
    underexplored:
      (WEIGHTS.underBlankness * (1 - s.coverage) + WEIGHTS.underThinness * (1 - agendasNorm)) *
      weakness(mat) *
      (1 - WEIGHTS.underDisputePenalty * s.disputed),
  }
}

/**
 * The min and max of each lens across our own ratings of all units, plus the
 * agenda count to normalise thinness against. This is the fixed ruler the scoring
 * uses. Computed lazily and cached: it depends only on the static dataset.
 */
interface Reference {
  min: Affinity
  max: Affinity
  maxAgendas: number
}
const referenceCache = new Map<FocusUnit, Reference>()

function reference(unit: FocusUnit): Reference {
  const cached = referenceCache.get(unit)
  if (cached) return cached

  const signals = unitIds(unit).map((id) => signalsOf(unit, id, 'research', null))
  const maxAgendas = Math.max(1, ...signals.map((s) => s.agendas))
  const affs = signals.filter((s) => s.maturity !== null).map((s) => affinities(s, s.agendas / maxAgendas))

  const lo = (pick: (a: Affinity) => number) => Math.min(...affs.map(pick))
  const hi = (pick: (a: Affinity) => number) => Math.max(...affs.map(pick))
  const value: Reference = {
    min: { easy: lo((a) => a.easy), bottleneck: lo((a) => a.bottleneck), underexplored: lo((a) => a.underexplored) },
    max: { easy: hi((a) => a.easy), bottleneck: hi((a) => a.bottleneck), underexplored: hi((a) => a.underexplored) },
    maxAgendas,
  }
  referenceCache.set(unit, value)
  return value
}

/** Position of v within a lens's reference range, clamped to 0..1. */
function onRuler(v: number, min: number, max: number): number {
  return max > min ? clamp01((v - min) / (max - min)) : 0
}

/* ======================================================================== *
 *  Reasons, factual, generated from the signals                          *
 * ======================================================================== */

/** The nearest ladder tier to an average, for display. Off-axis tiers excluded. */
function nearestTier(maturity: number): Tier {
  const ladder: Array<[number, Tier]> = [
    [4, 'Robust (small scale)'],
    [3, 'Strong existence proof'],
    [2, 'Early / partial'],
    [1, 'Untested'],
  ]
  return ladder.reduce((best, cur) =>
    Math.abs(cur[0] - maturity) < Math.abs(best[0] - maturity) ? cur : best,
  )[1]
}

function reasonFor(item: Omit<FocusItem, 'reason'>): string {
  const s = item.signals
  const tierWord = s.maturity === null ? 'unrated' : nearestTier(s.maturity).toLowerCase()
  const cellWord = item.unit === 'area' ? 'problems' : 'areas'

  switch (item.category) {
    case 'easy':
      return `Typical maturity ${tierWord} across ${PROBLEM_N - s.blanks || s.agendas} ${cellWord}, ${s.agendas} agendas, traction to build on, not yet robust.`
    case 'bottleneck':
      return item.unit === 'problem'
        ? `${s.agendas} agendas across ${AREA_N - s.blanks} areas depend on it; still ${tierWord}. Solving it unblocks them.`
        : `Its problems are shared with ${s.connectivity.toFixed(1)} other areas on average; still ${tierWord}. Progress here spills over.`
    case 'underexplored':
      return `${s.blanks} of ${item.unit === 'area' ? PROBLEM_N : AREA_N} ${cellWord} blank, ${s.agendas} agendas, typical maturity ${tierWord}, thin and largely untested.`
  }
}

/* ======================================================================== *
 *  Public entry point                                                      *
 * ======================================================================== */

/** The ids of every unit of a type, in dataset order. */
function unitIds(unit: FocusUnit): string[] {
  return unit === 'area'
    ? areaOrder.map((name) => getAreaByName(name)?.tag ?? name)
    : problems.map((p) => p.id)
}

function signalsOf(unit: FocusUnit, id: string, source: GridSource, summary: Summary | null): FocusSignals {
  return unit === 'area' ? areaSignals(id, source, summary) : problemSignals(id, source, summary)
}

function nameOf(unit: FocusUnit, id: string): string {
  return unit === 'area' ? getAreaByTag(id)?.name ?? id : problems.find((p) => p.id === id)?.name ?? id
}

function build(unit: FocusUnit, source: GridSource, summary: Summary | null): FocusBoard {
  const { min, max, maxAgendas } = reference(unit)
  const board: FocusBoard = { easy: [], bottleneck: [], underexplored: [] }

  // Every unit that has a signal (in the community source, that someone has voted
  // on) and is not already essentially solved.
  const rows = unitIds(unit)
    .map((id) => ({ id, name: nameOf(unit, id), signals: signalsOf(unit, id, source, summary) }))
    .filter((r) => r.signals.maturity !== null && (r.signals.maturity as number) < ROBUST_FLOOR)

  for (const r of rows) {
    const raw = affinities(r.signals, r.signals.agendas / maxAgendas)
    // Onto the shared ruler, then file under the strongest lens.
    const lenses: Array<[FocusCategoryId, number]> = [
      ['easy', onRuler(raw.easy, min.easy, max.easy)],
      ['bottleneck', onRuler(raw.bottleneck, min.bottleneck, max.bottleneck)],
      ['underexplored', onRuler(raw.underexplored, min.underexplored, max.underexplored)],
    ]
    lenses.sort((a, b) => b[1] - a[1])
    const [category, score] = lenses[0]!
    if (score < MIN_SCORE) continue
    const base = { id: r.id, name: r.name, unit, category, score, signals: r.signals }
    board[category].push({ ...base, reason: reasonFor(base) })
  }

  for (const id of Object.keys(board) as FocusCategoryId[]) {
    board[id] = board[id].sort((a, b) => b.score - a.score).slice(0, MAX_PER_CATEGORY)
  }
  return board
}

export function focusAreas(source: GridSource, summary: Summary | null): FocusBoard {
  return build('area', source, summary)
}

export function focusProblems(source: GridSource, summary: Summary | null): FocusBoard {
  return build('problem', source, summary)
}
