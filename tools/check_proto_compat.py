#!/usr/bin/env python3
import re
import sys
from pathlib import Path

MESSAGE = re.compile(r"^message\s+(\w+)\s+\{")
FIELD = re.compile(r"^\s*(?:repeated\s+)?[\w.]+\s+\w+\s*=\s*(\d+)\s*;")
RESERVED = re.compile(r"^\s*reserved\s+(\d+)\s+to\s+(\d+)\s*;")


def check(path: Path) -> list[str]:
    errors: list[str] = []
    stack: list[tuple[str, set[int], list[range]]] = []
    for lineno, raw in enumerate(path.read_text().splitlines(), 1):
        line = raw.split("//", 1)[0].strip()
        if not line:
            continue
        if match := MESSAGE.match(line):
            stack.append((match.group(1), set(), []))
            continue
        if line == "}":
            stack.pop()
            continue
        if not stack:
            continue
        name, fields, reserved = stack[-1]
        if match := RESERVED.match(line):
            reserved.append(range(int(match.group(1)), int(match.group(2)) + 1))
            continue
        if match := FIELD.match(line):
            number = int(match.group(1))
            if number in fields:
                errors.append(f"{path}:{lineno}: duplicate field {number} in {name}")
            if any(number in r for r in reserved):
                errors.append(f"{path}:{lineno}: field {number} is reserved in {name}")
            fields.add(number)
    if stack:
        errors.append(f"{path}: unclosed message {stack[-1][0]}")
    return errors


def main() -> int:
    files = [Path(arg) for arg in sys.argv[1:]]
    if not files:
        files = sorted(Path("contracts/proto").rglob("*.proto"))
    errors = [error for path in files for error in check(path)]
    if errors:
        print("\n".join(errors), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
