#!/usr/bin/env python3
"""Validate data/atlas.json.

The dataset is the single source of truth for everything the site shows, and it
is sometimes hand-edited when a new iteration is cut. This script is the check
that a hand edit did not break an invariant. It fails loudly and says which
invariant broke.

    python scripts/validate_atlas.py            # validate data/atlas.json
    python scripts/validate_atlas.py path.json  # validate another file

Exit code 0 means every check passed.
"""

from __future__ import annotations

import json
import sys
from collections import Counter
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_ATLAS = REPO_ROOT / "data" / "atlas.json"

# Counts of iteration 0. A deliberate change to the map should update these.
EXPECTED_AREAS = 12
EXPECTED_PROBLEMS = 12
EXPECTED_AGENDAS = 58
EXPECTED_PRIORITIES = 6


class ValidationError(Exception):
    """A broken invariant, phrased for whoever edited the file."""


def _require(condition: bool, message: str) -> None:
    if not condition:
        raise ValidationError(message)


def validate(atlas: dict) -> list[str]:
    """Check every invariant. Returns a list of human-readable facts checked."""
    checked: list[str] = []

    for key in ("meta", "legend", "problems", "areas", "agendas", "matrix", "area_matrix", "priorities"):
        _require(key in atlas, f"top-level key {key!r} is missing")
    checked.append("all top-level keys present")

    # --- counts -------------------------------------------------------------
    areas = atlas["areas"]
    problems = atlas["problems"]
    agendas = atlas["agendas"]

    _require(len(areas) == EXPECTED_AREAS, f"expected {EXPECTED_AREAS} areas, found {len(areas)}")
    _require(len(problems) == EXPECTED_PROBLEMS, f"expected {EXPECTED_PROBLEMS} problems, found {len(problems)}")
    _require(len(agendas) == EXPECTED_AGENDAS, f"expected {EXPECTED_AGENDAS} agendas, found {len(agendas)}")
    checked.append(f"{len(areas)} areas, {len(problems)} problems, {len(agendas)} agendas")

    # --- ids are unique and cross-referenced --------------------------------
    problem_ids = [p["id"] for p in problems]
    agenda_ids = [a["id"] for a in agendas]
    area_tags = [a["tag"] for a in areas]

    for name, ids in (("problem", problem_ids), ("agenda", agenda_ids), ("area tag", area_tags)):
        dupes = [k for k, n in Counter(ids).items() if n > 1]
        _require(not dupes, f"duplicate {name} ids: {dupes}")
    checked.append("ids are unique")

    area_names = {a["name"] for a in areas}
    tag_to_name = {a["tag"]: a["name"] for a in areas}

    listed_in_areas = [aid for area in areas for aid in area["agenda_ids"]]
    _require(
        sorted(listed_in_areas) == sorted(agenda_ids),
        "areas[].agenda_ids and agendas[] disagree about which agendas exist",
    )
    checked.append("every agenda belongs to exactly one area")

    for agenda in agendas:
        _require(
            agenda["area_tag"] in tag_to_name,
            f"{agenda['id']} has unknown area_tag {agenda['area_tag']!r}",
        )
        _require(
            agenda["area_name"] == tag_to_name[agenda["area_tag"]],
            f"{agenda['id']} area_name does not match its area_tag",
        )
    checked.append("agenda area_tag and area_name agree")

    # --- tiers --------------------------------------------------------------
    tier_order = atlas["legend"]["tier_order"]
    tier_set = set(tier_order)
    _require(len(tier_order) == len(tier_set), "tier_order contains duplicates")

    for tier in tier_order:
        _require(tier in atlas["legend"]["tiers"], f"tier {tier!r} has no definition in legend.tiers")
    checked.append(f"{len(tier_order)} tiers, each defined")

    for agenda in agendas:
        _require(
            agenda["overall_maturity"] in tier_set,
            f"{agenda['id']} overall_maturity {agenda['overall_maturity']!r} is not in tier_order",
        )

    # --- the matrix ---------------------------------------------------------
    matrix = atlas["matrix"]
    _require(
        sorted(matrix.keys()) == sorted(agenda_ids),
        "matrix rows and agenda ids disagree",
    )

    rated = 0
    for agenda_id, row in matrix.items():
        _require(
            sorted(row.keys()) == sorted(problem_ids),
            f"matrix row {agenda_id} does not cover exactly the {len(problem_ids)} problems",
        )
        for problem_id, tier in row.items():
            if tier is None:
                continue
            _require(
                tier in tier_set,
                f"matrix[{agenda_id}][{problem_id}] = {tier!r} is not in tier_order",
            )
            rated += 1
    checked.append(f"matrix is {len(matrix)} x {len(problem_ids)} with {rated} rated cells")

    # --- the condensed view must be derivable from the matrix ---------------
    rank = {tier: i for i, tier in enumerate(tier_order)}
    agendas_by_area: dict[str, list[str]] = {}
    for agenda in agendas:
        agendas_by_area.setdefault(agenda["area_name"], []).append(agenda["id"])

    area_matrix = atlas["area_matrix"]
    _require(
        sorted(area_matrix["area_order"]) == sorted(area_names),
        "area_matrix.area_order does not match the areas",
    )
    _require(
        sorted(area_matrix["cells"].keys()) == sorted(problem_ids),
        "area_matrix.cells does not cover exactly the problems",
    )

    for problem_id, cells in area_matrix["cells"].items():
        _require(
            sorted(cells.keys()) == sorted(area_names),
            f"area_matrix.cells[{problem_id}] does not cover exactly the areas",
        )
        for area_name, cell in cells.items():
            tiers = [
                matrix[aid][problem_id]
                for aid in agendas_by_area[area_name]
                if matrix[aid][problem_id] is not None
            ]
            expected = None if not tiers else {
                "tier": min(tiers, key=lambda t: rank[t]),
                "count": len(tiers),
            }
            _require(
                cell == expected,
                f"area_matrix.cells[{problem_id}][{area_name}] is {cell!r}, "
                f"but the agenda matrix says it should be {expected!r}",
            )
    checked.append("every condensed cell equals best-tier + count of its agendas")

    # --- priorities ---------------------------------------------------------
    priorities = atlas["priorities"]
    _require(
        len(priorities["items"]) == EXPECTED_PRIORITIES,
        f"expected {EXPECTED_PRIORITIES} priority items, found {len(priorities['items'])}",
    )
    for item in priorities["items"]:
        _require(
            item["problem_id"] in problem_ids,
            f"priority item points at unknown problem {item['problem_id']!r}",
        )
    checked.append(f"{len(priorities['items'])} priority items, all pointing at real problems")

    # --- meta ---------------------------------------------------------------
    _require(isinstance(atlas["meta"]["iteration"], int), "meta.iteration must be an integer")
    checked.append(f"meta.iteration = {atlas['meta']['iteration']}")

    return checked


def main(argv: list[str]) -> int:
    path = Path(argv[1]) if len(argv) > 1 else DEFAULT_ATLAS
    if not path.exists():
        print(f"atlas file not found: {path}", file=sys.stderr)
        return 2

    atlas = json.loads(path.read_text(encoding="utf-8"))

    try:
        checked = validate(atlas)
    except ValidationError as err:
        print(f"FAIL  {path}\n      {err}", file=sys.stderr)
        return 1

    print(f"OK  {path}")
    for line in checked:
        print(f"  - {line}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
