# Cloud-Native Deployment Plan

**Status**: 🚧 Partially Implemented
- ✅ GCS Template Support (v1.0.1)
- ⏳ Terraform Infrastructure (Planned)
- ⏳ Cloud Run Job Deployment (Planned)
- ⏳ Cloud Scheduler Automation (Planned)

## Overview

Extend signatured to support cloud-native deployment with:
1. ~~Template loading from Google Cloud Storage~~ ✅ **DONE in v1.0.1**
2. Terraform-based deployment as Cloud Run Job
3. Automated scheduling via Cloud Scheduler

---

## 1. Google Cloud Storage Template Support ✅

**Status**: Implemented in v1.0.1 (2026-03-15)

### Implementation Summary

Templates can now be loaded from GCS buckets in addition to local files.

**Supported URL formats**:
```bash
# Local file (original behavior)
--template ./templates/signatured.md

# GCS bucket URL
--template gs://my-org-signatures/templates/signatured.md

# GCS HTTPS URL
--template https://storage.googleapis.com/my-org-signatures/templates/signatured.md
```

**How it works**:
- Uses `cloud.google.com/go/storage` package
- Reuses existing service account credentials (Application Default Credentials)
- Auto-detects URL type and routes to appropriate loader
- Falls back to local file if not a GCS URL

**IAM Permissions**:
- Service account needs `roles/storage.objectViewer` on template bucket
- Grant access: `gsutil iam ch serviceAccount:signatured@project.iam.gserviceaccount.com:objectViewer gs://bucket-name`

**Documentation**: See [docs/GCS_SUPPORT.md](../../docs/GCS_SUPPORT.md)

**Testing**:
```bash
# Validate GCS template
./signatured validate --template gs://my-bucket/templates/signatured.md

# Apply with GCS template
./signatured apply --all --template gs://my-bucket/templates/signatured.md --impersonate admin@example.com
```

---

## 2. Terraform Infrastructure (Planned)

**Status**: ⏳ Not yet implemented

### Architecture

```
┌─────────────────────────────────────────────┐
│ Google Cloud Project                        │
│                                             │
│  ┌──────────────────┐                      │
│  │ Cloud Scheduler  │ (Cron: Daily 2 AM)   │
│  └────────┬─────────┘                      │
│           │ triggers                        │
│           v                                 │
│  ┌──────────────────┐                      │
│  │ Cloud Run Job    │                      │
│  │  - signatured    │                      │
│  │  - Run on demand │                      │
│  └────────┬─────────┘                      │
│           │ uses                            │
│           v                                 │
│  ┌──────────────────┐  ┌─────────────────┐│
│  │ Secret Manager   │  │ Cloud Storage   ││
│  │  - credentials   │  │  - templates    ││
│  │  - .env vars     │  │  - audit logs   ││
│  └──────────────────┘  └─────────────────┘│
│                                             │
│           ↓ accesses                        │
│  ┌──────────────────────────────────────┐  │
│  │ Google Workspace APIs                │  │
│  │  - Admin SDK (Directory)             │  │
│  │  - Gmail API                         │  │
│  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### Terraform Structure

```
terraform/
├── main.tf                 # Main configuration
├── variables.tf            # Input variables
├── outputs.tf              # Output values
├── versions.tf             # Provider versions
├── apis.tf                 # Enable required APIs
├── modules/
│   ├── cloud-run-job/     # Cloud Run Job module
│   ├── secrets/           # Secret Manager module
│   └── storage/           # GCS buckets module
├── environments/
│   ├── dev/
│   │   └── terraform.tfvars
│   ├── staging/
│   │   └── terraform.tfvars
│   └── prod/
│       └── terraform.tfvars
└── README.md
```

### Core Resources

#### Service Account
```hcl
resource "google_service_account" "signatured" {
  account_id   = "signatured"
  display_name = "Signatured Signature Manager"
  description  = "Service account for automated email signature management"
  project      = var.project_id
}
```

#### Secret Manager
```hcl
# Service account credentials
resource "google_secret_manager_secret" "credentials" {
  secret_id = "signatured-credentials"
  replication { auto {} }
}

