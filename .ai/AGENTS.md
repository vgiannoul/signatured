# AI Agent Instructions - signatured

**Project**: Google Workspace Signature Manager
**Language**: Go 1.26
**Current Version**: 1.0.2
**Versioning**: Semantic Versioning (SemVer 2.0.0)

## Project Overview

signatured is a single-binary CLI tool that automates email signature management for Google Workspace organizations. It reads Markdown-based signature templates, fetches user data from the Google Workspace Directory API, and applies signatures via the Gmail API.

### Key Features
- Single static binary with no runtime dependencies
- Template support: local files and Google Cloud Storage (GCS)
- Batch operations with concurrency control
- OAuth 2.0 service account authentication with domain-wide delegation
- Environment variable configuration via `.env` files
- Cross-platform support (macOS, Linux, Windows)

## Architecture

### Tech Stack
- **Language**: Go 1.26
- **CLI Framework**: github.com/spf13/cobra
- **Markdown Parser**: github.com/yuin/goldmark
- **Google APIs**: google.golang.org/api (Admin SDK Directory v1, Gmail v1)
- **OAuth**: golang.org/x/oauth2
- **GCS Support**: cloud.google.com/go/storage
- **Env Loading**: github.com/joho/godotenv

### Project Structure
```
signatured/
├── .ai/                    # AI agent instructions (this directory)
│   ├── AGENTS.md          # This file
│   ├── README.md          # Directory overview
│   ├── plans/             # Task plans (future)
│   └── memory/            # Session context (future)
├── .github/workflows/     # CI/CD automation
│   ├── ci.yml            # Test, lint, build on push/PR
│   ├── release.yml       # Automated releases on tags
│   └── docs.yml          # Documentation site deployment
├── cmd/signatured/       # Main CLI application
│   └── main.go           # Entry point, commands, logic
├── internal/
│   ├── google/           # Google API clients
│   │   ├── auth.go      # OAuth authentication
│   │   ├── directory.go # Directory API wrapper
│   │   └── gmail.go     # Gmail API wrapper
│   ├── models/           # Data models
│   │   └── user.go      # User model with placeholder mapping
│   └── template/         # Template engine
│       ├── template.go   # Parser and renderer
│       ├── gcs.go       # GCS template loading
│       └── *_test.go    # Unit tests
├── docs/                 # Documentation (Just the Docs site)
│   ├── index.md         # Documentation home
│   ├── ENV_CONFIG.md    # Environment configuration guide
│   ├── GCS_SUPPORT.md   # GCS template guide
│   └── changelog.md     # User-facing changelog
├── templates/            # Example signature templates
│   ├── signatured.md    # Default simple template
│   └── relevance.md     # Alternative template
├── .env.example          # Environment configuration template
├── .gitignore           # Git exclusions
├── CHANGELOG.md         # Developer changelog (Keep a Changelog format)
├── Makefile             # Development automation
├── VERSION              # Current version (single source of truth)
├── README.md            # Main user documentation
├── go.mod               # Go module definition
└── go.sum               # Dependency checksums
```

## Version Management

