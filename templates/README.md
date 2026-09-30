# Email Signature Templates

This directory contains pre-built email signature templates for use with signatured.

## Available Templates

### example.html

Example HTML table layout with a logo and contact block, branded for a fictional company
("Signatured Co"). It's a `.html` file, so it's used as literal HTML (no Markdown parsing) —
see "Template Formats" below. Copy it and replace the branding with your own.

**Features:**
- Two-column layout with logo and user information
- Company contact information (email, phone, address, website)
- Mobile phone support (conditional)

**Usage:**

First, configure company settings in your `.env` file:

```bash
COMPANY_WEBSITE=https://example.com
COMPANY_LOGO=https://example.com/logo.png
COMPANY_PHONE=+1-555-0100
COMPANY_ADDRESS=123 Main St, City, State 12345
```

Preview it locally before applying anything:

```bash
./signatured preview --sample --template ./templates/example.html
```

Then apply the template:

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template ./templates/example.html
```

**Required Placeholders:**
- `{{firstName}}`, `{{lastName}}` - User's name
- `{{jobTitle}}` - User's job title
- `{{email}}` - User's email
- `{{companyWebsite}}` - Company website (from `.env`)
- `{{companyLogo}}` - Company logo URL (from `.env`)
- `{{companyPhone}}` - Main company phone (from `.env`)
- `{{companyAddress}}` - Company address (from `.env`)

**Optional Placeholders:**
- `{{phoneMobile}}` - User's mobile phone (conditionally displayed)

## Creating Custom Templates

Templates support:
- Markdown formatting, converted to HTML (`.md` files)
- Literal HTML, used as-is with no Markdown parsing (`.html`/`.htm` files) — use this for
  precise, complex layouts where Markdown's HTML-block handling would be fragile
- Handlebars-style placeholders: `{{fieldName}}` (always HTML-escaped before substitution)
- Conditional blocks: `{{#if fieldName}}content{{/if}}`

Use `./signatured preview --sample --template <path>` to see the rendered output locally
before running `validate`, `--dry-run`, or `apply` against real users.

### Example Custom Template

```markdown
<div style="font-family: Arial, sans-serif;">
  <strong>{{firstName}} {{lastName}}</strong><br>
  {{#if jobTitle}}<em>{{jobTitle}}</em><br>{{/if}}
  <a href="mailto:{{email}}">{{email}}</a>
</div>
```

Save as `custom.md` and use with:

```bash
./signatured apply --template ./custom.md --all --impersonate admin@example.com
```

## Placeholder Reference

### User-Specific Fields

| Placeholder | Source | Always Available? |
|-------------|--------|-------------------|
| `{{email}}` | User's primary email | Yes |
| `{{firstName}}` | User's given name | Yes |
| `{{lastName}}` | User's family name | Yes |
| `{{phone}}` | User's work phone (falls back to any non-mobile entry) | No (use `{{#if phone}}`) |
| `{{phoneLabel}}` | Label for `{{phone}}` - normal or internal, from settings | No (use `{{#if phoneLabel}}`) |
| `{{phoneIsInternal}}` | Truthy only when the work phone is an internal extension | No (use `{{#if phoneIsInternal}}`) |
| `{{phoneMobile}}` | User's mobile phone | No (use `{{#if phoneMobile}}`) |
| `{{jobTitle}}` | User's job title | No (use `{{#if jobTitle}}`) |
| `{{organization}}` | User's organization | No (use `{{#if organization}}`) |
| `{{orgUnit}}` | Organizational unit path | No (use `{{#if orgUnit}}`) |

### Company-Wide Fields

Set via `.env` file (same for all users):

| Placeholder | Environment Variable | Description |
|-------------|---------------------|-------------|
| `{{companyWebsite}}` | `COMPANY_WEBSITE` | Company website URL |
| `{{companyLogo}}` | `COMPANY_LOGO` | Company logo image URL |
| `{{companyPhone}}` | `COMPANY_PHONE` | Main company phone |
| `{{companyAddress}}` | `COMPANY_ADDRESS` | Company address |

## Best Practices

1. **Preview before testing against real users**:
   ```bash
   ./signatured preview --sample --template ./templates/example.html
   ```

2. **Always use conditionals for optional fields** to avoid blank spaces:
   ```markdown
   {{#if phone}}Phone: {{phone}}{{/if}}
   ```

3. **Test with a single user first**:
   ```bash
   ./signatured apply --user test@example.com --template ./templates/example.html --dry-run
   ```

4. **Validate before applying**:
   ```bash
   ./signatured validate --template ./templates/example.html
   ```

5. **Keep HTML simple** - Email clients have limited HTML support. Avoid:
   - External CSS files
   - JavaScript
   - Complex positioning (flexbox, grid)
   - Background images

6. **Use inline styles** - All CSS should be inline style attributes

7. **Test in multiple email clients** - Gmail, Outlook, Apple Mail all render differently