# Environment variables (.env)
resource "google_secret_manager_secret" "env_vars" {
  secret_id = "signatured-env"
  replication { auto {} }
}

# Grant access
resource "google_secret_manager_secret_iam_member" "credentials_access" {
  secret_id = google_secret_manager_secret.credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.signatured.email}"
}
```

#### Cloud Storage Buckets
```hcl
# Templates bucket
resource "google_storage_bucket" "templates" {
  name     = "${var.project_id}-signatured-templates"
  location = var.region

  uniform_bucket_level_access = true
  versioning { enabled = true }

  lifecycle_rule {
    action { type = "Delete" }
    condition { num_newer_versions = 5 }
  }
}

# Audit logs bucket
resource "google_storage_bucket" "logs" {
  name     = "${var.project_id}-signatured-logs"
  location = var.region

  lifecycle_rule {
    action { type = "Delete" }
    condition { age = 90 }  # Delete logs older than 90 days
  }
}

# Grant permissions
resource "google_storage_bucket_iam_member" "template_reader" {
  bucket = google_storage_bucket.templates.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.signatured.email}"
}

resource "google_storage_bucket_iam_member" "log_writer" {
  bucket = google_storage_bucket.logs.name
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${google_service_account.signatured.email}"
}
```

#### Artifact Registry
```hcl
resource "google_artifact_registry_repository" "signatured" {
  location      = var.region
  repository_id = "signatured"
  format        = "DOCKER"
}
```

#### Cloud Run Job
```hcl
resource "google_cloud_run_v2_job" "signatured" {
  name     = "signatured"
  location = var.region

  template {
    template {
      service_account = google_service_account.signatured.email
      max_retries     = 1
      timeout         = "3600s"

      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/signatured/signatured:${var.image_tag}"

        args = [
          "apply",
          "--all",
          "--impersonate", var.impersonate_user,
          "--template", "gs://${google_storage_bucket.templates.name}/templates/signatured.md",
          "--credentials", "/secrets/credentials.json"
        ]

        volume_mounts {
          name       = "credentials"
          mount_path = "/secrets"
        }

        resources {
          limits = {
            cpu    = "1"
            memory = "512Mi"
          }
        }
      }

      volumes {
        name = "credentials"
        secret {
          secret       = google_secret_manager_secret.credentials.secret_id
          default_mode = 0400
          items {
            version = "latest"
            path    = "credentials.json"
          }
        }
      }
    }
  }
}
```

#### Cloud Scheduler
```hcl
resource "google_cloud_scheduler_job" "daily_sync" {
  name        = "signatured-daily-sync"
  description = "Daily signature synchronization at 2 AM"
  schedule    = "0 2 * * *"  # Cron: Daily at 2 AM
  time_zone   = var.timezone

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.signatured.name}:run"

    oauth_token {
      service_account_email = google_service_account.scheduler.email
    }
  }
}

# Separate service account for scheduler
resource "google_service_account" "scheduler" {
  account_id   = "signatured-scheduler"
  display_name = "Signatured Scheduler"
}

# Grant invoke permission
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  location = var.region
  name     = google_cloud_run_v2_job.signatured.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.scheduler.email}"
}
```

#### Enable APIs
```hcl
locals {
  required_apis = [
    "cloudscheduler.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com",
    "storage.googleapis.com",
    "artifactregistry.googleapis.com",
    "admin.googleapis.com",
    "gmail.googleapis.com",
  ]
}

resource "google_project_service" "required_apis" {
  for_each = toset(local.required_apis)
  service  = each.value
  disable_on_destroy = false
}
```

### Variables
```hcl
variable "project_id" {
  description = "Google Cloud project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud region"
  type        = string
  default     = "us-central1"
}

variable "impersonate_user" {
  description = "Admin user email for domain-wide delegation"
  type        = string
}

variable "company_website" {
  description = "Company website URL"
  type        = string
}

variable "company_logo" {
  description = "Company logo URL"
  type        = string
}

variable "company_phone" {
  description = "Company phone number"
  type        = string
}

variable "company_address" {
  description = "Company address"
  type        = string
}

