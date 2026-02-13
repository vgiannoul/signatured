# Changelog

## [Unreleased]

### Changed
- **Project renamed** from `signature-manager` to `signatured`
  - Binary name: `signature-manager` → `signatured`
  - Module path: `github.com/vgiannoul/signature-manager` → `github.com/vgiannoul/signatured`
  - Command directory: `cmd/signature-manager` → `cmd/signatured`
- **Template filename** changed from `signature.md` to `signatured.md`
  - Default template path updated in CLI flags
  - All documentation updated to reference new filename

### Added
- Conditional syntax support (`{{#if field}}...{{/if}}`) to hide sections when user data is missing
- Per-user Gmail API authentication for proper domain-wide delegation
- Comprehensive test coverage for conditional processing
- Detailed conditional syntax documentation in README

### Fixed
- Gmail API delegation error: Now creates per-user authenticated clients instead of single admin client
- Missing fields no longer create blank lines in signatures

## [1.0.0] - 2026-02-13

### Initial Release
- Single-binary CLI tool for Google Workspace signature management
- Markdown-based template system with placeholder support
- Google Workspace API integration (Directory + Gmail)
- Service account authentication with domain-wide delegation
- Batch operations with concurrency control
- Dry-run mode and structured logging
- Cross-platform builds (macOS, Linux, Windows)
