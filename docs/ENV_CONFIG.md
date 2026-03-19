---
layout: default
title: Environment Configuration
nav_order: 3
description: "Configure signatured CLI defaults and company settings via environment variables"
permalink: /env-config/
---

# Environment Variable Configuration
{: .no_toc }

Configure CLI defaults and company settings using `.env` file for easier usage and deployment.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

Signatured automatically loads configuration from a `.env` file in the current directory. This allows you to set default values for CLI flags, making commands shorter and deployments easier.

## Precedence

Configuration precedence (highest to lowest):

1. **Command-line flags** (explicit `--flag` arguments)
2. **Environment variables** (from `.env` file or shell)
3. **Hardcoded defaults** (built into the application)

## Supported Variables

### CLI Defaults

| Variable | Flag | Default | Description |
|----------|------|---------|-------------|
| `TEMPLATE_PATH` | `--template` | `./templates/signatured.md` | Path to signature template (local or GCS) |
| `CREDENTIALS_PATH` | `--credentials` | `./credentials.json` | Path to service account credentials |
| `IMPERSONATE_USER` | `--impersonate` | *(none)* | Admin user email for domain-wide delegation |
| `VERBOSE` | `--verbose` | `false` | Enable verbose logging (`true` or `false`) |

### Company-Wide Settings

These variables are used in templates for company branding:

| Variable | Description | Template Placeholder |
|----------|-------------|---------------------|
| `COMPANY_WEBSITE` | Company website URL | `{{companyWebsite}}` |
| `COMPANY_LOGO` | Company logo image URL | `{{companyLogo}}` |
| `COMPANY_PHONE` | Company phone number | `{{companyPhone}}` |
| `COMPANY_ADDRESS` | Company physical address | `{{companyAddress}}` |

## Setup

### 1. Create .env File

```bash
cp .env.example .env
```

### 2. Configure Values

Edit `.env` with your settings:

```bash
# CLI defaults
TEMPLATE_PATH=./templates/signatured.md
CREDENTIALS_PATH=./credentials.json
IMPERSONATE_USER=admin@example.com
VERBOSE=false

# Company branding
COMPANY_WEBSITE=https://example.com
COMPANY_LOGO=https://example.com/logo.png
COMPANY_PHONE=+1-555-0100
COMPANY_ADDRESS=123 Main St, City, State 12345
```

### 3. Run Commands

Now you can run commands without specifying flags:

```bash
# Before (without .env)
./signatured apply --all \
  --template ./templates/signatured.md \
  --credentials ./credentials.json \
  --impersonate admin@example.com

# After (with .env)
./signatured apply --all
```

## Usage Examples

### Using .env Defaults

```bash
# .env file contains:
# TEMPLATE_PATH=./templates/signatured.md
# IMPERSONATE_USER=admin@example.com

# Command uses defaults from .env
./signatured validate
# Uses: --template ./templates/signatured.md

./signatured apply --all
# Uses: --template ./templates/signatured.md --impersonate admin@example.com
```

### Overriding with Flags

Flags take precedence over environment variables:

```bash
# .env file contains:
# TEMPLATE_PATH=./templates/signatured.md

# Override with explicit flag
./signatured apply --all --template ./templates/custom.md
# Uses: --template ./templates/custom.md (flag overrides .env)
```

### Google Cloud Storage Templates

Set GCS URLs in `.env`:

```bash
# Local development
TEMPLATE_PATH=./templates/signatured.md

# Production deployment
TEMPLATE_PATH=gs://my-org-signatures/templates/signatured.md
```

Then run without flags:

```bash
./signatured apply --all
# Automatically uses GCS template from .env
```

### Per-Environment Configuration

Use different `.env` files for different environments:

```bash
# Development
cp .env.dev .env
./signatured apply --all

# Production
cp .env.prod .env
./signatured apply --all
```

Example `.env.dev`:
```bash
TEMPLATE_PATH=./templates/signatured.md
IMPERSONATE_USER=admin-dev@example.com
VERBOSE=true
```

Example `.env.prod`:
```bash
TEMPLATE_PATH=gs://my-org-signatures/templates/signatured.md
IMPERSONATE_USER=admin@example.com
VERBOSE=false
```

## Cloud Deployment

### Cloud Run Job

Set environment variables in job configuration:

