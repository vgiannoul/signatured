---
layout: default
title: Changelog
nav_order: 99
description: "Version history and release notes for signatured"
permalink: /changelog/
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **`--exclude` flag** (and `EXCLUDE_USERS` env var) to skip specific user emails on `apply`
  (e.g. shared mailboxes, service accounts). Excluded users are reported as skipped in the
  output and summary rather than silently dropped.

### Fixed
- **Release binaries reporting `dev` as their version**: `release.yml` builds now inject the git
  tag into `main.version` via `-ldflags`, matching what the Makefile already did for local builds

## [1.1.0] - 2026-09-30

### Added
- **`preview` command** to render a signature to a local HTML file without applying it
  - `--sample` uses built-in sample data (no Google API calls, no credentials needed)
  - `--user` fetches real Directory data for one user (read-only, no Gmail changes)
  - `--output` sets the destination file (default: `./signature-preview.html`)
- **Literal HTML template support**: templates with a `.html`/`.htm` extension are used as-is,
  skipping Markdown parsing entirely (`.md` templates are unaffected)
- `templates/example.html` - example HTML table template branded for a fictional company
  ("Signatured Co"), demonstrating the literal-HTML template format
- **Internal phone labels**: `{{phoneLabel}}` and `{{phoneIsInternal}}` let a template show a
  different label (e.g. "Ext.") next to a work phone that's an internal extension - marked in
  Directory as a "Custom" phone type whose custom type mentions "internal" - configured via new
  `COMPANY_PHONE_LABEL` / `COMPANY_INTERNAL_PHONE_LABEL` environment variables
- `Dockerfile` for running signatured in containers (e.g. as a scheduled Cloud Run Job), plus a
  CI job that builds and smoke-tests the image

### Fixed
- **HTML/script injection in rendered signatures**: placeholder values (e.g. a user's
  self-editable name or job title) are now HTML-escaped before substitution, so they can no
  longer break out of the surrounding HTML in templates rendered with raw-HTML support enabled
- **Query injection via `--org-unit`**: the flag value is now validated against an allowlist
  pattern before being used to build the Directory API query
- **Duplicate phone number in signatures**: if a user's only phone entry in Directory was typed
  "Mobile" (no "Work" entry), `{{phone}}` incorrectly fell back to that same mobile number,
  showing it twice in the signature. The fallback now excludes mobile-typed entries.

## [1.0.2] - 2026-03-19

### Added
- **Environment variable configuration** support for CLI defaults
  - `.env` file automatically loaded on startup
  - `TEMPLATE_PATH` - Set default template path (supports local files and GCS URLs)
  - `CREDENTIALS_PATH` - Set default credentials file path
  - `IMPERSONATE_USER` - Set default admin user for domain-wide delegation
  - `VERBOSE` - Enable verbose logging by default
  - Command-line flags take precedence over environment variables
- Comprehensive environment configuration documentation ([Environment Configuration Guide](env-config))
- Updated `.env.example` with detailed comments and examples

### Fixed
- **Version management** now works correctly
  - Changed version from `const` to `var` to allow build-time injection
  - Makefile reads version from `VERSION` file instead of hardcoded value
  - Builds without ldflags show "dev" version instead of outdated hardcoded version
- Cleaned up documentation to remove customer-specific references

## [1.0.1] - 2026-03-19

### Added
- **Google Cloud Storage (GCS) template support**
  - Load templates from GCS buckets in addition to local files
  - Auto-detection of GCS URLs (`gs://` and `https://storage.googleapis.com/`)
  - Support for both gs:// protocol and HTTPS URLs
  - Uses Application Default Credentials (same service account as Workspace APIs)
  - Minimal permissions required: `roles/storage.objectViewer` on template bucket
- New dependency: `cloud.google.com/go/storage@v1.61.3`
- Comprehensive GCS documentation ([GCS Support Guide](gcs-support))
  - Setup instructions
  - IAM permission requirements
  - Usage examples
  - Troubleshooting guide
  - Security best practices
- Unit tests for GCS functionality
  - URL detection tests
  - URL parsing tests for all supported formats
  - 74.7% code coverage in template package

### Changed
- Updated README with GCS template examples
- Modified template loader to support both local and remote templates
- Enhanced global flags documentation to indicate GCS URL support

## [1.0.0] - 2026-02-13

### Changed
- **Project renamed** from `signature-manager` to `signatured`
  - Binary name: `signature-manager` → `signatured`
  - Module path: `github.com/vgiannoul/signature-manager` → `github.com/vgiannoul/signatured`
  - Command directory: `cmd/signature-manager` → `cmd/signatured`
- **Template filename** changed from `signature.md` to `signatured.md`
  - Default template path updated in CLI flags
  - All documentation updated to reference new filename

### Added
- Conditional syntax support (`{% raw %}{{#if field}}...{{/if}}{% endraw %}`) to hide sections when user data is missing
- Per-user Gmail API authentication for proper domain-wide delegation
- Comprehensive test coverage for conditional processing
- Detailed conditional syntax documentation in README

### Fixed
- Gmail API delegation error: Now creates per-user authenticated clients instead of single admin client
- Missing fields no longer create blank lines in signatures

### Initial Release
- Single-binary CLI tool for Google Workspace signature management
- Markdown-based template system with placeholder support
- Google Workspace API integration (Directory + Gmail)
- Service account authentication with domain-wide delegation
- Batch operations with concurrency control
- Dry-run mode and structured logging
- Cross-platform builds (macOS, Linux, Windows)
