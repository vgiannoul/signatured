# Email Signature Templates

This directory contains pre-built email signature templates for use with signatured.

## Available Templates

### html-table.md

Professional HTML table layout with company branding elements.

**Features:**
- Two-column layout with logo and user information
- Social media icons
- Company contact information
- Mobile phone support (conditional)
- Recruitment banner at the bottom

**Usage:**

First, configure company settings in your `.env` file:

```bash
COMPANY_WEBSITE=https://example.com
COMPANY_LOGO=https://example.com/logo.png
COMPANY_PHONE=+1-555-0100
COMPANY_ADDRESS=123 Main St, City, State 12345
```

Then apply the template:

```bash
./signatured apply \
  --all \
  --impersonate admin@example.com \
  --template ./templates/html-table.md
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
- Markdown formatting (converted to HTML)
- Raw HTML (for complex layouts)
- Handlebars-style placeholders: `{{fieldName}}`
- Conditional blocks: `{{#if fieldName}}content{{/if}}`

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
| `{{phone}}` | User's work phone | No (use `{{#if phone}}`) |
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

1. **Always use conditionals for optional fields** to avoid blank spaces:
   ```markdown
   {{#if phone}}Phone: {{phone}}{{/if}}
   ```

2. **Test with a single user first**:
   ```bash
   ./signatured apply --user test@example.com --template ./templates/html-table.md --dry-run
   ```

3. **Validate before applying**:
   ```bash
   ./signatured validate --template ./templates/html-table.md
   ```

4. **Keep HTML simple** - Email clients have limited HTML support. Avoid:
   - External CSS files
   - JavaScript
   - Complex positioning (flexbox, grid)
   - Background images

5. **Use inline styles** - All CSS should be inline style attributes

6. **Test in multiple email clients** - Gmail, Outlook, Apple Mail all render differently
