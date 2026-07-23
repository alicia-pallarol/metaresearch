#!/usr/bin/env python3
"""Regenerate data/atlas.json from the source workbooks, OPTIONAL, provenance only.

data/atlas.json is authoritative on its own and is committed. Nothing in this repo
needs the spreadsheets to build. This script exists for one case: the operator has
updated the workbooks and wants the JSON re-derived from them instead of hand-edited.

    python scripts/build_atlas.py            # dry run: parse, validate, report a diff
    python scripts/build_atlas.py --write    # also write data/atlas.json

With no workbooks present it prints one line and exits 0. It is never a build step
and never blocks anything.

WHAT IT READS
-------------
Any .xlsx in data/source/ (override with --source DIR). Two sheets are looked for,
matched case-insensitively on the sheet name:

  "Agendas"           one row per agenda, with a header row containing (in any
                      column order) these headers, matched loosely:
                          ID | Area | Agenda | Mechanism | Assumption |
                          Overall maturity | Best evidence against | Key work |
                          Source check
  "Agenda x Problem"  one row per agenda: an ID column plus one column per problem
                      id (P1 … P12). A cell is a tier name or empty.

Everything else, problems, areas, the legend, the priority write-ups and meta, is
carried over unchanged from the existing data/atlas.json, because those are prose
that lives in the JSON rather than in a grid. The condensed area_matrix is not read
at all: it is *derived* from the agenda matrix (best tier + how many agendas reach
the problem), which is exactly what it claims to be.

HONEST LIMITATION
-----------------
This parser was written against the documented layout above, not against the actual
workbooks, they are kept offline by the operator and were not available when the
repo was built. So treat the first run as an experiment: it refuses to write
anything that fails validation or that changes the headline counts, and it prints
what it would change before it changes it. If the headers in your workbook differ,
adjust HEADER_SYNONYMS below, that is the only place names are matched.

The supported, boring alternative is to edit data/atlas.json directly and run
`python scripts/validate_atlas.py`.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

sys.path.insert(0, str(Path(__file__).resolve().parent))
from validate_atlas import ValidationError, validate  # noqa: E402

REPO_ROOT = Path(__file__).resolve().parent.parent
ATLAS = REPO_ROOT / "data" / "atlas.json"
DEFAULT_SOURCE_DIR = REPO_ROOT / "data" / "source"

AGENDA_SHEET_NAMES = {"agendas", "agenda", "agendas sheet"}
MATRIX_SHEET_NAMES = {"agenda x problem", "agendaxproblem", "agenda × problem", "matrix"}

# Canonical field -> header spellings accepted for it. Comparison is done on a
# normalised form (lowercased, non-alphanumerics stripped).
HEADER_SYNONYMS: dict[str, tuple[str, ...]] = {
    "id": ("id", "agendaid", "code"),
    "area_name": ("area", "researcharea", "areaname"),
    "agenda": ("agenda", "agendaname", "name"),
    "mechanism": ("mechanism", "howitworks"),
    "assumption": ("assumption", "assumptionitrestson", "keyassumption"),
    "overall_maturity": ("overallmaturity", "maturity"),
    "best_evidence_against": ("bestevidenceagainst", "bestevidence"),
    "key_work": ("keywork", "keyworks", "references"),
    "source_check": ("sourcecheck", "sourcechecked"),
}

REQUIRED_AGENDA_FIELDS = ("id", "agenda", "mechanism", "assumption", "overall_maturity", "key_work", "source_check")


def norm(value: Any) -> str:
    return "".join(ch for ch in str(value or "").lower() if ch.isalnum())


def show_path(path: Path) -> str:
    """Repo-relative when it can be, absolute otherwise, never raises."""
    try:
        return str(path.resolve().relative_to(REPO_ROOT))
    except ValueError:
        return str(path)


def load_workbooks(source_dir: Path):
    """Returns (agenda_sheet, matrix_sheet) or (None, None) when nothing is there."""
    if not source_dir.exists():
        return None, None

    books = sorted(p for p in source_dir.glob("*.xlsx") if not p.name.startswith("~$"))
    if not books:
        return None, None

    try:
        from openpyxl import load_workbook  # type: ignore import-not-found
    except ImportError:
        print("build_atlas: found workbooks but openpyxl is not installed. `pip install openpyxl` to use them.")
        return None, None

    agenda_sheet = matrix_sheet = None
    for path in books:
        wb = load_workbook(path, data_only=True, read_only=True)
        for sheet_name in wb.sheetnames:
            key = norm(sheet_name)
            if agenda_sheet is None and key in {norm(n) for n in AGENDA_SHEET_NAMES}:
                agenda_sheet = (path.name, list(wb[sheet_name].values))
            elif matrix_sheet is None and key in {norm(n) for n in MATRIX_SHEET_NAMES}:
                matrix_sheet = (path.name, list(wb[sheet_name].values))
    return agenda_sheet, matrix_sheet


def find_header_row(rows: list[tuple], wanted: set[str]) -> int:
    """The header row is the first row that contains at least half of what we want."""
    for index, row in enumerate(rows[:20]):
        hits = sum(1 for cell in row if norm(cell) in wanted)
        if hits >= max(2, len(wanted) // 2):
            return index
    raise ValidationError(
        "could not find a header row; expected headers like: " + ", ".join(sorted(wanted))
    )


def map_columns(header: tuple) -> dict[str, int]:
    """canonical field name -> column index."""
    lookup = {}
    for field, spellings in HEADER_SYNONYMS.items():
        for index, cell in enumerate(header):
            if norm(cell) in {norm(s) for s in spellings}:
                lookup[field] = index
                break
    return lookup


def parse_agendas(rows: list[tuple], areas_by_name: dict[str, str]) -> list[dict]:
    wanted = {norm(s) for spellings in HEADER_SYNONYMS.values() for s in spellings}
    header_index = find_header_row(rows, wanted)
    columns = map_columns(rows[header_index])

    missing = [f for f in REQUIRED_AGENDA_FIELDS if f not in columns]
    if missing:
        raise ValidationError(
            "the Agendas sheet is missing columns for: "
            + ", ".join(missing)
            + ", add the spelling your workbook uses to HEADER_SYNONYMS in this script"
        )

    def cell(row: tuple, field: str) -> str:
        index = columns.get(field)
        if index is None or index >= len(row):
            return ""
        return str(row[index] or "").strip()

    agendas = []
    for row in rows[header_index + 1 :]:
        agenda_id = cell(row, "id")
        if not agenda_id:
            continue

        area_name = cell(row, "area_name")
        # The workbook may hold either the area's name or its tag.
        resolved = areas_by_name.get(norm(area_name))
        if resolved is None:
            raise ValidationError(f"{agenda_id}: unknown research area {area_name!r}")

        agendas.append(
            {
                "id": agenda_id,
                "area_tag": resolved[0],
                "area_name": resolved[1],
                "agenda": cell(row, "agenda"),
                "mechanism": cell(row, "mechanism"),
                "assumption": cell(row, "assumption"),
                "overall_maturity": cell(row, "overall_maturity"),
                "best_evidence_against": cell(row, "best_evidence_against"),
                "key_work": cell(row, "key_work"),
                "source_check": cell(row, "source_check"),
            }
        )
    return agendas


def parse_matrix(rows: list[tuple], problem_ids: list[str], tier_order: list[str]) -> dict[str, dict]:
    problem_lookup = {norm(p): p for p in problem_ids}
    tier_lookup = {norm(t): t for t in tier_order}

    header_index = find_header_row(rows, set(problem_lookup) | {"id", "agendaid"})
    header = rows[header_index]

    problem_columns: dict[int, str] = {}
    id_column = None
    for index, cell in enumerate(header):
        key = norm(cell)
        if key in problem_lookup:
            problem_columns[index] = problem_lookup[key]
        elif id_column is None and key in {"id", "agendaid", "code"}:
            id_column = index

    if id_column is None:
        raise ValidationError("the matrix sheet has no agenda id column")
    if len(problem_columns) != len(problem_ids):
        found = sorted(problem_columns.values())
        raise ValidationError(f"the matrix sheet covers {found}, expected all of {problem_ids}")

    matrix: dict[str, dict] = {}
    for row in rows[header_index + 1 :]:
        if id_column >= len(row):
            continue
        agenda_id = str(row[id_column] or "").strip()
        if not agenda_id:
            continue

        cells: dict[str, Any] = {}
        for index, problem_id in problem_columns.items():
            raw = str(row[index] or "").strip() if index < len(row) else ""
            if not raw:
                cells[problem_id] = None
                continue
            tier = tier_lookup.get(norm(raw))
            if tier is None:
                raise ValidationError(
                    f"{agenda_id} x {problem_id}: {raw!r} is not one of the tiers in legend.tier_order"
                )
            cells[problem_id] = tier
        matrix[agenda_id] = {pid: cells.get(pid) for pid in problem_ids}
    return matrix


def derive_area_matrix(atlas: dict) -> dict:
    """The condensed grid is a view of the agenda matrix, never a separate input."""
    tier_order = atlas["legend"]["tier_order"]
    rank = {tier: i for i, tier in enumerate(tier_order)}

    agendas_by_area: dict[str, list[str]] = {}
    for agenda in atlas["agendas"]:
        agendas_by_area.setdefault(agenda["area_name"], []).append(agenda["id"])

    area_order = [area["name"] for area in atlas["areas"]]
    cells: dict[str, dict] = {}
    for problem in atlas["problems"]:
        row: dict[str, Any] = {}
        for area_name in area_order:
            tiers = [
                atlas["matrix"][aid][problem["id"]]
                for aid in agendas_by_area.get(area_name, [])
                if atlas["matrix"][aid][problem["id"]] is not None
            ]
            row[area_name] = (
                None if not tiers else {"tier": min(tiers, key=lambda t: rank[t]), "count": len(tiers)}
            )
        cells[problem["id"]] = row

    return {"area_order": area_order, "cells": cells}


def summarise_diff(old: dict, new: dict) -> list[str]:
    notes = []
    if len(old["agendas"]) != len(new["agendas"]):
        notes.append(f"agendas: {len(old['agendas'])} -> {len(new['agendas'])}")

    old_by_id = {a["id"]: a for a in old["agendas"]}
    new_by_id = {a["id"]: a for a in new["agendas"]}
    for agenda_id in sorted(set(old_by_id) | set(new_by_id)):
        if agenda_id not in old_by_id:
            notes.append(f"new agenda {agenda_id}")
        elif agenda_id not in new_by_id:
            notes.append(f"removed agenda {agenda_id}")
        else:
            changed = [k for k, v in new_by_id[agenda_id].items() if old_by_id[agenda_id].get(k) != v]
            if changed:
                notes.append(f"{agenda_id}: changed {', '.join(changed)}")

    changed_cells = 0
    for agenda_id, row in new["matrix"].items():
        old_row = old["matrix"].get(agenda_id, {})
        changed_cells += sum(1 for pid, tier in row.items() if old_row.get(pid) != tier)
    if changed_cells:
        notes.append(f"{changed_cells} matrix cells differ")

    return notes or ["no differences"]


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--source", type=Path, default=DEFAULT_SOURCE_DIR, help="directory holding the .xlsx files")
    parser.add_argument("--write", action="store_true", help="write data/atlas.json (default is a dry run)")
    args = parser.parse_args(argv[1:])

    current = json.loads(ATLAS.read_text(encoding="utf-8"))

    agenda_sheet, matrix_sheet = load_workbooks(args.source)
    if agenda_sheet is None or matrix_sheet is None:
        print(
            f"build_atlas: nothing to parse from {show_path(args.source)} - nothing to do.\n"
            "            data/atlas.json is authoritative and already committed; this script is optional."
        )
        return 0

    print(f"build_atlas: Agendas from {agenda_sheet[0]}, matrix from {matrix_sheet[0]}")

    areas_by_name: dict[str, tuple[str, str]] = {}
    for area in current["areas"]:
        areas_by_name[norm(area["name"])] = (area["tag"], area["name"])
        areas_by_name[norm(area["tag"])] = (area["tag"], area["name"])

    try:
        agendas = parse_agendas(agenda_sheet[1], areas_by_name)
        problem_ids = [p["id"] for p in current["problems"]]
        matrix = parse_matrix(matrix_sheet[1], problem_ids, current["legend"]["tier_order"])
    except ValidationError as err:
        print(f"build_atlas: FAILED to parse the workbooks\n            {err}", file=sys.stderr)
        return 1

    rebuilt = dict(current)
    rebuilt["agendas"] = agendas
    rebuilt["matrix"] = matrix
    # areas[].agenda_ids follows the agenda list, so a new agenda lands in its area.
    rebuilt["areas"] = [
        {**area, "agenda_ids": [a["id"] for a in agendas if a["area_name"] == area["name"]]}
        for area in current["areas"]
    ]
    rebuilt["area_matrix"] = derive_area_matrix(rebuilt)

    try:
        validate(rebuilt)
    except ValidationError as err:
        print(
            f"build_atlas: the rebuilt atlas FAILED validation, nothing was written\n            {err}",
            file=sys.stderr,
        )
        return 1

    print("build_atlas: rebuilt atlas passes every check. Differences from the committed file:")
    for note in summarise_diff(current, rebuilt):
        print(f"  - {note}")

    if not args.write:
        print("build_atlas: dry run. Re-run with --write to overwrite data/atlas.json.")
        return 0

    ATLAS.write_text(json.dumps(rebuilt, indent=2, ensure_ascii=False) + "\n", encoding="utf-8", newline="\n")
    print(f"build_atlas: wrote {show_path(ATLAS)}")
    print("            next: python scripts/gen_agenda_ids.py && cd backend && go test ./...")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