variable "image_tag" {
  description = "Container image tag"
  type        = string
  default     = "latest"
}

variable "timezone" {
  description = "Timezone for scheduled jobs"
  type        = string
  default     = "America/Los_Angeles"
}

variable "schedule" {
  description = "Cron schedule for signature sync"
  type        = string
  default     = "0 2 * * *"
}
```

### Environment Configuration
```hcl
# terraform/environments/prod/terraform.tfvars
project_id       = "my-org-workspace-prod"
region           = "us-central1"
impersonate_user = "admin@example.com"

company_website  = "https://example.com"
company_logo     = "https://example.com/logo.png"
company_phone    = "+1-555-0100"
company_address  = "123 Main St, City, State 12345"

image_tag        = "1.0.2"
timezone         = "America/New_York"
schedule         = "0 2 * * *"
```

---

## 3. Container Image (Planned)

### Dockerfile
```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o signatured ./cmd/signatured

# Runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/signatured /app/signatured

USER nonroot:nonroot
ENTRYPOINT ["/app/signatured"]
```

### Build Script
```bash
#!/bin/bash
# scripts/build-and-push.sh

set -euo pipefail

PROJECT_ID="${1:-}"
REGION="${2:-us-central1}"
TAG="${3:-latest}"

if [ -z "$PROJECT_ID" ]; then
  echo "Usage: $0 <project-id> [region] [tag]"
  exit 1
fi

IMAGE_NAME="${REGION}-docker.pkg.dev/${PROJECT_ID}/signatured/signatured:${TAG}"

echo "Building container image..."
docker build -t "$IMAGE_NAME" .

echo "Pushing to Artifact Registry..."
docker push "$IMAGE_NAME"

echo "Image pushed: $IMAGE_NAME"
```

---

## 4. Deployment Guide (Future)

### Initial Deployment
```bash
# 1. Initialize Terraform
cd terraform/environments/prod
terraform init

# 2. Review plan
terraform plan -out=tfplan

# 3. Apply infrastructure
terraform apply tfplan

# 4. Build and push container
cd ../../..
./scripts/build-and-push.sh my-org-workspace-prod us-central1 1.0.2

# 5. Configure domain-wide delegation (manual step)
# - Get service account client ID from terraform output
# - Go to Google Workspace Admin Console
# - Security > API Controls > Domain-wide Delegation
# - Add client ID with scopes:
#   https://www.googleapis.com/auth/admin.directory.user.readonly,https://www.googleapis.com/auth/gmail.settings.basic

# 6. Upload template to GCS
gsutil cp templates/signatured.md \
  gs://my-org-workspace-prod-signatured-templates/templates/

# 7. Test manual execution
gcloud run jobs execute signatured \
  --project=my-org-workspace-prod \
  --region=us-central1 \
  --wait

# 8. Verify scheduler
gcloud scheduler jobs describe signatured-daily-sync \
  --project=my-org-workspace-prod \
  --location=us-central1
```

### Ongoing Operations
```bash
# Update template
gsutil cp templates/signatured.md gs://bucket/templates/

# Manual trigger
gcloud run jobs execute signatured --region=us-central1

# View logs
gcloud logging read "resource.type=cloud_run_job AND resource.labels.job_name=signatured" --limit=50

