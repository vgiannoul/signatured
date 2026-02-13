---
layout: default
title: Home
nav_order: 1
description: "Google Workspace Signature Manager - Single-binary CLI tool to manage email signatures for Google Workspace organization members"
permalink: /
---

<div align="center">
  <img src="assets/logo.svg" alt="signatured logo" width="400">

  # Google Workspace Signature Manager

  Single-binary CLI tool to manage email signatures for Google Workspace organization members.
</div>

## Features

- Single static binary (no runtime dependencies)
- Markdown-based signature templates with placeholder support
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

### 7. Create Signature Template

Create a `signatured.md` file in your project directory:

```markdown
**{% raw %}{{firstName}}{% endraw %} {% raw %}{{lastName}}{% endraw %}**
{% raw %}{{#if jobTitle}}{% endraw %}{% raw %}{{jobTitle}}{% endraw %}
{% raw %}{{/if}}{% endraw %}{% raw %}{{#if organization}}{% endraw %}{% raw %}{{organization}}{% endraw %}

{% raw %}{{/if}}{% endraw %}---

{% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}
{% raw %}{{/if}}{% endraw %}✉️ {% raw %}{{email}}{% endraw %}
{% raw %}{{#if orgUnit}}{% endraw %}🏢 {% raw %}{{orgUnit}}{% endraw %}{% raw %}{{/if}}{% endraw %}
```

**Note**: The `{% raw %}{{#if field}}{% endraw %}...{% raw %}{{/if}}{% endraw %}` syntax automatically hides sections when a user's profile is missing that field, preventing awkward blank lines in signatures.

### 8. Test Your Configuration

Validate the template:

```bash
./signatured validate
```

Expected output:
```
time=2026-02-13T10:00:00.000+02:00 level=INFO msg="Validating template" path=./signatured.md
time=2026-02-13T10:00:00.001+02:00 level=INFO msg="Template is valid"
```

### 9. Run Dry Run Test

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

### 10. Apply to Single User

Remove `--dry-run` to actually apply:

```bash
./signatured apply \
  --user testuser@example.com \
  --impersonate admin@example.com
```

### 11. Verify Signature Update

1. Open Gmail as `testuser@example.com`
2. Click **Settings** (gear icon)
3. Scroll to **Signature** section
4. Verify signature is updated correctly

### 12. Roll Out to Organization

