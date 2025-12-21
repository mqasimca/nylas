# Nylas CLI Utilities

This directory contains **non-Nylas API** utility services that provide offline-capable tools and features.

## Architecture

Follows the hexagonal architecture pattern:
- **Port Interface**: `internal/ports/utilities.go` - Defines service contracts
- **Adapters**: `internal/adapters/utilities/` - Implements the services
- **Domain Models**: `internal/domain/utilities.go` - Data structures
- **CLI Commands**: `internal/cli/` - User-facing commands

## Services

### 1. Time Zone Service (`timezone/`)
**Implements**: `ports.TimeZoneService`

**Purpose**: Time zone conversion and meeting time finder

**Features**:
- Convert times between time zones
- Find overlapping working hours across multiple zones
- Get DST transition dates
- List available IANA time zones
- Get detailed time zone information

**Pain Point Addressed**: 83% of professionals struggle with time zone scheduling

### 2. Webhook Service (`webhook/`)
**Implements**: `ports.WebhookService`

**Purpose**: Local webhook server for testing without ngrok

**Features**:
- Start/stop local HTTP server to receive webhooks
- Capture and display webhook payloads in real-time
- Validate webhook signatures (HMAC-SHA256)
- Save/load webhooks for replay
- Replay webhooks to target URLs

**Pain Point Addressed**: Developers frustrated by ngrok's URL changes on restart

### 3. Email Utility Service (`email/`)
**Implements**: `ports.EmailUtilityService`

**Purpose**: Email template building, validation, and deliverability checking

**Features**:
- Build email templates with variable substitution
- Preview templates with test data
- Check email deliverability (SPF/DKIM/DMARC)
- Sanitize HTML for email compatibility
- Inline CSS for email clients
- Parse/generate .eml files
- Validate email addresses (format + DNS MX)
- Analyze spam score using local rules

**Pain Points Addressed**:
- 50%+ of emails opened on mobile (need responsive design)
- 46% of email traffic is spam (need deliverability checks)
- Email testing across 100+ clients is expensive

### 4. Contact Utility Service (`contacts/`)
**Implements**: `ports.ContactUtilityService`

**Purpose**: Contact deduplication and vCard utilities

**Features**:
- Find and merge duplicate contacts (fuzzy matching)
- Parse/export vCard (.vcf) files
- Map vCard fields between providers (Outlook, Google, Nylas)
- Merge contacts with conflict resolution
- Import/export CSV files
- Enrich contacts (e.g., Gravatar lookup)

**Pain Point Addressed**: Data duplication and vCard field transfer issues

## Mock Implementation

`mock.go` provides a mock implementation for all utility services, useful for testing.

**Usage**:
```go
mockServices := utilities.NewMockUtilityServices()

// Customize behavior for specific tests
mockServices.ConvertTimeFunc = func(ctx context.Context, fromZone, toZone string, t time.Time) (time.Time, error) {
    return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), nil
}
```

## Development Guidelines

### Adding a New Utility Service

1. **Define interface** in `internal/ports/utilities.go`:
   ```go
   type MyNewService interface {
       DoSomething(ctx context.Context, input string) (string, error)
   }
   ```

2. **Add domain models** in `internal/domain/utilities.go`:
   ```go
   type MyRequest struct {
       Input string `json:"input"`
   }
   ```

3. **Create adapter** in `internal/adapters/utilities/mynew/`:
   ```go
   // service.go
   type Service struct{}

   func NewService() *Service {
       return &Service{}
   }

   func (s *Service) DoSomething(ctx context.Context, input string) (string, error) {
       // Implementation
   }
   ```

4. **Add mock** in `internal/adapters/utilities/mock.go`:
   ```go
   type MockUtilityServices struct {
       // ... existing fields
       DoSomethingFunc func(ctx context.Context, input string) (string, error)
   }
   ```

5. **Create CLI command** in `internal/cli/mynew/`:
   ```go
   func NewMyNewCmd() *cobra.Command {
       // CLI implementation
   }
   ```

6. **Register command** in `cmd/nylas/main.go`:
   ```go
   rootCmd.AddCommand(mynew.NewMyNewCmd())
   ```

### Testing

Each service should have:
- Unit tests (`*_test.go`)
- Mock implementations for dependencies
- Integration tests if applicable

### Linting

Always run before committing:
```bash
go fmt ./...
golangci-lint run --timeout=5m
go test ./... -short
make build
```

## Separation from Nylas API

**Why separate?**
- These utilities work **offline** without Nylas API access
- Provides value to users who don't have API credentials yet
- Privacy-focused: data never leaves the user's machine
- Lower barrier to entry for trying the CLI

**Key Differences**:
- Nylas API code: `internal/adapters/nylas/`
- Utility code: `internal/adapters/utilities/`
- Nylas uses HTTP client to call external API
- Utilities process data locally

## Future Enhancements

Planned features from `local/suggestions.md`:

**Phase 1** (High Priority):
- [ ] Complete time zone meeting finder logic
- [ ] Add CSS inlining for email templates
- [ ] Implement vCard parser (RFC 6350)
- [ ] Add webhook tunneling support

**Phase 2** (Medium Priority):
- [ ] AI-powered email categorization (local LLM)
- [ ] iCal (.ics) generator and parser
- [ ] Email signature generator
- [ ] Multi-calendar conflict detector

**Phase 3** (Advanced):
- [ ] Email thread analyzer & visualizer
- [ ] MIME message builder & inspector
- [ ] Migration helper (Gmail → Nylas)
- [ ] Batch operations planner

See `local/suggestions.md` for full roadmap with user research and pain points.

## Usage Examples

### Time Zone Conversion
```bash
nylas timezone convert --from America/Los_Angeles --to Asia/Kolkata
nylas timezone find --zones "PST,EST,IST" --duration 1h
```

### Webhook Testing
```bash
nylas webhook serve --port 3000 --persist
```

### Email Utilities
```bash
nylas email template create --preview --mobile
nylas email check-deliverability message.eml
```

### Contact Deduplication
```bash
nylas contacts dedupe --input contacts.vcf --fuzzy-match 0.8
```

## Contributing

When contributing to utilities:
1. Ensure feature works **100% offline**
2. Follow existing patterns (hexagonal architecture)
3. Add comprehensive tests
4. Update this README with new features
5. Run full quality checks (`make check`)

## Resources

- [Hexagonal Architecture](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software))
- [IANA Time Zone Database](https://www.iana.org/time-zones)
- [vCard RFC 6350](https://tools.ietf.org/html/rfc6350)
- [Email Deliverability Best Practices](https://www.nylas.com/blog/)
