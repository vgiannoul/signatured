# MVP History - signatured

**Status**: ✅ Complete (v1.0.0 released 2026-02-13)

This document captures the original design decisions and implementation summary for the MVP release.

## Original Design Decisions

### Language Choice: Go vs Python

| Criteria | Go | Python |
|----------|----|----|
| Single binary | ✅ Native | ⚠️ Requires PyInstaller |
| Binary size | ~10MB | ~50MB+ (bundled interpreter) |
| Startup time | <10ms | ~100ms |
| API clients | `google.golang.org/api` (official) | `google-api-python-client` (official) |
| Deployment | Single file copy | Single file, larger |
| Maintenance | Stricter typing, compile-time checks | Faster prototyping |

**Decision**: Go - native single-binary support, smaller footprint, better suited for CLI distribution.

### Architecture

```
┌──────────────┐
│ signatured.md │ (Markdown template with placeholders)
└──────┬───────┘
       │
       v
┌─────────────────────────────────────┐
│   Signature Manager CLI             │
│                                     │
│  1. Parse template                  │
│  2. Authenticate (service account)  │
│  3. Fetch user data (Directory API) │
│  4. Render signature per user       │
│  5. Apply via Gmail API             │
└─────────────────────────────────────┘
       │
       v
┌─────────────────────────────────────┐
│  Google Workspace APIs              │
│  - Admin SDK Directory API          │
│  - Gmail API (settings.sendAs)      │
└─────────────────────────────────────┘
```

### Core Technical Decisions

**Template Format**:
- Markdown with placeholder syntax: `{{firstName}}`, `{{lastName}}`, etc.
- Convert to HTML using goldmark parser
- Support conditional blocks: `{{#if field}}...{{/if}}`

**Authentication**:
- Service account with domain-wide delegation
- OAuth scopes: `admin.directory.user.readonly`, `gmail.settings.basic`
- Minimal permissions (least privilege principle)

**CLI Framework**:
- Cobra for command structure
- Flags: `--user`, `--org-unit`, `--all`, `--dry-run`, `--template`, `--credentials`, `--verbose`
- Structured logging with `log/slog`

## Implementation Phases

### Phase 1: Template & Placeholder Engine ✅
- Markdown parser using `github.com/yuin/goldmark`
- Placeholder replacement system (7 fields)
- Conditional block support
- Unit tests with 100% pass rate

### Phase 2: Google API Integration ✅
- Service account authentication
- Directory API client (list/get users)
- Gmail API client (update signatures)
- Rate limiting and retry logic with exponential backoff

### Phase 3: CLI Interface ✅
- Command structure with Cobra
- Flag parsing and validation
- Dry-run mode implementation
- Progress tracking and summary reports

### Phase 4: Production Readiness ✅
- Structured logging
- GitHub Actions CI/CD pipelines
- Binary release pipeline
- Comprehensive documentation (README, setup guide, troubleshooting)

## What Was Built

### Core Features

1. **Template Engine**
   - Markdown-based signature templates
   - 7 placeholder fields + 4 company-wide fields
   - Conditional syntax for optional fields
   - Graceful handling of missing user data

2. **Google Workspace Integration**
   - Service account authentication with domain-wide delegation
   - Directory API client (fetch users, list by OU)
   - Gmail API client (update signatures)
   - Automatic retry with exponential backoff for rate limits

3. **CLI Application**
   - Two main commands: `validate` and `apply`
   - Multiple targeting options (single user, OU, entire domain)
   - Dry-run mode for safe testing
   - Configurable concurrency (default: 10)
   - Environment variable configuration via `.env`

4. **Developer Experience**
   - Makefile with common tasks
   - GitHub Actions CI/CD (test, lint, build, release)
   - Cross-platform builds (macOS ARM64/AMD64, Linux, Windows)
   - Comprehensive documentation

### Placeholder System

| Placeholder | Google Directory API Field | Fallback |
|-------------|---------------------------|----------|
| `{{firstName}}` | `name.givenName` | "" |
| `{{lastName}}` | `name.familyName` | "" |
| `{{email}}` | `primaryEmail` | *required* |
| `{{phone}}` | `phones[0].value` (type=work) | "" |
| `{{phoneMobile}}` | `phones[type=mobile].value` | "" |
| `{{orgUnit}}` | `orgUnitPath` | "" |
| `{{jobTitle}}` | `organizations[0].title` | "" |
| `{{organization}}` | `organizations[0].name` | "" |

Company-wide fields (from `.env`):
- `{{companyWebsite}}`, `{{companyLogo}}`, `{{companyPhone}}`, `{{companyAddress}}`

### Binary Characteristics

- **Single file**: No runtime dependencies
- **Size**: ~10-12MB (compiled, stripped)
- **Platforms**: macOS (ARM64/AMD64), Linux (AMD64), Windows (AMD64)
- **Startup**: <10ms
- **Performance**: Concurrent API calls, 10+ users/second

## Success Criteria (All Met)

- ✅ Single binary builds for Linux, macOS, Windows
- ✅ Reads `signatured.md` template with placeholders
- ✅ Authenticates with service account
- ✅ Fetches user data from Directory API
- ✅ Applies signatures via Gmail API
- ✅ Supports `--user`, `--org-unit`, `--all` flags
- ✅ Implements `--dry-run` mode
- ✅ Handles API errors gracefully
- ✅ Produces audit logs
- ✅ Documentation includes setup guide

## Post-MVP Enhancements

### v1.0.1 - GCS Template Support
- Load templates from Google Cloud Storage buckets
- Support `gs://` and `https://storage.googleapis.com/` URLs
- Uses Application Default Credentials
- Minimal permissions: `roles/storage.objectViewer`

### v1.0.2 - Environment Configuration
- `.env` file automatically loaded on startup
- Environment variables for CLI defaults
- Command-line flags take precedence
- Fixed version management (VERSION file as single source of truth)

## Future Enhancement Ideas

Ideas considered but not implemented in MVP:
- Multiple signature templates per department/role
- Rollback capability (restore previous signatures)
- Web UI for template editing
- Integration with HRIS for auto-sync
- Multi-language signature support
- A/B testing different formats
- Template approval workflow
- Self-service portal for template customization

See [cloud-native-deployment.md](cloud-native-deployment.md) for cloud deployment plans.

## Key Learnings

1. **Gmail API Delegation**: Initial implementation had issues with domain-wide delegation. Solution: Create per-user authenticated clients instead of single admin client.

2. **Conditional Templates**: Added `{{#if field}}...{{/if}}` syntax to prevent blank lines when user data is missing. Critical for clean signatures.

3. **Version Injection**: Version must be `var` (not `const`) in Go to allow build-time injection via ldflags.

4. **Rate Limiting**: Exponential backoff essential for large organizations. Default concurrency of 10 works well for most use cases.

5. **Environment Variables**: Users prefer `.env` files over command-line flags for regular operations. Implemented in v1.0.2 based on feedback.

## References

- Original planning document: Consolidated into this file
- Implementation summary: Consolidated into this file
- Current project state: See [AGENTS.md](../AGENTS.md)
- Future plans: See [cloud-native-deployment.md](cloud-native-deployment.md)
