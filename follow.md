# Claude Code Prompt — Repo Deep Refactor + Claude Optimization (NO GIT COMMIT)

You are Claude Code running in my repository.

## Mode
**ultrathink**. Do a deep, repo-wide review (not just a surface pass).

## Primary goals
1. **Eliminate duplicate logic** across the repo by introducing **shared helper functions/modules** where appropriate.
2. **Refactor files to stay small**: **no source file should exceed ~500–600 lines**. If a file is larger, split it logically (modules/classes/helpers) without breaking public APIs.
3. **Optimize the repo for Claude context/token efficiency**:
   - Ensure files are **properly tagged/organized** (clear structure, consistent naming, predictable entry points).
   - Make it easy for Claude to understand the project layout quickly (clear folder boundaries, minimal redundancy).
4. **Security + code quality must be excellent** (public repo):
   - Remove insecure patterns, secrets risk, weak validation, etc.
   - Improve error handling, input validation, and logging where needed.
5. **Testing quality**
   - Unit test coverage must be **near 100%** (practically as high as possible).
   - `make ci` must run **all unit tests** and enforce coverage thresholds.
   - `make ci-full` must run **all integration tests** and they must pass, covering major logic paths.

## Hard constraints
- **DO NOT commit to git** (no `git commit`, no pushing). Local changes only.
- **Do not touch admin revoke-token skip tests**:
  - There are tests skipped for admin flows (revoke token, etc.). **Leave them skipped and do not modify those files.**
- If you add **new integration tests** that create remote resources, you **must clean up** those resources (teardown/finalizers).
- Keep refactors safe: do not introduce breaking changes unless you also update all callers/tests accordingly.

## Cleanup tasks
- Identify and remove **unnecessary/duplicative files** that shouldn’t be in the public repo.
  - Example: `REFACTORING_GUIDE.md` likely isn’t needed if the line-limit rule exists elsewhere.
- Check for duplicate **Claude rules / skills docs**:
  - Deduplicate and consolidate.
  - Keep only one canonical set of rules.
- Ensure repository documentation and rules are consistent, minimal, and non-redundant.

## Required workflow (do in this order)
1. **Inventory + structure map**
   - Print a concise repo structure overview.
   - Identify large files (>600 LOC) and areas with repeated logic.
2. **Duplicate logic audit**
   - Find duplicated functions/flows (utilities, parsing, API calls, validation, error mapping, etc.).
   - Propose helper modules and refactor plan.
3. **Refactor execution**
   - Implement shared helpers.
   - Split oversized files into smaller modules.
   - Keep imports/exports clean and discoverable.
   - Maintain consistent naming and folder conventions.
4. **Claude-optimization pass**
   - Ensure files are clearly grouped and tagged by responsibility.
   - Remove redundant docs/rules and keep a single source of truth.
5. **Tests**
   - Add/expand unit tests to reach near-100% coverage.
   - Ensure `make ci` runs unit tests + coverage.
   - Run `make ci-full` to validate integration tests.
   - If any integration test creates resources, ensure teardown is reliable.
   - Leave admin revoke-token skip tests untouched.

## Nylas API verification (when stuck)
If you get stuck on behavior/fields/endpoints:
- Consult **Nylas API V3 docs**.
- Validate assumptions with **curl** using my already-set environment:
  - `NYLAS_CLIENT_ID`
  - `NYLAS_API_KEY`
  - `NYLAS_GRANT_ID`
- Prefer minimal, reproducible curl calls to confirm request/response shapes.

## Playwright requirement
- There are Playwright tests for **Nylas Air** under `tests`.
- Check the `Makefile` targets and ensure Playwright tests run and pass as part of the appropriate CI target(s).
- Fix flaky selectors/timeouts if needed, but keep tests stable and deterministic.

## Output expectations (what you should report back)
- A clear summary of:
  - What duplicates were removed and what helpers were created
  - Which files were split and how the structure changed
  - Which unnecessary files were removed and why
  - What was deduplicated in Claude rules/skills/docs
  - Test additions and coverage result
  - Results of running:
    - `make ci`
    - `make ci-full`
- If anything cannot be completed, explain exactly why and what remains.

## Reminder
**Do not commit to git.**
