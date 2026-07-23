#!/usr/bin/env python3
"""Regenerate backend/internal/atlas/ids.go from data/atlas.json.

The backend does not need the curated map, only the identifiers it must recognise
so it can reject feedback for things that do not exist: the agenda ids, the area
tags, the problem ids, the tier vocabulary a maturity vote may use, and which area
each agenda belongs to (so a vote on a subarea can be checked against the area it
was cast under). They live in a generated Go file rather than an embedded copy of
the dataset, so the API carries no content it could serve stale.

    python scripts/gen_agenda_ids.py

backend/internal/atlas/ids_test.go fails if this file drifts from the dataset, so
run it after any change to the agendas, areas or problems.
"""

from __future__ import annotations

import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
ATLAS = REPO_ROOT / "data" / "atlas.json"
TARGET = REPO_ROOT / "backend" / "internal" / "atlas" / "ids.go"

PER_LINE = 6

HEADER = """// Package atlas carries the identifiers the API must recognise.
//
// The curated map itself is static and shipped with the frontend; the backend only
// needs the id sets so it can reject feedback for things that do not exist.
// Generated from data/atlas.json by scripts/gen_agenda_ids.py. Do not edit by hand;
// ids_test.go fails loudly if this file drifts from the dataset.
package atlas
"""


AREA_OF_AGENDA = """// AreaOfAgenda returns the area tag an agenda belongs to.
func AreaOfAgenda(id string) (string, bool) {
\ttag, ok := agendaArea[id]
\treturn tag, ok
}"""


def go_slice(name: str, comment: str, values: list[str]) -> str:
    lines = [f"// {comment}", f"var {name} = []string{{"]
    for start in range(0, len(values), PER_LINE):
        chunk = values[start : start + PER_LINE]
        lines.append("\t" + " ".join(f'"{v}",' for v in chunk))
    lines.append("}")
    return "\n".join(lines)


def go_quoted_slice(name: str, comment: str, values: list[str]) -> str:
    """One value per line, for strings too long to read six to a row."""
    lines = [f"// {comment}", f"var {name} = []string{{"]
    lines.extend(f'\t"{v}",' for v in values)
    lines.append("}")
    return "\n".join(lines)


def go_map(name: str, comment: str, pairs: list[tuple[str, str]]) -> str:
    lines = [f"// {comment}", f"var {name} = map[string]string{{"]
    lines.extend(f'\t"{k}": "{v}",' for k, v in pairs)
    lines.append("}")
    return "\n".join(lines)


def go_set_and_lookup(set_name: str, list_name: str, fn_name: str, doc: str) -> str:
    return f"""var {set_name} = func() map[string]struct{{}} {{
\tm := make(map[string]struct{{}}, len({list_name}))
\tfor _, id := range {list_name} {{
\t\tm[id] = struct{{}}{{}}
\t}}
\treturn m
}}()

// {fn_name} reports whether id is {doc}.
func {fn_name}(id string) bool {{
\t_, ok := {set_name}[id]
\treturn ok
}}"""


def main() -> int:
    atlas = json.loads(ATLAS.read_text(encoding="utf-8"))
    agenda_ids = [a["id"] for a in atlas["agendas"]]
    area_tags = [a["tag"] for a in atlas["areas"]]
    problem_ids = [p["id"] for p in atlas["problems"]]
    tiers = atlas["legend"]["tier_order"]
    agenda_areas = [(a["id"], a["area_tag"]) for a in atlas["agendas"]]

    parts = [
        HEADER,
        "",
        go_slice("Iteration0AgendaIDs", f"Iteration0AgendaIDs are the {len(agenda_ids)} agenda ids of the curated map.", agenda_ids),
        "",
        go_slice("Iteration0AreaTags", f"Iteration0AreaTags are the {len(area_tags)} research-area tags.", area_tags),
        "",
        go_slice("Iteration0ProblemIDs", f"Iteration0ProblemIDs are the {len(problem_ids)} problem ids.", problem_ids),
        "",
        go_quoted_slice(
            "Iteration0Tiers",
            f"Iteration0Tiers are the {len(tiers)} maturity tiers of legend.tier_order,\n"
            "// strongest evidence first. A maturity vote must name one of these.",
            tiers,
        ),
        "",
        go_map(
            "agendaArea",
            "agendaArea maps each agenda to the research area it belongs to, so a vote\n"
            "// on a subarea can be checked against the area it was cast under.",
            agenda_areas,
        ),
        "",
        AREA_OF_AGENDA,
        "",
        go_set_and_lookup("agendaIDSet", "Iteration0AgendaIDs", "KnownAgendaID", "one of the curated agendas"),
        "",
        go_set_and_lookup("areaTagSet", "Iteration0AreaTags", "KnownAreaTag", "one of the research-area tags"),
        "",
        go_set_and_lookup("problemIDSet", "Iteration0ProblemIDs", "KnownProblemID", "one of the problem ids"),
        "",
        go_set_and_lookup("tierSet", "Iteration0Tiers", "KnownTier", "one of the maturity tiers"),
        "",
    ]

    TARGET.write_text("\n".join(parts), encoding="utf-8", newline="\n")
    print(
        f"wrote {len(agenda_ids)} agendas, {len(area_tags)} areas, {len(problem_ids)} problems, "
        f"{len(tiers)} tiers to {TARGET.relative_to(REPO_ROOT)}"
    )
    print("run `cd backend && gofmt -w internal/atlas/ids.go && go test ./...` to confirm")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
