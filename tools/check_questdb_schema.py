#!/usr/bin/env python3
import re
import sys
from pathlib import Path

REQUIRED_TABLES = {
    "market_actual_observations": {"series_id", "event_time", "value", "unit", "revision"},
    "market_forecast_observations": {
        "series_id",
        "forecast_run_id",
        "issued_at",
        "target_time",
        "horizon_seconds",
        "value",
        "unit",
    },
}
FORBIDDEN = {"canonical_name", "display_name", "description", "license", "owner", "data_product"}
DRIFT = {"SERIAL", "JSONB", "UUID", "ENGINE =", "MergeTree"}


def check(path: Path) -> list[str]:
    text = path.read_text()
    errors: list[str] = []
    upper = text.upper()
    for marker in DRIFT:
        if marker.upper() in upper:
            errors.append(f"{path}: non-QuestDB marker {marker}")
    tables = parse_tables(text)
    for table, columns in REQUIRED_TABLES.items():
        if table not in tables:
            errors.append(f"{path}: missing table {table}")
            continue
        missing = sorted(columns - tables[table])
        if missing:
            errors.append(f"{path}: {table} missing {', '.join(missing)}")
        leaked = sorted(FORBIDDEN & tables[table])
        if leaked:
            errors.append(f"{path}: {table} includes Catalog semantics {', '.join(leaked)}")
    return errors


def parse_tables(text: str) -> dict[str, set[str]]:
    out: dict[str, set[str]] = {}
    pattern = re.compile(r"CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)\s*\((.*?)\)\s*TIMESTAMP", re.I | re.S)
    for match in pattern.finditer(text):
        columns = set()
        for raw in match.group(2).splitlines():
            line = raw.strip().rstrip(",")
            if line and not line.startswith("--"):
                columns.add(line.split()[0].lower())
        out[match.group(1).lower()] = columns
    return out


def main() -> int:
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("contracts/questdb/market_observations.sql")
    errors = check(path)
    if errors:
        print("\n".join(errors), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
