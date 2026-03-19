---
layout: default
title: Google Cloud Storage
nav_order: 2
description: "Load signature templates from Google Cloud Storage buckets"
permalink: /gcs-support/
---

# Google Cloud Storage Template Support
{: .no_toc }

Load signature templates from GCS buckets for centralized template management and cloud-native deployments.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

Signatured now supports loading templates from Google Cloud Storage (GCS) in addition to local files. This enables centralized template management and easier deployment in cloud environments.

## Features

- **Auto-detection**: Automatically detects local files vs GCS URLs
- **Multiple URL formats**: Supports both `gs://` and `https://storage.googleapis.com/` formats
- **Secure authentication**: Uses Application Default Credentials (same service account as Workspace APIs)
- **Minimal permissions**: Only requires `roles/storage.objectViewer` on the template bucket

## Supported URL Formats

### gs:// Protocol
```bash
gs://bucket-name/path/to/template.md
gs://my-org-signatures/templates/signatured.md
```

### HTTPS URL
```bash
https://storage.googleapis.com/bucket-name/path/to/template.md
https://storage.googleapis.com/my-org-signatures/templates/signatured.md
```

### Local File (existing behavior)
```bash
./templates/signatured.md
/absolute/path/to/template.md
```

## Setup

### 1. Create GCS Bucket

```bash
# Create bucket for templates
gsutil mb -p my-project-id -l us-central1 gs://my-org-signatures

# Enable versioning (recommended)
gsutil versioning set on gs://my-org-signatures

# Upload template
gsutil cp templates/signatured.md gs://my-org-signatures/templates/
```

### 2. Grant Service Account Access

Grant your service account read access to the bucket:

```bash
# Option 1: Using gsutil
gsutil iam ch serviceAccount:signatured@my-project.iam.gserviceaccount.com:objectViewer \
  gs://my-org-signatures

# Option 2: Using gcloud
gcloud storage buckets add-iam-policy-binding gs://my-org-signatures \
  --member="serviceAccount:signatured@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"
```

**Note**: Grant permission on the bucket, not project-level, following least privilege principle.

### 3. Verify Access

Test that the service account can access the template:

```bash
# Authenticate as the service account
gcloud auth activate-service-account --key-file=credentials.json

# Verify bucket access
gsutil ls gs://my-org-signatures/templates/

# Deactivate (return to your user account)
gcloud auth revoke signatured@my-project.iam.gserviceaccount.com
```

## Usage Examples

### Validate Template from GCS

```bash
# Using gs:// URL
./signatured validate --template gs://my-org-signatures/templates/signatured.md

# Using HTTPS URL
./signatured validate \
  --template https://storage.googleapis.com/my-org-signatures/templates/signatured.md
```

### Apply Signatures with GCS Template

```bash
# Single user
./signatured apply \
  --user alice@example.com \
  --impersonate admin@example.com \
  --template gs://my-org-signatures/templates/signatured.md

# Entire organization
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template gs://my-org-signatures/templates/signatured.md
```

### Dry Run with GCS Template

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template gs://my-org-signatures/templates/signatured.md \
  --dry-run
```

## IAM Permissions

### Required Role

**Service Account**: `roles/storage.objectViewer` on the template bucket

### Minimal Custom Role (Alternative)

If you prefer a more restrictive custom role:

```yaml
title: "Signatured Template Reader"
description: "Read-only access to signature templates"
stage: "GA"
includedPermissions:
- storage.objects.get
- storage.objects.list
```

Create and assign:

```bash
# Create custom role
gcloud iam roles create signaturedTemplateReader \
  --project=my-project-id \
  --title="Signatured Template Reader" \
  --description="Read-only access to signature templates" \
  --permissions=storage.objects.get,storage.objects.list

# Assign to service account
gsutil iam ch \
  serviceAccount:signatured@my-project.iam.gserviceaccount.com:roles/signaturedTemplateReader \
  gs://my-org-signatures
```

## Best Practices

### Template Management

1. **Version Control**: Enable GCS object versioning to track template changes
2. **Separate Buckets**: Use different buckets for dev/staging/prod templates
3. **Access Control**: Grant bucket access only to specific service accounts
4. **Audit Logging**: Enable Cloud Audit Logs for template access tracking

### Security

1. **Least Privilege**: Grant `objectViewer` only on template bucket, not project-wide
2. **Private Buckets**: Keep template buckets private (no `allUsers` or `allAuthenticatedUsers`)
3. **Uniform Access**: Use uniform bucket-level access (disable ACLs)
4. **VPC Service Controls**: Consider VPC-SC for additional isolation in production

### Deployment

```bash
# Development
--template gs://my-org-signatures-dev/templates/signatured.md

