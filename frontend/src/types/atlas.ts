/**
 * Types for data/atlas.json, the curated, manually verified dataset that is the
 * single source of truth for everything the site displays.
 *
 * These mirror the JSON exactly. If the shape here and the file disagree, the
 * file wins and this is the bug.
 */

/** The six maturity tiers, in the order given by legend.tier_order. */
export type Tier =
  | 'Robust (small scale)'
  | 'Strong existence proof'
  | 'Early / partial'
  | 'Untested'
  | 'Never demonstrated'
  | 'Contested'

/** A matrix cell: a tier, or null meaning "no meaningful relationship". */
export type CellTier = Tier | null

export interface AtlasMeta {
  iteration: number
  iteration_label: string
  title: string
  collaborators: string[]
  source_check_date: string
}

export interface AtlasLegend {
  tiers: Record<string, string>
  tier_order: Tier[]
  familiarity_scale: Record<string, string>
  notes: Record<string, string>
}

export interface Problem {
  id: string
  name: string
  short_def: string
}

export interface Area {
  tag: string
  name: string
  definition: string
  agenda_ids: string[]
}

export interface Agenda {
  id: string
  area_tag: string
  area_name: string
  agenda: string
  mechanism: string
  assumption: string
  overall_maturity: Tier
  best_evidence_against: string
  key_work: string
  source_check: string
}

/** agenda id -> problem id -> tier (or null). */
export type Matrix = Record<string, Record<string, CellTier>>

export interface AreaCell {
  /** The best tier any agenda in this area reaches against this problem. */
  tier: Tier
  /** How many agendas in this area reach the problem at all. */
  count: number
}

export interface AreaMatrix {
  area_order: string[]
  /** problem id -> area name -> cell (or null when no agenda reaches it). */
  cells: Record<string, Record<string, AreaCell | null>>
}

export interface PriorityItem {
  tier: string
  tier_label: string
  problem_id: string
  problem_name: string
  why_unsolved: string
  evidence_today: string
  most_connected: string
  best_tier_reached: string
}

export interface Priorities {
  intro: string
  tier_descriptions: Record<string, string>
  items: PriorityItem[]
  reading_note: string
}

export interface Atlas {
  meta: AtlasMeta
  legend: AtlasLegend
  problems: Problem[]
  areas: Area[]
  agendas: Agenda[]
  matrix: Matrix
  area_matrix: AreaMatrix
  priorities: Priorities
}

/* --- API types (backend/internal/store) ----------------------------------- */

/**
 * The public aggregate for one research area or one agenda: two histograms and
 * nothing that could identify anyone.
 *
 * `per_area` counts one vote per submission. `per_agenda` counts only the
 * submissions that rated that agenda specifically, so the two are different
 * populations, an agenda's count is never inflated by readers who only rated
 * its area.
 */
export interface RatingSummary {
  count: number
  /** Tier name -> how many readers voted it. */
  maturity: Record<string, number>
  /** "0".."3" -> how many readers reported that familiarity. */
  familiarity: Record<string, number>
  avg_familiarity: number
}

export interface Summary {
  iteration: number
  totals: {
    total_submissions: number
    unique_submitters: number
  }
  per_area: Record<string, RatingSummary | undefined>
  per_agenda: Record<string, RatingSummary | undefined>
}

/* --- Feedback -------------------------------------------------------------- */

/**
 * What a feedback submission is about: a single agenda, or a whole Area × Problem
 * cell. The form and the API accept exactly one of these shapes.
 */
export type FeedbackSubject =
  | { kind: 'agenda'; agenda: Agenda }
  | { kind: 'cell'; areaTag: string; areaName: string; problemId: string; problemName: string }

/* --- Swiss cheese interpretive overlay (Phase 2, operator-authored) -------- */

export interface SwissLayer {
  id: string
  name: string
  rationale: string
  /** Area tags and/or agenda ids that make up this defensive layer. */
  members: string[]
}

export interface SwissLayers {
  enabled: boolean
  default_hazard: string
  caveat: string
  layers: SwissLayer[]
}