Once verified, apply to all users:

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com
```

### Setup Checklist

After completing setup, verify you have:

```
signatured/                # Project directory
├── signatured             # Binary (executable)
├── credentials.json       # Service account key (600 permissions)
├── signatured.md          # Signature template
└── .gitignore             # Excludes credentials.json
```

### Template Syntax

#### Placeholders

| Placeholder | Description | Source Field |
|-------------|-------------|--------------|
| `{% raw %}{{firstName}}{% endraw %}` | First name | `name.givenName` |
| `{% raw %}{{lastName}}{% endraw %}` | Last name | `name.familyName` |
| `{% raw %}{{email}}{% endraw %}` | Email address | `primaryEmail` |
| `{% raw %}{{phone}}{% endraw %}` | Phone number (prefers work) | `phones[0].value` |
| `{% raw %}{{orgUnit}}{% endraw %}` | Organizational unit path | `orgUnitPath` |
| `{% raw %}{{jobTitle}}{% endraw %}` | Job title | `organizations[0].title` |
| `{% raw %}{{organization}}{% endraw %}` | Organization name | `organizations[0].name` |

#### Conditional Blocks

The signature manager supports conditional blocks that automatically hide sections when user data is missing, preventing awkward blank lines in signatures.

**Syntax:**
```markdown
{% raw %}{{#if fieldName}}{% endraw %}content{% raw %}{{/if}}{% endraw %}
```

- If `fieldName` has a value → renders `content`
- If `fieldName` is empty → removes entire block

**Example 1: Phone Number**

Template:
```markdown
{% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}
{% raw %}{{/if}}{% endraw %}✉️ {% raw %}{{email}}{% endraw %}
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
**{% raw %}{{firstName}}{% endraw %} {% raw %}{{lastName}}{% endraw %}**
{% raw %}{{#if jobTitle}}{% endraw %}{% raw %}{{jobTitle}}{% endraw %}
{% raw %}{{/if}}{% endraw %}{% raw %}{{#if organization}}{% endraw %}{% raw %}{{organization}}{% endraw %}

{% raw %}{{/if}}{% endraw %}---

{% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}
{% raw %}{{/if}}{% endraw %}✉️ {% raw %}{{email}}{% endraw %}
{% raw %}{{#if orgUnit}}{% endraw %}🏢 {% raw %}{{orgUnit}}{% endraw %}{% raw %}{{/if}}{% endraw %}
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
   {% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}
   {% raw %}{{/if}}{% endraw %}
   ```

   ❌ Bad:
   ```markdown
   📞 {% raw %}{{phone}}{% endraw %}
   ```
   *This leaves "📞 " in the signature even when phone is missing*

2. **Include newlines inside conditionals** - If you want the newline to disappear when the field is missing:

   ✅ Good:
   ```markdown
   {% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}
   {% raw %}{{/if}}{% endraw %}✉️ {% raw %}{{email}}{% endraw %}
   ```

   ❌ Bad:
   ```markdown
   {% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}{% raw %}{{/if}}{% endraw %}

   ✉️ {% raw %}{{email}}{% endraw %}
   ```
   *This leaves a blank line when phone is missing*

3. **Required fields don't need conditionals** - Fields like `firstName`, `lastName`, and `email` are typically always present:

   ```markdown
   **{% raw %}{{firstName}}{% endraw %} {% raw %}{{lastName}}{% endraw %}**
   ✉️ {% raw %}{{email}}{% endraw %}
   ```

4. **Conditionals work with markdown** - You can use markdown formatting inside conditional blocks:

   ```markdown
   {% raw %}{{#if jobTitle}}{% endraw %}**Job:** *{% raw %}{{jobTitle}}{% endraw %}*
   {% raw %}{{/if}}{% endraw %}
   ```

**Common Patterns:**

Contact information block:
```markdown
{% raw %}{{#if phone}}{% endraw %}📞 {% raw %}{{phone}}{% endraw %}
{% raw %}{{/if}}{% endraw %}✉️ {% raw %}{{email}}{% endraw %}
{% raw %}{{#if website}}{% endraw %}🌐 {% raw %}{{website}}{% endraw %}{% raw %}{{/if}}{% endraw %}
```

Job title and company:
```markdown
{% raw %}{{#if jobTitle}}{% endraw %}{% raw %}{{jobTitle}}{% endraw %}{% raw %}{{#if organization}}{% endraw %} at {% raw %}{{organization}}{% endraw %}{% raw %}{{/if}}{% endraw %}
{% raw %}{{/if}}{% endraw %}
```

This prevents awkward blank lines in signatures for users with incomplete profile data.

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

### Custom Paths

```bash
./signatured apply \
  --user alice@example.com \
  --impersonate admin@example.com \
  --template ./custom-signatured.md \
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
| `--template` | `./signatured.md` | Path to signature template |
| `--credentials` | `./credentials.json` | Path to service account key |
| `--impersonate` | *(required)* | Admin user for domain-wide delegation |
| `--verbose` | `false` | Enable debug logging |

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
3. Use `{% raw %}{{#if field}}{% endraw %}...{% raw %}{{/if}}{% endraw %}` conditionals to gracefully hide missing fields (see Template Syntax section)

## Security

### Credentials Storage

- Never commit `credentials.json` to version control
- Store in secure location with restricted permissions: `chmod 600 credentials.json`
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
│   └── signatured/    # Main CLI application
│       └── main.go
├── internal/
│   ├── google/               # Google API clients
│   │   ├── auth.go          # Authentication
│   │   ├── directory.go     # Directory API
│   │   └── gmail.go         # Gmail API
│   ├── models/              # Data models
│   │   └── user.go
│   └── template/            # Template engine
│       ├── template.go
│       └── template_test.go
├── signatured.md             # Signature template
├── credentials.json         # Service account key (gitignored)
├── go.mod
├── go.sum
├── README.md
└── PLAN.md                  # Design documentation
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
