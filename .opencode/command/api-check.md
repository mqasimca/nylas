---
description: Verify code uses Nylas API v3 only
agent: code-reviewer
subtask: true
---

Review the codebase to ensure only Nylas API v3 is used:
$ARGUMENTS

Check for:
1. No references to v1 or v2 API endpoints
2. All API calls use the correct v3 base URL
3. Proper error handling for API responses
4. Correct use of domain types from internal/domain/

API v3 documentation: https://developer.nylas.com/docs/api/v3/
