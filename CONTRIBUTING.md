# Contributing

Vantrel uses small, focused pull requests. Each change should be independently reviewable and should not implement unrelated roadmap work.

## Required Workflow

1. Start from `main`.
2. Pull the latest `main` from GitHub.
3. Confirm the working tree is clean.
4. Create a new branch for one task.
5. Make the smallest complete change for that task.
6. Add or update tests and documentation when relevant.
7. Run applicable tests, formatters and linters.
8. Review the diff before committing.
9. Commit with a meaningful message.
10. Push the branch and open a pull request to `main`.
11. Wait for review and merge before starting dependent work.

## Repository Rules

- Do not commit secrets.
- Do not commit employer code.
- Do not commit proprietary datasets.
- Do not include Centrica-specific source code or confidential concepts.
- Prefer dependencies compatible with an eventual open-source release.
- Keep domain boundaries explicit.
- Do not add placeholder implementations for future systems.

## Pull Request Checklist

Every pull request should include:

- Summary
- Architecture implications
- Tests executed
- Known limitations
- Recommended next task
