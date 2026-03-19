<div align="center">
  <img src="assets/logo.svg" alt="signatured logo" width="400">

  # Google Workspace Signature Manager

  Single-binary CLI tool to manage email signatures for Google Workspace organization members.
</div>

## Features

- Single static binary (no runtime dependencies)
- Markdown-based signature templates with placeholder support
- Template loading from local files or Google Cloud Storage
- Batch signature updates for entire organization or specific OUs
- OAuth 2.0 service account authentication with domain-wide delegation
- Rate limiting and automatic retry with exponential backoff
- Dry-run mode to preview changes
- Concurrent API calls for fast processing
- Structured logging for audit trail

## Installation

### Download Pre-built Binary

Download the latest release for your platform:

```bash
# macOS (ARM64)
curl -L https://github.com/vgiannoul/signatured/releases/latest/download/signatured-darwin-arm64 -o signatured
chmod +x signatured

# macOS (Intel)
curl -L https://github.com/vgiannoul/signatured/releases/latest/download/signatured-darwin-amd64 -o signatured
chmod +x signatured

# Linux
curl -L https://github.com/vgiannoul/signatured/releases/latest/download/signatured-linux-amd64 -o signatured
chmod +x signatured

# Windows
# Download signatured-windows-amd64.exe from releases page
```

### Build from Source

```bash
git clone https://github.com/vgiannoul/signatured.git
cd signatured
go build -o signatured ./cmd/signatured
```

## Prerequisites

Before setting up signatured, ensure you have:

- Google Workspace admin access
- Access to Google Cloud Console
- Binary downloaded or built from source

## Setup

Follow these steps to configure signatured for your Google Workspace organization.

### 1. Google Cloud Project Setup

1. Navigate to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project or select an existing one

### 2. Create Service Account

1. Go to **IAM & Admin** > **Service Accounts**
2. Click **Create Service Account**
3. Enter name: `signatured`
4. Description: `Automated email signature management`
5. Click **Create and Continue**
6. Skip role assignment (not needed for domain-wide delegation)
7. Click **Done**

### 3. Enable Required APIs

1. Go to **APIs & Services** > **Library**
2. Search and enable each:
   - ✅ **Admin SDK API**
   - ✅ **Gmail API**

### 4. Download Service Account Credentials

1. Click on the `signatured` service account
2. Go to **Keys** tab
3. Click **Add Key** > **Create new key**
4. Select **JSON** format
5. Click **Create**
6. Save the downloaded file as `credentials.json` in your project directory
7. Set secure permissions: `chmod 600 credentials.json`

**IMPORTANT**: Never commit `credentials.json` to version control!

### 5. Get Client ID for Domain-Wide Delegation

1. In the service account details page
2. Copy the **Client ID** (numeric, e.g., `123456789012345678901`)
### 6. Configure Domain-Wide Delegation

