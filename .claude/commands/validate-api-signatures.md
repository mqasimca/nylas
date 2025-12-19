# Validate API Signatures

Verify that implementation matches Nylas v3 API documentation.

Endpoint/Feature: $ARGUMENTS

## Instructions

1. **Reference Nylas v3 API Documentation**

   Base documentation: https://developer.nylas.com/docs/api/v3/

   Key endpoints:
   - Messages: `/v3/grants/{grant_id}/messages`
   - Threads: `/v3/grants/{grant_id}/threads`
   - Drafts: `/v3/grants/{grant_id}/drafts`
   - Folders: `/v3/grants/{grant_id}/folders`
   - Calendars: `/v3/grants/{grant_id}/calendars`
   - Events: `/v3/grants/{grant_id}/events`
   - Contacts: `/v3/grants/{grant_id}/contacts`
   - Webhooks: `/v3/webhooks` (admin-level)

2. **Check implementation against docs**

   For each API method, verify:
   - HTTP method (GET, POST, PUT, DELETE, PATCH)
   - URL path and parameters
   - Query parameters
   - Request body structure
   - Response body structure
   - Required vs optional fields

3. **Compare domain types**

   Read domain types in `internal/domain/`:
   ```bash
   cat internal/domain/{feature}.go
   ```

   Compare JSON field names with API documentation:
   - Field names match exactly
   - Types are correct (string, int, bool, array, object)
   - Required fields are not omitempty
   - Optional fields have omitempty

4. **Compare adapter implementation**

   Read adapter in `internal/adapters/nylas/`:
   ```bash
   cat internal/adapters/nylas/{feature}.go
   ```

   Verify:
   - Correct HTTP method used
   - URL path matches API docs
   - Query parameters correctly encoded
   - Request body correctly structured
   - Response parsing matches expected format

5. **Validation checklist by endpoint**

   ```markdown
   ## Endpoint: {METHOD} /v3/{path}

   ### URL
   - [ ] Path matches documentation
   - [ ] Path parameters correctly substituted
   - [ ] Query parameters match (names, types)

   ### Request
   - [ ] HTTP method correct
   - [ ] Content-Type header (application/json)
   - [ ] Request body fields match docs
   - [ ] Required fields enforced

   ### Response
   - [ ] Response structure matches docs
   - [ ] All fields parsed
   - [ ] Pagination handled (next_cursor)
   - [ ] Error responses handled
   ```

6. **Common discrepancies to check**

   | Issue | How to Check |
   |-------|--------------|
   | Field name mismatch | Compare JSON tags with API docs |
   | Missing field | Check if new fields added to API |
   | Wrong type | Compare Go type with JSON schema |
   | Deprecated endpoint | Check API changelog |
   | Missing pagination | Verify next_cursor handling |
   | Wrong HTTP method | Check if PUT vs PATCH |

7. **Test against live API**

   ```bash
   # Set up credentials
   export NYLAS_API_KEY="your_api_key"
   export NYLAS_GRANT_ID="your_grant_id"

   # Test endpoint directly
   curl -X GET \
     -H "Authorization: Bearer $NYLAS_API_KEY" \
     -H "Content-Type: application/json" \
     "https://api.us.nylas.com/v3/grants/$NYLAS_GRANT_ID/{endpoint}"

   # Compare response with domain type
   ```

8. **Document findings**

   For each discrepancy found:
   ```markdown
   ### Discrepancy: {description}

   **Location:** `{file}:{line}`
   **API Docs:** {what docs say}
   **Implementation:** {what code does}
   **Fix:** {suggested fix}
   ```

## API Signature Verification Template

```markdown
# API Signature Verification: {Feature}

## Endpoints Checked

### GET /v3/grants/{grant_id}/{resource}s
- [ ] URL path correct
- [ ] Query params: limit, page_token, {others}
- [ ] Response: `{ data: [], next_cursor: "" }`
- [ ] Domain type: `domain.{Resource}`

### GET /v3/grants/{grant_id}/{resource}s/{id}
- [ ] URL path correct
- [ ] Response: `{ data: {} }`
- [ ] Domain type: `domain.{Resource}`

### POST /v3/grants/{grant_id}/{resource}s
- [ ] URL path correct
- [ ] Request body: `domain.Create{Resource}Request`
- [ ] Response: `{ data: {} }`

### PUT /v3/grants/{grant_id}/{resource}s/{id}
- [ ] URL path correct
- [ ] Request body: `domain.Update{Resource}Request`
- [ ] Response: `{ data: {} }`

### DELETE /v3/grants/{grant_id}/{resource}s/{id}
- [ ] URL path correct
- [ ] Response: 200 OK or 204 No Content

## Domain Type Verification

### domain.{Resource}
| Field | API Docs | Implementation | Match |
|-------|----------|----------------|-------|
| id | string | string | ✅ |
| name | string | string | ✅ |
| created_at | integer | int64 | ✅ |

### domain.Create{Resource}Request
| Field | API Docs | Implementation | Match |
|-------|----------|----------------|-------|
| name | string (required) | string | ✅ |

## Discrepancies Found

1. {None found / List discrepancies}

## Recommendations

1. {Any suggested changes}
```

## Quick Validation Commands

```bash
# Check domain types for JSON tags
grep -n "json:" internal/domain/{feature}.go

# Check API paths in adapter
grep -n "fmt.Sprintf.*grants" internal/adapters/nylas/{feature}.go

# Check HTTP methods used
grep -n "c.get\|c.post\|c.put\|c.delete" internal/adapters/nylas/{feature}.go

# Verify build (catches type mismatches)
go build ./...
```

## Checklist

- [ ] Read API documentation for endpoint
- [ ] Compared domain types with API schema
- [ ] Verified adapter implementation
- [ ] Checked URL paths and parameters
- [ ] Verified request/response structures
- [ ] Tested against live API (if possible)
- [ ] Documented any discrepancies
- [ ] Created fix tasks for issues found