### Semantic Versioning
This project follows [SemVer 2.0.0](https://semver.org/spec/v2.0.0.html):

- **MAJOR.MINOR.PATCH** (e.g., 1.0.2)
- **MAJOR**: Breaking API changes
- **MINOR**: New features (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)

### Version Storage
The **single source of truth** for the current version is the `VERSION` file in the project root.

```bash
# Read current version
cat VERSION
# Output: 1.0.2
```

### Version Injection
The Makefile reads `VERSION` and injects it into the binary at build time:

```makefile
VERSION=$(shell cat VERSION 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
```

The version is stored as a `var` (not `const`) in `cmd/signatured/main.go` to allow build-time injection.

## Release Process

### Step-by-Step Release Workflow

#### 1. Prepare Release

**Update VERSION file:**
```bash
# For a patch release (bug fixes)
echo "1.0.3" > VERSION

# For a minor release (new features)
echo "1.1.0" > VERSION

# For a major release (breaking changes)
echo "2.0.0" > VERSION
```

**Update CHANGELOG.md:**
- Move items from `[Unreleased]` section to new version section
- Add release date in format `YYYY-MM-DD`
- Create new empty `[Unreleased]` section
- Follow [Keep a Changelog](https://keepachangelog.com) format

Example:
```markdown
## [Unreleased]

## [1.0.3] - 2026-03-20

### Fixed
- Fix template parsing for edge cases

### Added
- Support for new placeholder {{phoneMobile}}
```

**Update docs/changelog.md:**
- Mirror the CHANGELOG.md updates for user-facing documentation

#### 2. Commit and Push

```bash
# Stage version and changelog
git add VERSION CHANGELOG.md docs/changelog.md

# Commit with version number
git commit -m "Release v1.0.3"

# Push to main
git push origin main
```

#### 3. Create and Push Tag

```bash
# Create annotated tag (include 'v' prefix)
git tag -a v1.0.3 -m "Release v1.0.3"

# Push tag to trigger release workflow
git push origin v1.0.3
```

#### 4. Automated Release

The GitHub Actions `release.yml` workflow automatically:
1. Runs tests (`go test -v ./...`)
2. Builds binaries for all platforms:
   - `signatured-darwin-arm64` (macOS ARM64)
   - `signatured-darwin-amd64` (macOS Intel)
   - `signatured-linux-amd64` (Linux)
   - `signatured-windows-amd64.exe` (Windows)
3. Generates SHA256 checksums (`checksums.txt`)
4. Creates GitHub release with:
   - All binaries
   - Checksums file
   - Auto-generated release notes from commits
   - Changelog entries

#### 5. Verify Release

- Check [GitHub Releases](https://github.com/vgiannoul/signatured/releases)
- Verify all binaries are attached
- Confirm checksums are present
- Review auto-generated release notes

### Version Bumping Rules

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Bug fixes only | PATCH | 1.0.2 → 1.0.3 |
| New features (backwards-compatible) | MINOR | 1.0.3 → 1.1.0 |
| Breaking changes | MAJOR | 1.1.0 → 2.0.0 |
| Breaking changes (pre-1.0) | MINOR | 0.9.0 → 0.10.0 |

### Pre-release Versions

For beta/RC releases:
```bash
# Beta release
echo "1.1.0-beta.1" > VERSION
git tag -a v1.1.0-beta.1 -m "Beta release v1.1.0-beta.1"

# Release candidate
echo "2.0.0-rc.1" > VERSION
git tag -a v2.0.0-rc.1 -m "Release candidate v2.0.0-rc.1"
```

## Documentation Standards

### Documentation Locations

1. **README.md** - Main user documentation
   - Installation instructions
   - Setup guide (Google Cloud, domain-wide delegation)
   - Usage examples
   - Template syntax
   - Troubleshooting
   - Security best practices

2. **docs/** - Extended documentation (Just the Docs site)
   - `index.md` - Documentation home
   - `ENV_CONFIG.md` - Environment variable configuration
   - `GCS_SUPPORT.md` - Google Cloud Storage template support
   - `changelog.md` - User-facing changelog

3. **CHANGELOG.md** - Developer changelog
   - Follows [Keep a Changelog](https://keepachangelog.com) format
   - Sections: Added, Changed, Deprecated, Removed, Fixed, Security
   - Each version includes release date

4. **Code Comments**
   - Exported functions/types require godoc comments
   - Complex logic should have inline comments explaining WHY, not WHAT

### Updating Documentation

**When adding a feature:**
1. Add entry to `CHANGELOG.md` under `[Unreleased]` → `### Added`
2. Update relevant section in `README.md` (e.g., new flag, new feature)
3. Create new doc file in `docs/` if feature is substantial (e.g., `GCS_SUPPORT.md`)
4. Update code examples if usage changes

**When fixing a bug:**
1. Add entry to `CHANGELOG.md` under `[Unreleased]` → `### Fixed`
2. Update troubleshooting section in `README.md` if relevant

**When making breaking changes:**
1. Add entry to `CHANGELOG.md` under `[Unreleased]` → `### Changed`
2. Update migration guide in `README.md` or create `docs/MIGRATION.md`
3. Clearly document what changed and how users should adapt

### Documentation Style
- Use clear, direct language
- Provide complete, working examples
- Include expected output for CLI commands
- Use code blocks with language hints (```bash, ```go, ```markdown)
- Link to external docs when referencing Google APIs or libraries

## Development Workflow

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/vgiannoul/signatured.git
cd signatured

# Install dependencies
make deps

# Build binary
make build

# Run tests
make test
```

### Common Development Tasks

**Build for current platform:**
```bash
make build
# Produces: ./signatured
```

**Build for all platforms:**
```bash
make dist
# Produces: dist/signatured-{platform}-{arch}
```

**Run tests:**
```bash
make test                # Standard tests
make test-coverage       # Tests with coverage report
make run-tests-verbose   # Verbose output with race detector
```

**Code quality:**
```bash
make lint               # Run go vet and gofmt
make fmt                # Format code with go fmt
```

**Validate template:**
```bash
make validate-template  # Runs: ./signatured validate
```

**Dry run:**
```bash
# Requires TEST_USER and ADMIN_USER environment variables
export TEST_USER=user@example.com
export ADMIN_USER=admin@example.com
make dry-run
```

### Code Standards

**Follow global Go standards:**
- Use `gofmt` for formatting (enforced in CI)
- Run `go vet` before committing (enforced in CI)
- Avoid `panic()` in production code - return errors
- Use structured logging with `log/slog`

**Testing requirements:**
- Unit tests for template engine (`internal/template/*_test.go`)
- Test coverage for new features
- Use table-driven tests for multiple scenarios
- Mock external dependencies (Google APIs) in tests

**Error handling:**
- Return errors, don't panic
- Wrap errors with context: `fmt.Errorf("failed to load template: %w", err)`
- Log errors with structured fields: `slog.Error("operation failed", "error", err, "user", email)`

### Commit Message Format

Use conventional commit style:
```
<type>: <subject>

<body>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring (no functional changes)
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks (dependencies, build, etc.)

**Examples:**
```bash
git commit -m "feat: add support for phoneMobile placeholder"
git commit -m "fix: handle missing user data gracefully"
git commit -m "docs: update GCS setup instructions"
```

## Testing Requirements

### Test Coverage Standards

- **Template engine**: 100% coverage for core logic
- **New features**: Minimum 70% coverage
- **Bug fixes**: Add test case that reproduces the bug

### Running Tests

```bash
# All tests
go test -v ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# With race detector
go test -v -race ./...
```

### Test Locations

```
internal/
├── template/
│   ├── template.go
│   ├── template_test.go  # Template parsing and rendering tests
│   ├── gcs.go
│   └── gcs_test.go       # GCS URL detection and parsing tests
└── ...
```

### Writing Tests

**Table-driven tests:**
```go
func TestTemplateParsing(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"simple placeholder", "Hello {{firstName}}", "Hello John"},
        {"missing data", "Hello {{phone}}", "Hello "},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ParseTemplate(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

## CI/CD Pipeline

### GitHub Actions Workflows

#### 1. CI Workflow (`.github/workflows/ci.yml`)
**Triggers:** Push to `main`, Pull Requests to `main`

**Jobs:**
- **test**: Run tests with coverage, upload to Codecov
- **lint**: Run `go vet` and `gofmt`
- **build**: Build binaries for all platforms (Linux, macOS, Windows)

**Matrix:**
- GOOS: `[linux, darwin, windows]`
- GOARCH: `[amd64, arm64]`
- Excludes: `windows/arm64`

#### 2. Release Workflow (`.github/workflows/release.yml`)
**Triggers:** Push tags matching `v*` (e.g., `v1.0.3`)

**Steps:**
1. Checkout code with full history
2. Set up Go 1.26
3. Run tests
4. Build binaries for all platforms
5. Generate SHA256 checksums
6. Create GitHub release with binaries and checksums
7. Auto-generate release notes from commits

#### 3. Docs Workflow (`.github/workflows/docs.yml`)
**Triggers:** Push to `main`, changes to `docs/` directory

**Purpose:** Deploy documentation site using Just the Docs theme

### Action Security

All GitHub Actions are pinned to SHA hashes with version comments:
```yaml
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11  # v4.1.1
  with:
    persist-credentials: false
```

## Environment Configuration

### .env File Support

signatured loads environment variables from `.env` file automatically.

**Configuration precedence:**
1. Command-line flags (highest priority)
2. Environment variables (from `.env` or shell)
3. Hardcoded defaults (lowest priority)

### Environment Variables

#### CLI Defaults
```bash
TEMPLATE_PATH=./templates/signatured.md  # Local file or GCS URL
CREDENTIALS_PATH=./credentials.json      # Service account credentials
IMPERSONATE_USER=admin@example.com       # Admin user for delegation
VERBOSE=false                             # Enable debug logging
```

#### Company-Wide Settings
```bash
COMPANY_WEBSITE=https://example.com
COMPANY_LOGO=https://example.com/logo.png
COMPANY_PHONE=+1-555-0100
COMPANY_ADDRESS=123 Main St, City, State
```

### Security Notes

- **Never commit** `.env` or `credentials.json` to git (in `.gitignore`)
- Set file permissions: `chmod 600 .env credentials.json`
- Use Secret Manager in production (GCP Secret Manager, Vault, etc.)

## Common Agent Tasks

### Task: Add New Placeholder

1. **Update models** (`internal/models/user.go`):
   ```go
   func (u *User) ToPlaceholders() map[string]string {
       return map[string]string{
           "firstName":   u.FirstName,
           "newField":    u.NewField,  // Add here
       }
   }
   ```

2. **Add field extraction** (`internal/google/directory.go`):
   ```go
   NewField: extractNewField(user),
   ```

3. **Update documentation** (`README.md` - Template Syntax section):
   ```markdown
   | `{{newField}}` | Description | Source Field |
   ```

4. **Add test case** (`internal/template/template_test.go`):
   ```go
   {"new field", "{{newField}}", user, "ExpectedValue"},
   ```

5. **Update CHANGELOG.md**:
   ```markdown
   ### Added
   - Support for {{newField}} placeholder
   ```

### Task: Add New CLI Flag

1. **Add flag** (`cmd/signatured/main.go`):
   ```go
   applyCmd.Flags().String("new-flag", "default", "Description")
   ```

2. **Add to global flags if applicable** (persistent flags)

3. **Add environment variable support** if needed:
   ```go
   viper.BindEnv("new-flag", "NEW_FLAG")
   ```

4. **Update documentation** (`README.md` - Command Reference section)

5. **Update CHANGELOG.md**:
   ```markdown
   ### Added
   - New CLI flag `--new-flag` for feature X
   ```

### Task: Fix Bug

1. **Write failing test** that reproduces the bug

2. **Fix the bug** in minimal way

3. **Verify test passes**

4. **Update CHANGELOG.md**:
   ```markdown
   ### Fixed
   - Fix issue with X when Y occurs
   ```

5. **Commit**: `git commit -m "fix: handle edge case in template parsing"`

### Task: Update Dependencies

```bash
# Update all dependencies to latest compatible versions
go get -u ./...

# Update specific dependency
go get -u github.com/spf13/cobra@latest

# Tidy modules (remove unused, add missing)
make mod-tidy

# Run tests to verify
make test

# Commit changes
git add go.mod go.sum
git commit -m "chore: update dependencies"
```

### Task: Create New Documentation

1. **Create file in `docs/`**:
   ```bash
   touch docs/NEW_FEATURE.md
   ```

2. **Add front matter** (for Just the Docs):
   ```markdown
   ---
   layout: default
   title: New Feature
   nav_order: 5
   ---

   # New Feature Guide
   ```

3. **Write documentation** with examples

4. **Link from main docs** (`docs/index.md` or `README.md`)

5. **Update CHANGELOG.md**:
   ```markdown
   ### Added
   - Documentation for new feature at docs/NEW_FEATURE.md
   ```

## Troubleshooting

### Build Issues

**Issue**: `go build` fails with version error

**Solution**: Ensure `VERSION` file exists:
```bash
echo "1.0.2" > VERSION
make build
```

### Test Failures

**Issue**: Tests fail with GCS-related errors

**Solution**: GCS tests may require credentials. Mock GCS in tests or skip integration tests:
```bash
go test -v -short ./...  # Skip long-running tests
```

### Release Issues

**Issue**: GitHub release not triggered

**Solution**: Ensure tag has `v` prefix:
```bash
git tag -a v1.0.3 -m "Release v1.0.3"  # Correct
git tag -a 1.0.3 -m "Release 1.0.3"    # Wrong - missing 'v'
```

## Additional Resources

- **Google Workspace Admin SDK**: https://developers.google.com/admin-sdk
- **Gmail API**: https://developers.google.com/gmail/api
- **Go Documentation**: https://go.dev/doc/
- **Cobra CLI Framework**: https://github.com/spf13/cobra
- **Keep a Changelog**: https://keepachangelog.com
- **Semantic Versioning**: https://semver.org

## Contact and Support

For questions about this project:
- File an issue: https://github.com/vgiannoul/signatured/issues
- Read troubleshooting: [README.md](../README.md#troubleshooting)
- Review documentation: [docs/](../docs/)
