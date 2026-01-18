---
description: Generate and analyze test coverage report
---

Generate test coverage for:
$ARGUMENTS

Commands to run:
!`go test -coverprofile=coverage.out $ARGUMENTS`
!`go tool cover -func=coverage.out`

Analyze the coverage report and identify:
1. Packages with low coverage (<80%)
2. Critical functions that need tests
3. Suggestions for improving coverage