# Staging
--template gs://my-org-signatures-staging/templates/signatured.md

# Production
--template gs://my-org-signatures-prod/templates/signatured.md
```

## Troubleshooting

### Error: "failed to create GCS client"

**Cause**: Application Default Credentials not configured

**Solution**:
```bash
# Set credentials explicitly
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json

# Or use gcloud auth
gcloud auth application-default login
```

### Error: "failed to read object gs://bucket/template.md: Permission denied"

**Cause**: Service account lacks bucket access

**Solution**:
```bash
# Verify IAM bindings
gsutil iam get gs://my-org-signatures

# Grant access
gsutil iam ch serviceAccount:signatured@project.iam.gserviceaccount.com:objectViewer \
  gs://my-org-signatures
```

### Error: "invalid GCS URL format"

**Cause**: Malformed GCS URL

**Solution**: Ensure URL follows one of these formats:
- `gs://bucket-name/object-path`
- `https://storage.googleapis.com/bucket-name/object-path`

### Error: "object not found"

**Cause**: Template doesn't exist at specified path

**Solution**:
```bash
# List bucket contents
gsutil ls gs://my-org-signatures/templates/

# Upload template
gsutil cp templates/signatured.md gs://my-org-signatures/templates/
```

## Implementation Details

### Code Changes

Modified files:
- `internal/template/template.go`: Added GCS loading support
- `internal/template/gcs_test.go`: Added unit tests for GCS functionality
- `go.mod`: Added `cloud.google.com/go/storage` dependency

### Functions Added

- `isGCSPath(path string) bool`: Detects if path is a GCS URL
- `parseGCSURL(url string) (bucket, object string, err error)`: Parses GCS URLs
- `loadFromGCS(ctx context.Context, gcsURL string) ([]byte, error)`: Downloads from GCS

### Authentication Flow

1. Application uses `storage.NewClient()` with default credentials
2. Client uses service account credentials from `GOOGLE_APPLICATION_CREDENTIALS`
3. GCS client reuses same service account as Workspace APIs
4. No additional authentication configuration required

## Testing

Run unit tests:

```bash
# All template tests
go test ./internal/template/... -v

# Just GCS tests
go test ./internal/template/... -v -run TestIsGCSPath
go test ./internal/template/... -v -run TestParseGCSURL
```

## Migration Guide

### From Local Files to GCS

1. **Upload templates to GCS**:
   ```bash
   gsutil cp -r templates/ gs://my-org-signatures/
   ```

2. **Update command-line invocations**:
   ```bash
   # Before
   ./signatured apply --all --template ./templates/signatured.md

   # After
   ./signatured apply --all --template gs://my-org-signatures/templates/signatured.md
   ```

3. **Update automation scripts**:
   ```bash
   # In cron jobs, CI/CD, etc.
   sed -i 's|--template ./templates/|--template gs://my-org-signatures/templates/|g' scripts/*.sh
   ```

### Gradual Rollout

1. **Phase 1**: Test with single user
   ```bash
   ./signatured apply --user test@example.com \
     --template gs://my-org-signatures/templates/signatured.md
   ```

2. **Phase 2**: Test with small OU
   ```bash
   ./signatured apply --org-unit /IT \
     --template gs://my-org-signatures/templates/signatured.md
   ```

3. **Phase 3**: Roll out to all users
   ```bash
   ./signatured apply --all \
     --template gs://my-org-signatures/templates/signatured.md
   ```

## Cost Considerations

### GCS Pricing

For typical usage (500-user organization, 1 template read/day):

| Operation | Usage | Monthly Cost |
|-----------|-------|--------------|
| Storage | ~1 KB template | $0.00 |
| Class A Ops (reads) | 30 reads/month | $0.00 |
| Egress | ~30 KB/month | $0.00 |
| **Total** | | **~$0.00** |

**Note**: Effectively free for small-scale template storage and retrieval.

## Next Steps

- See [PLAN_EXTENSION.md](../PLAN_EXTENSION.md) for full cloud deployment with Cloud Run
- See [README.md](../README.md#google-cloud-storage-templates) for general usage
- See [terraform/](../terraform/) for infrastructure as code (coming soon)