# Update container image
./scripts/build-and-push.sh my-project us-central1 1.0.3
terraform apply -var="image_tag=1.0.3"
```

---

## Security Best Practices

### IAM Principles
1. **Least Privilege**: Grant minimum required permissions
2. **Service Account per Function**: Separate SA for scheduler vs job
3. **No Long-Lived Keys**: Use Workload Identity where possible
4. **Secret Manager**: Never commit credentials to git
5. **Audit Logging**: Enable Cloud Audit Logs for all API calls

### Recommended IAM Roles

**Service Account (signatured)**:
- `roles/secretmanager.secretAccessor` (on specific secrets only)
- `roles/storage.objectViewer` (on template bucket only)
- `roles/storage.objectCreator` (on log bucket only)
- No project-level roles

**Scheduler Service Account**:
- `roles/run.invoker` (on Cloud Run Job only)

**Workspace APIs**:
- Configured via domain-wide delegation (not IAM)
- Scopes: `admin.directory.user.readonly`, `gmail.settings.basic`

---

## Monitoring and Alerting (Planned)

### Cloud Logging
```hcl
resource "google_logging_project_sink" "signatured_logs" {
  name        = "signatured-logs"
  destination = "bigquery.googleapis.com/projects/${var.project_id}/datasets/${google_bigquery_dataset.logs.dataset_id}"

  filter = <<-EOT
    resource.type="cloud_run_job"
    resource.labels.job_name="signatured"
  EOT

  unique_writer_identity = true
}
```

### Cloud Monitoring
```hcl
resource "google_monitoring_alert_policy" "job_failure" {
  display_name = "Signatured Job Failures"

  conditions {
    display_name = "Job execution failed"
    condition_matched_log {
      filter = <<-EOT
        resource.type="cloud_run_job"
        resource.labels.job_name="signatured"
        severity="ERROR"
      EOT
    }
  }

  notification_channels = [google_monitoring_notification_channel.email.id]
}

resource "google_monitoring_notification_channel" "email" {
  display_name = "Signatured Alerts"
  type         = "email"
  labels = { email_address = var.alert_email }
}
```

---

## Cost Estimation

### Monthly Cost (500 users, daily sync)

| Service | Usage | Monthly Cost |
|---------|-------|--------------|
| Cloud Run Job | 1 execution/day × 5 min | ~$0.22 |
| Cloud Scheduler | 1 job | $0.10 |
| Secret Manager | 2 secrets × 6 versions | $0.36 |
| Cloud Storage | 10 GB templates + logs | $0.20 |
| Artifact Registry | 500 MB images | $0.05 |
| Network Egress | ~1 GB | $0.12 |
| **Total** | | **~$1.05/month** |

**Note**: Workspace API calls are free for domain users.

---

## Migration Path

### Phase 1: Deploy Infrastructure (Week 1)
1. Create Terraform configuration
2. Run `terraform apply` to provision resources
3. Configure domain-wide delegation manually
4. Upload initial template to GCS
5. Test with single user (`--user` flag)

### Phase 2: Gradual Rollout (Week 2)
1. Test with small OU (`--org-unit /IT`)
2. Monitor logs and errors
3. Refine template based on feedback
4. Enable scheduled execution (initially disabled)

### Phase 3: Full Production (Week 3+)
1. Apply to all users (`--all`)
2. Enable daily scheduler
3. Set up monitoring alerts
4. Document runbook for common issues

---

## Success Criteria

Extension is complete when:
- ✅ Template can be loaded from GCS bucket URL (v1.0.1)
- ⏳ Terraform deploys all infrastructure to clean GCP project
- ⏳ Cloud Run Job executes successfully via scheduler
- ⏳ Secrets are managed via Secret Manager (no credentials in code)
- ⏳ Monitoring alerts trigger on failures
- ⏳ Documentation includes deployment guide
- ⏳ Cost is under $5/month for typical organization
- ⏳ Zero manual steps required after initial domain-wide delegation

---

## Open Questions

1. **Template Versioning**: Track template changes in GCS versioning or separate metadata?
   - **Recommendation**: Use GCS object versioning + Cloud Audit Logs

2. **Rollback Strategy**: How to revert to previous signatures if template has issues?
   - **Recommendation**: Store previous template version in GCS, create manual rollback script

3. **Multi-Region Deployment**: Should Cloud Run Job run in multiple regions for HA?
   - **Recommendation**: Single region sufficient for daily batch job; retry handles transient failures

4. **Template Validation**: Validate templates before applying to all users?
   - **Recommendation**: Add `validate` step in CI/CD before uploading to GCS

5. **Audit Requirements**: Track every signature change per user?
   - **Recommendation**: Cloud Logging captures all operations; export to BigQuery for long-term analysis

---

## References

- GCS Support Documentation: [docs/GCS_SUPPORT.md](../../docs/GCS_SUPPORT.md)
- Current implementation: See [AGENTS.md](../AGENTS.md)
- MVP history: See [mvp-history.md](mvp-history.md)