1. Navigate to [Google Workspace Admin Console](https://admin.google.com)
2. Go to **Security** > **Access and data control** > **API controls**
3. Click **Manage Domain Wide Delegation**
4. Click **Add new**
5. Enter the **Client ID** from step 5
6. Add **OAuth scopes** (comma-separated):
   ```
   https://www.googleapis.com/auth/admin.directory.user.readonly,https://www.googleapis.com/auth/gmail.settings.basic
   ```
7. Click **Authorize**

### 7. Create Signature Template (Optional)

The default template is located at `templates/signatured.md`. You can customize it or create your own:

```markdown
**{{firstName}} {{lastName}}**
{{#if jobTitle}}{{jobTitle}}
{{/if}}{{#if organization}}{{organization}}

{{/if}}---

{{#if phone}}📞 {{phone}}
{{/if}}✉️ {{email}}
{{#if orgUnit}}🏢 {{orgUnit}}{{/if}}
```

**Note**: The `{{#if field}}...{{/if}}` syntax automatically hides sections when a user's profile is missing that field, preventing awkward blank lines in signatures.

### 8. Configure Environment

Copy `.env.example` to `.env` and configure your settings:

```bash
cp .env.example .env
```

Then edit `.env` with your values:

```bash
# Google Workspace Signature Manager Configuration

# Path to signature template (supports local files and GCS URLs)
TEMPLATE_PATH=./templates/signatured.md
# Or use GCS: TEMPLATE_PATH=gs://my-bucket/templates/signatured.md

# Path to service account credentials JSON file
CREDENTIALS_PATH=./credentials.json

# Admin user email for domain-wide delegation
IMPERSONATE_USER=admin@example.com

# Enable verbose logging (true/false)
VERBOSE=false

# Company-wide signature configuration
COMPANY_WEBSITE=https://example.com
COMPANY_LOGO=https://example.com/logo.png
COMPANY_PHONE=+1-555-0100
COMPANY_ADDRESS=123 Main St, City, State 12345
```

**IMPORTANT**: Never commit `.env` to version control! It's already in `.gitignore`.

**Environment Variable Precedence**:
- Command-line flags override environment variables
- Environment variables override hardcoded defaults
- Example: `./signatured apply --all` uses `TEMPLATE_PATH` from `.env`
- Example: `./signatured apply --all --template custom.md` uses `custom.md` (flag overrides `.env`)

### 9. Verify Setup

After completing setup, verify you have:

```
signatured/                # Project directory
├── signatured             # Binary (executable)
├── credentials.json       # Service account key (600 permissions)
├── .env                   # Environment configuration (gitignored)
├── .env.example           # Example environment configuration
├── templates/             # Signature templates
│   ├── signatured.md     # Default simple template
│   └── html-table.md     # HTML table template
└── .gitignore             # Excludes credentials.json and .env
```

### 10. Validate Template

Test that your template is valid:

```bash
./signatured validate
```

Expected output:
```
time=2026-02-13T10:00:00.000+02:00 level=INFO msg="Validating template" path=./templates/signatured.md
time=2026-02-13T10:00:00.001+02:00 level=INFO msg="Template is valid"
```

### 11. Run Dry Run Test

Test with a single user (replace with your emails):

```bash
./signatured apply \
  --user testuser@example.com \
  --impersonate admin@example.com \
  --dry-run
```

Expected output:
```
...
✓ testuser@example.com - would update signature (dry run)

Summary:
  Success: 1
  Failed:  0
  Skipped: 0
  Total:   1
  Duration: 1.2s
```

### 12. Apply to Single User

Remove `--dry-run` to actually apply:

```bash
./signatured apply \
  --user testuser@example.com \
  --impersonate admin@example.com
```

### 13. Verify Signature Update

1. Open Gmail as `testuser@example.com`
2. Click **Settings** (gear icon)
3. Scroll to **Signature** section
4. Verify signature is updated correctly

### 14. Roll Out to Organization

Once verified, apply to all users:

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com
```

### Template Syntax

#### Placeholders

##### User-Specific Fields

| Placeholder | Description | Source Field |
|-------------|-------------|--------------|
| `{{firstName}}` | First name | `name.givenName` |
| `{{lastName}}` | Last name | `name.familyName` |
| `{{email}}` | Email address | `primaryEmail` |
| `{{phone}}` | Phone number (prefers work) | `phones[type=work].value` |
| `{{phoneMobile}}` | Mobile phone number | `phones[type=mobile].value` |
| `{{orgUnit}}` | Organizational unit path | `orgUnitPath` |
| `{{jobTitle}}` | Job title | `organizations[0].title` |
| `{{organization}}` | Organization name | `organizations[0].name` |

##### Company-Wide Fields

These fields are set via `.env` file and apply to all users:

| Placeholder | Description | Environment Variable |
|-------------|-------------|---------------------|
| `{{companyWebsite}}` | Company website URL | `COMPANY_WEBSITE` |
| `{{companyLogo}}` | Company logo URL | `COMPANY_LOGO` |
| `{{companyPhone}}` | Company phone number | `COMPANY_PHONE` |
| `{{companyAddress}}` | Company address | `COMPANY_ADDRESS` |

**Note**: The top-level organizational unit (`/`) is treated as empty when using conditionals, so `{{#if orgUnit}}` will hide content for users in the root organization.

#### Conditional Blocks

The signature manager supports conditional blocks that automatically hide sections when user data is missing, preventing awkward blank lines in signatures.

**Syntax:**
```markdown
{{#if fieldName}}content{{/if}}
```

- If `fieldName` has a value → renders `content`
- If `fieldName` is empty → removes entire block

**Example 1: Phone Number**

Template:
```markdown
{{#if phone}}📞 {{phone}}
{{/if}}✉️ {{email}}
```

User WITH phone number:
```
📞 +1-555-0100
✉️ alice@example.com
```

User WITHOUT phone number:
```
✉️ alice@example.com
```

Notice: No blank line where the phone would be.

**Example 2: Multiple Optional Fields**

Template:
```markdown
**{{firstName}} {{lastName}}**
{{#if jobTitle}}{{jobTitle}}
{{/if}}{{#if organization}}{{organization}}

{{/if}}---

{{#if phone}}📞 {{phone}}
{{/if}}✉️ {{email}}
{{#if orgUnit}}🏢 {{orgUnit}}{{/if}}
```

Full profile user:
```
John Doe
Software Engineer
Acme Corp

---

📞 +1-555-0100
✉️ john@example.com
🏢 /Engineering
```

Minimal profile user (only name and email):
```
Jane Smith

---

✉️ jane@example.com
```

**Best Practices:**

1. **Always use conditionals for optional fields** - Fields like `phone`, `jobTitle`, and `organization` that users might not have should always use conditionals:

   ✅ Good:
   ```markdown
   {{#if phone}}📞 {{phone}}
   {{/if}}
   ```

   ❌ Bad:
   ```markdown
   📞 {{phone}}
   ```
   *This leaves "📞 " in the signature even when phone is missing*

2. **Include newlines inside conditionals** - If you want the newline to disappear when the field is missing:

   ✅ Good:
   ```markdown
   {{#if phone}}📞 {{phone}}
   {{/if}}✉️ {{email}}
   ```

   ❌ Bad:
   ```markdown
   {{#if phone}}📞 {{phone}}{{/if}}

   ✉️ {{email}}
   ```
   *This leaves a blank line when phone is missing*

3. **Required fields don't need conditionals** - Fields like `firstName`, `lastName`, and `email` are typically always present:

   ```markdown
   **{{firstName}} {{lastName}}**
   ✉️ {{email}}
   ```

4. **Conditionals work with markdown** - You can use markdown formatting inside conditional blocks:

   ```markdown
   {{#if jobTitle}}**Job:** *{{jobTitle}}*
   {{/if}}
   ```

**Common Patterns:**

Contact information block:
```markdown
{{#if phone}}📞 {{phone}}
{{/if}}✉️ {{email}}
{{#if website}}🌐 {{website}}{{/if}}
```

Job title and company:
```markdown
{{#if jobTitle}}{{jobTitle}}{{#if organization}} at {{organization}}{{/if}}
{{/if}}
```

This prevents awkward blank lines in signatures for users with incomplete profile data.

#### Built-in Templates

The project includes pre-built templates in the `templates/` directory:

- **signatured.md** (default) - Simple markdown-based template
- **templates/html-table.md** - Professional HTML table layout with company branding

To use a built-in HTML template, configure company settings in `.env` and run:

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template ./templates/html-table.md
```

The company information (`COMPANY_WEBSITE`, `COMPANY_LOGO`, etc.) from your `.env` file will be automatically applied to all user signatures.

You can also create custom templates by combining markdown formatting with HTML for more advanced layouts.

## Usage

### Validate Template

Test that your template is valid:

```bash
./signatured validate
```

### Apply to Single User

```bash
./signatured apply \
  --user alice@example.com \
  --impersonate admin@example.com
```

### Apply to Organizational Unit

```bash
./signatured apply \
  --org-unit /Engineering \
  --impersonate admin@example.com
```

### Apply to Entire Domain

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com
```

### Dry Run (Preview Changes)

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --dry-run
```

### Custom Template

Use a different template file:

```bash
./signatured apply \
  --user alice@example.com \
  --impersonate admin@example.com \
  --template ./templates/html-table.md
```

### Google Cloud Storage Templates

Load templates from Google Cloud Storage buckets:

```bash
# Using gs:// URL
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template gs://my-org-signatures/templates/signatured.md

# Using HTTPS URL
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template https://storage.googleapis.com/my-org-signatures/templates/signatured.md
```

**Requirements:**
- Service account needs `roles/storage.objectViewer` permission on the bucket
- Uses Application Default Credentials (same credentials as Workspace APIs)

**Grant bucket access:**
```bash
gsutil iam ch serviceAccount:signatured@project.iam.gserviceaccount.com:objectViewer \
  gs://my-org-signatures
```

### Custom Paths

```bash
./signatured apply \
  --user alice@example.com \
  --impersonate admin@example.com \
  --template ./custom-template.md \
  --credentials ./path/to/credentials.json
```

### Verbose Logging

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --verbose
```

### Concurrency Control

Adjust the number of concurrent API calls (default: 10):

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --concurrency 5
```

## Command Reference

### Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--template` | `./templates/signatured.md` | Path to signature template (local file or GCS URL: `gs://bucket/path` or `https://storage.googleapis.com/bucket/path`) |
| `--credentials` | `./credentials.json` | Path to service account key |
| `--impersonate` | *(required)* | Admin user for domain-wide delegation |
| `--verbose` | `false` | Enable debug logging |

### Environment Variables

Configuration via `.env` file (automatically loaded):

#### CLI Defaults
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `TEMPLATE_PATH` | Default template path (local or GCS) | `./templates/signatured.md` | `gs://my-bucket/templates/signatured.md` |
| `CREDENTIALS_PATH` | Path to service account credentials | `./credentials.json` | `/secrets/credentials.json` |
| `IMPERSONATE_USER` | Admin user for domain-wide delegation | *(none)* | `admin@example.com` |
| `VERBOSE` | Enable verbose logging | `false` | `true` |

#### Company-Wide Settings
| Variable | Description | Example |
|----------|-------------|---------|
| `COMPANY_WEBSITE` | Company website URL | `https://example.com` |
| `COMPANY_LOGO` | Company logo image URL | `https://example.com/logo.png` |
| `COMPANY_PHONE` | Company phone number | `+1-555-0100` |
| `COMPANY_ADDRESS` | Company address | `123 Main St, City, State` |

**Note**: Command-line flags take precedence over environment variables.

### Apply Command Flags

| Flag | Description |
|------|-------------|
| `--user` | Apply to single user by email |
| `--org-unit` | Apply to users in specific OU |
| `--all` | Apply to all users in domain |
| `--dry-run` | Preview without applying |
| `--concurrency` | Concurrent API calls (default: 10) |

**Note**: Must specify exactly one of `--user`, `--org-unit`, or `--all`.

## Example Output

```
time=2026-02-13T10:00:00.000+02:00 level=INFO msg="Loading template" path=./signatured.md
time=2026-02-13T10:00:00.001+02:00 level=INFO msg="Template loaded" size="324 bytes"
time=2026-02-13T10:00:00.002+02:00 level=INFO msg="Authenticating with Google Workspace" impersonate=admin@example.com
time=2026-02-13T10:00:00.500+02:00 level=INFO msg="Authentication successful"
time=2026-02-13T10:00:00.501+02:00 level=INFO msg="Fetching all users in domain"
time=2026-02-13T10:00:02.123+02:00 level=INFO msg="Users found" count=142
time=2026-02-13T10:00:02.124+02:00 level=INFO msg="Processing users"

✓ alice@example.com - signature updated
✓ bob@example.com - signature updated
✓ charlie@example.com - signature updated
...

Summary:
  Success: 140
  Failed: 1
  Skipped: 1
  Total: 142
  Duration: 23.4s
```

## Automation

### Scheduled Updates (cron)

Update signatures daily at 2 AM:

```bash
0 2 * * * /path/to/signatured apply --all --impersonate admin@example.com >> /var/log/signatured.log 2>&1
```

### Git Hook (Update on Template Changes)

Create `.git/hooks/post-commit`:

```bash
#!/bin/bash
if git diff --name-only HEAD~1 HEAD | grep -q "signatured.md"; then
    ./signatured apply --all --impersonate admin@example.com
fi
```

### Cloud Run (Serverless)

Deploy as a Cloud Run job triggered by Cloud Scheduler for automated updates.

## Troubleshooting

### Setup Issues

**Issue**: `oauth2: cannot fetch token`

**Cause**: Service account not configured for domain-wide delegation

**Solution**:
1. Complete domain-wide delegation setup (see Setup steps 5-6)
2. Ensure Client ID matches the service account
3. Verify OAuth scopes are exactly:
   ```
   https://www.googleapis.com/auth/admin.directory.user.readonly,https://www.googleapis.com/auth/gmail.settings.basic
   ```

---

**Issue**: `Not Authorized to access this resource/api`

**Cause**: Missing API enablement or incorrect scopes

**Solution**:
1. Verify Admin SDK API and Gmail API are enabled in Cloud Console
2. Check OAuth scopes in Admin Console match exactly (see above)
3. Wait 5-10 minutes after authorizing scopes for changes to propagate

---

**Issue**: `failed to read credentials file`

**Cause**: `credentials.json` not found

**Solution**:
1. Ensure file is in current directory
2. Or use `--credentials /path/to/credentials.json`
3. Verify file permissions: `chmod 600 credentials.json`

---

**Issue**: `impersonate user cannot be empty`

**Cause**: Missing `--impersonate` flag

**Solution**: Always provide `--impersonate admin@example.com` with an admin email address

### Runtime Errors

**Error**: `failed to create directory service: oauth2: cannot fetch token`

**Solution**: Verify domain-wide delegation is configured correctly with the right Client ID and scopes.

---

**Error**: `failed to list users: googleapi: Error 403: Not Authorized`

**Solution**: Ensure the service account has domain-wide delegation enabled and the correct OAuth scopes are authorized in Admin Console.

---

**Error**: `googleapi: Error 429: Too Many Requests`

**Solution**: The tool automatically retries with exponential backoff. Reduce `--concurrency` if issues persist.

### Template Issues

**Error**: `failed to read template file: no such file or directory`

**Solution**: Create `signatured.md` in the current directory or specify path with `--template`.

---

**Issue**: Users have incomplete signatures

**Cause**: Missing data in user profiles

**Solution**:
1. Check user profiles in Admin Console
2. Ensure fields (phone, job title, etc.) are populated
3. Use `{{#if field}}...{{/if}}` conditionals to gracefully hide missing fields (see Template Syntax section)

## Security

### Credentials and Configuration Storage

- Never commit `credentials.json` or `.env` to version control
- Store credentials in secure location with restricted permissions: `chmod 600 credentials.json`
- Restrict `.env` file permissions: `chmod 600 .env`
- For production, use secret management (GCP Secret Manager, Vault, etc.)

### Service Account Key Rotation

Rotate service account keys quarterly:

1. Create new key in Cloud Console
2. Test with new key
3. Delete old key

### Minimal Scopes

The tool uses minimal required scopes:
- `admin.directory.user.readonly` - Read user data only
- `gmail.settings.basic` - Update signatures only

### Audit Logging

All operations are logged with structured logs. Redirect to file for audit trail:

```bash
./signatured apply --all --impersonate admin@example.com 2>> audit.log
```

### Security Checklist

- [ ] `credentials.json` has restricted permissions (`chmod 600`)
- [ ] `credentials.json` is in `.gitignore`
- [ ] `.env` has restricted permissions (`chmod 600`)
- [ ] `.env` is in `.gitignore`
- [ ] Service account uses minimal required scopes
- [ ] Domain-wide delegation is limited to signature management scopes
- [ ] Service account keys rotated quarterly
- [ ] Audit logs reviewed regularly for unauthorized access

## Next Steps

After completing the initial setup:

1. **Customize Template**: Edit `signatured.md` to match your organization's branding
2. **Test with OU**: Apply to a test organizational unit first before rolling out organization-wide
3. **Schedule Updates**: Set up cron job or Cloud Scheduler for automated daily sync
4. **Monitor Logs**: Review logs regularly for any failed updates
5. **User Communication**: Inform users about automatic signature management policy

### Recommended Rollout Strategy

**Phase 1: Testing (Week 1)**
- Apply to IT/Admin team only
- Gather feedback on signature format
- Refine template based on feedback

**Phase 2: Pilot (Week 2)**
- Apply to one department or OU
- Monitor for issues and edge cases
- Adjust template as needed

**Phase 3: Full Rollout (Week 3+)**
- Apply to all users organization-wide
- Set up automated daily sync
- Document support process for signature requests

## Development

### Run Tests

```bash
go test -v ./...
```

### Build for All Platforms

```bash
# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o dist/signatured-darwin-arm64 ./cmd/signatured

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o dist/signatured-darwin-amd64 ./cmd/signatured

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o dist/signatured-linux-amd64 ./cmd/signatured

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o dist/signatured-windows-amd64.exe ./cmd/signatured
```

### Project Structure

```
signatured/
├── cmd/
│   └── signatured/          # Main CLI application
│       └── main.go
├── internal/
│   ├── google/              # Google API clients
│   │   ├── auth.go          # Authentication
│   │   ├── directory.go     # Directory API
│   │   └── gmail.go         # Gmail API
│   ├── models/              # Data models
│   │   └── user.go
│   └── template/            # Template engine
│       ├── template.go
│       └── template_test.go
├── templates/               # Signature templates
│   ├── signatured.md        # Default simple template
│   ├── html-table.md        # HTML table template
│   └── README.md            # Template documentation
├── .env                     # Environment config (gitignored)
├── .env.example             # Example environment config
├── credentials.json         # Service account key (gitignored)
├── go.mod
├── go.sum
└── README.md
```

## License

MIT License

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run `go test ./...` and `go vet ./...`
5. Submit pull request

## Support

For issues and questions:
- File an issue on GitHub
- Check troubleshooting section above
- Review Google Workspace API documentation