```bash
gcloud run jobs update signatured \
  --set-env-vars="TEMPLATE_PATH=gs://my-bucket/templates/signatured.md" \
  --set-env-vars="IMPERSONATE_USER=admin@example.com" \
  --region=us-central1
```

### Docker Container

Pass via `-e` flags:

```bash
docker run \
  -e TEMPLATE_PATH=gs://my-bucket/templates/signatured.md \
  -e IMPERSONATE_USER=admin@example.com \
  signatured apply --all
```

Or use `--env-file`:

```bash
docker run --env-file .env signatured apply --all
```

### Terraform

Set in Cloud Run Job resource:

```hcl
resource "google_cloud_run_v2_job" "signatured" {
  # ...

  template {
    template {
      containers {
        env {
          name  = "TEMPLATE_PATH"
          value = "gs://${google_storage_bucket.templates.name}/templates/signatured.md"
        }
        env {
          name  = "IMPERSONATE_USER"
          value = var.impersonate_user
        }
        # ...
      }
    }
  }
}
```

## Security

### Sensitive Variables

**Never commit `.env` to version control!**

The `.env` file is already in `.gitignore`, but be careful:

```bash
# Check .gitignore includes .env
grep "^\.env$" .gitignore

# Verify .env is not tracked
git status --ignored | grep .env
```

### Credentials in .env

Avoid storing credentials directly in `.env`:

❌ **Bad**:
```bash
GOOGLE_APPLICATION_CREDENTIALS_JSON='{"type":"service_account",...}'
```

✅ **Good**:
```bash
CREDENTIALS_PATH=./credentials.json
```

Then keep `credentials.json` secure:
```bash
chmod 600 credentials.json
```

### Cloud Deployments

For cloud deployments, use secret management:

- **Cloud Run**: Use Secret Manager
- **Docker**: Mount secrets as files
- **Kubernetes**: Use Kubernetes Secrets

Don't embed secrets in environment variables in cloud environments.

## Troubleshooting

### .env File Not Loaded

**Check file location**:
```bash
# .env must be in current working directory
pwd
ls -la .env
```

**Check file format**:
```bash
# No spaces around =
# Correct: TEMPLATE_PATH=./templates/signatured.md
# Wrong:   TEMPLATE_PATH = ./templates/signatured.md
```

### Environment Variable Not Used

**Check precedence**:
```bash
# If flag is specified, it overrides .env
./signatured apply --all --template custom.md
# Uses custom.md, not TEMPLATE_PATH from .env
```

**Check variable name**:
```bash
# Variable names are case-sensitive
# Correct: TEMPLATE_PATH
# Wrong:   template_path
```

### Verify Current Configuration

Use `--help` to see effective defaults:

```bash
./signatured apply --help
# Global Flags:
#   --template string      Path to signature template file (default "./templates/signatured.md")
#   --impersonate string   User email to impersonate... (default "admin@example.com")
```

The defaults shown reflect values from `.env`.

## Examples

### Minimal .env

For simple usage:

```bash
IMPERSONATE_USER=admin@example.com
```

Then run:

```bash
./signatured apply --all
# No need to specify --impersonate
```

### Complete .env

For production deployments:

```bash
# CLI configuration
TEMPLATE_PATH=gs://acme-corp-signatures/templates/signatured.md
CREDENTIALS_PATH=/secrets/credentials.json
IMPERSONATE_USER=admin@acme.com
VERBOSE=false

# Company branding
COMPANY_WEBSITE=https://acme.com
COMPANY_LOGO=https://acme.com/assets/logo.png
COMPANY_PHONE=+1-555-0100
COMPANY_ADDRESS=123 Main Street, San Francisco, CA 94105

# Testing
TEST_USER=test-user@acme.com
```

### Test-Specific .env

For local testing:

```bash
TEMPLATE_PATH=./templates/test-template.md
IMPERSONATE_USER=test-admin@example.com
VERBOSE=true
TEST_USER=alice@example.com
```

Run dry-run test:

```bash
./signatured apply --user $TEST_USER --dry-run
# Uses TEST_USER from .env
```

## Next Steps

- See [README.md](../README.md) for general usage
- See [GCS_SUPPORT.md](GCS_SUPPORT.md) for Google Cloud Storage templates
- See [PLAN_EXTENSION.md](../PLAN_EXTENSION.md) for cloud deployment
