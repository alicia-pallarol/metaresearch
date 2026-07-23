/**
 * Loads the curated dataset and derives the lookups the UI needs.
 *
 * The dataset is imported, not fetched, so the whole map renders from the
 * bundle even when the API is asleep or unreachable.
 */
import atlasJson from '@data/atlas.json'
import swissJson from '@root/swiss_layers.json'
import type { Agenda, Area, Atlas, CellTier, Problem, SwissLayers, Tier } from '@/types/atlas'

export const atlas = atlasJson as unknown as Atlas
export const swissLayers = swissJson as unknown as SwissLayers

export const tierOrder: Tier[] = atlas.legend.tier_order

/** Rank in tier_order; lower is stronger evidence. Unknown tiers sort last. */
export function tierRank(tier: CellTier): number {
  if (tier === null) return tierOrder.length + 1
  const i = tierOrder.indexOf(tier)
  return i === -1 ? tierOrder.length : i
}

const agendaById = new Map<string, Agenda>(atlas.agendas.map((a) => [a.id, a]))
const problemById = new Map<string, Problem>(atlas.problems.map((p) => [p.id, p]))
const areaByName = new Map<string, Area>(atlas.areas.map((a) => [a.name, a]))
const areaByTag = new Map<string, Area>(atlas.areas.map((a) => [a.tag, a]))

export function getAgenda(id: string): Agenda | undefined {
  return agendaById.get(id)
}

export function getProblem(id: string): Problem | undefined {
  return problemById.get(id)
}

export function getAreaByName(name: string): Area | undefined {
  return areaByName.get(name)
}

export function getAreaByTag(tag: string): Area | undefined {
  return areaByTag.get(tag)
}

export const problemIds: string[] = atlas.problems.map((p) => p.id)

/** The tier for one agenda against one problem; null means "not addressed". */
export function cellTier(agendaId: string, problemId: string): CellTier {
  return atlas.matrix[agendaId]?.[problemId] ?? null
}

/** Agendas belonging to an area, in dataset order. */
export function agendasOfArea(areaName: string): Agenda[] {
  return atlas.agendas.filter((a) => a.area_name === areaName)
}

/**
 * The agendas of one area that actually reach one problem, strongest first.
 * This is what the drill-down panel lists: a condensed cell is only as strong
 * as its best agenda, so the reader should see which one that is.
 */
export function agendasReaching(areaName: string, problemId: string): Array<{ agenda: Agenda; tier: Tier }> {
  return agendasOfArea(areaName)
    .map((agenda) => ({ agenda, tier: cellTier(agenda.id, problemId) }))
    .filter((row): row is { agenda: Agenda; tier: Tier } => row.tier !== null)
    .sort((a, b) => tierRank(a.tier) - tierRank(b.tier))
}

/** How many problems an agenda is rated against at all. */
export function coverageOf(agendaId: string): number {
  const row = atlas.matrix[agendaId]
  if (!row) return 0
  return problemIds.filter((p) => row[p] !== null && row[p] !== undefined).length
}

/** Every problem an agenda reaches, with its tier, strongest first. */
export function problemsReachedBy(agendaId: string): Array<{ problem: Problem; tier: Tier }> {
  const row = atlas.matrix[agendaId] ?? {}
  return atlas.problems
    .map((problem) => ({ problem, tier: row[problem.id] ?? null }))
    .filter((r): r is { problem: Problem; tier: Tier } => r.tier !== null)
    .sort((a, b) => tierRank(a.tier) - tierRank(b.tier))
}

/** Counts per tier across the whole agenda-level matrix, for the legend. */
export function tierCounts(): Record<string, number> {
  const counts: Record<string, number> = {}
  for (const tier of tierOrder) counts[tier] = 0
  for (const agendaId of Object.keys(atlas.matrix)) {
    for (const problemId of problemIds) {
      const tier = cellTier(agendaId, problemId)
      if (tier !== null) counts[tier] = (counts[tier] ?? 0) + 1
    }
  }
  return counts
}

/** Short label for a problem, e.g. "P1 · Alignment measurement…". */
export function problemLabel(problemId: string): string {
  const p = getProblem(problemId)
  return p ? `${p.id} · ${p.name}` : problemId
}
