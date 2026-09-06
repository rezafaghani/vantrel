#!/usr/bin/env python3
import json
import re
import sys
from pathlib import Path

TOPIC = re.compile(r"^vantrel\.[a-z0-9-]+(?:\.[a-z0-9-]+)*\.v[0-9]+$")
POLICIES = {"delete", "compact"}
REQUIRED = {
    "name",
    "owner",
    "purpose",
    "key",
    "value_contract",
    "partitions",
    "replication_factor",
    "retention_ms",
    "cleanup_policy",
    "replayable",
}


def check(path: Path) -> list[str]:
    errors: list[str] = []
    try:
        data = json.loads(path.read_text())
    except json.JSONDecodeError as exc:
        return [f"{path}:{exc.lineno}: invalid json: {exc.msg}"]

    if data.get("version") != 1:
        errors.append(f"{path}: version must be 1")
    topics = data.get("topics")
    if not isinstance(topics, list) or not topics:
        errors.append(f"{path}: topics must be a non-empty list")
        return errors

    seen: set[str] = set()
    for i, topic in enumerate(topics):
        label = f"{path}:topics[{i}]"
        if not isinstance(topic, dict):
            errors.append(f"{label}: topic must be an object")
            continue
        missing = sorted(REQUIRED - topic.keys())
        if missing:
            errors.append(f"{label}: missing {', '.join(missing)}")
        name = topic.get("name")
        if not isinstance(name, str) or not TOPIC.match(name):
            errors.append(f"{label}: invalid topic name {name!r}")
        elif name in seen:
            errors.append(f"{label}: duplicate topic name {name}")
        elif isinstance(name, str):
            seen.add(name)
        for field in ("owner", "purpose", "key", "value_contract"):
            if not isinstance(topic.get(field), str) or not topic[field].strip():
                errors.append(f"{label}: {field} must be a non-empty string")
        if not isinstance(topic.get("partitions"), int) or topic["partitions"] < 1:
            errors.append(f"{label}: partitions must be a positive integer")
        if not isinstance(topic.get("replication_factor"), int) or topic["replication_factor"] < 1:
            errors.append(f"{label}: replication_factor must be a positive integer")
        if not isinstance(topic.get("retention_ms"), int) or topic["retention_ms"] < -1:
            errors.append(f"{label}: retention_ms must be -1 or greater")
        if topic.get("cleanup_policy") not in POLICIES:
            errors.append(f"{label}: cleanup_policy must be delete or compact")
        if not isinstance(topic.get("replayable"), bool):
            errors.append(f"{label}: replayable must be boolean")
    return errors


def main() -> int:
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("contracts/kafka/topics.json")
    errors = check(path)
    if errors:
        print("\n".join(errors), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
