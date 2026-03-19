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
