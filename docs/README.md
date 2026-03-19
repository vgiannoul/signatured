# signatured Documentation Site

This directory contains the GitHub Pages documentation site for signatured, built with the [Just the Docs](https://just-the-docs.github.io/just-the-docs/) theme.

## Structure

```
docs/
├── _config.yml          # Jekyll configuration (Just the Docs theme)
├── index.md             # Main documentation (synced from README.md)
├── GCS_SUPPORT.md       # Google Cloud Storage guide
├── ENV_CONFIG.md        # Environment variable configuration guide
├── assets/              # Static assets (logo, images)
│   └── logo.svg
└── README.md            # This file
```

## How It Works

The documentation is automatically synced from the main repository files:

- **README.md** → `docs/index.md` (with frontmatter added)
- **CHANGELOG.md** → `docs/changelog.md` (with frontmatter added)

This ensures a single source of truth - you only need to update the main files, and the documentation site will automatically reflect the changes.

## GitHub Actions Workflow

The [docs.yml](../.github/workflows/docs.yml) workflow:

1. **Triggers on changes to:**
   - `docs/**` - Documentation directory
   - `README.md` - Main documentation
   - `CHANGELOG.md` - Version history
   - `.github/workflows/docs.yml` - Workflow itself

2. **Syncs files:**
   - Copies `README.md` to `docs/index.md`
   - Copies `CHANGELOG.md` to `docs/changelog.md`
   - Adds proper Jekyll frontmatter to both files

3. **Builds and deploys:**
   - Uses Jekyll to build the site
   - Deploys to GitHub Pages

## Enabling GitHub Pages

To enable the documentation site on GitHub:

1. Go to repository settings: https://github.com/vgiannoul/signatured/settings/pages
2. Under "Build and deployment":
   - **Source**: Select "GitHub Actions"
3. The site will automatically deploy when changes are pushed to the `main` branch

## Accessing the Site

Once deployed, your documentation will be available at:
```
https://vgiannoul.github.io/signatured/
```

## Theme Features

The [Just the Docs](https://just-the-docs.github.io/just-the-docs/) theme provides:

- **Search functionality** - Full-text search across all pages
- **Navigation** - Automatic navigation sidebar with ordering
- **Responsive design** - Mobile-friendly layout
- **Syntax highlighting** - Beautiful code blocks with Rouge
- **Callouts** - Warning, note, and important callouts
- **Anchor links** - Direct linking to headings
- **Edit on GitHub** - Quick edit links for contributors

## Local Development

To preview the documentation site locally:

```bash
# Install dependencies
cd docs
bundle install

# Serve locally
bundle exec jekyll serve

# Visit http://localhost:4000/signatured/
```

### Gemfile

Create a `Gemfile` in the `docs/` directory:

```ruby
source "https://rubygems.org"

gem "jekyll", "~> 4.3"
gem "just-the-docs", "0.8.0"

group :jekyll_plugins do
  gem "jekyll-seo-tag", "~> 2.8"
  gem "jekyll-github-metadata", "~> 2.16"
  gem "jekyll-include-cache", "~> 0.2"
end
```

## Updating Documentation

### Main Documentation (README.md)

1. Edit the root [README.md](../README.md) file
2. Commit and push to `main` branch
3. GitHub Actions will automatically:
   - Add Jekyll frontmatter
   - Copy to `docs/index.md`
   - Deploy the updated site

### Changelog

1. Edit the root [CHANGELOG.md](../CHANGELOG.md) file
2. Commit and push to `main` branch
3. GitHub Actions will automatically:
   - Add Jekyll frontmatter
   - Copy to `docs/changelog.md`
   - Deploy the updated site

### Adding New Pages

To add new documentation pages:

1. Create a new `.md` file in the `docs/` directory
2. Add frontmatter:
   ```yaml
   ---
   layout: default
   title: Your Page Title
   nav_order: 5
   description: "Page description"
   permalink: /your-page/
   ---
   ```
3. Write your content using Markdown
4. Commit and push - it will automatically deploy

## Customization

### Theme Configuration

Edit `_config.yml` to customize:
- Color scheme (`color_scheme`)
- Search settings
- Navigation
- Footer content
- Callout styles
- Aux links (top-right navigation)

### Custom Styling

To add custom CSS:

1. Create `docs/assets/css/custom.css`
2. Add your custom styles
3. Reference in frontmatter or `_config.yml`

### Logo

The logo is stored at `docs/assets/logo.svg` and configured in `_config.yml`:

```yaml
logo: "/assets/logo.svg"
```

To update the logo, replace `docs/assets/logo.svg` with your new logo file.

## Navigation Order

Pages are ordered using the `nav_order` frontmatter:

- `1` - Home (index.md)
- `2` - Google Cloud Storage (GCS_SUPPORT.md)
- `3` - Environment Configuration (ENV_CONFIG.md)
- `99` - Changelog (changelog.md)

Add new pages with appropriate `nav_order` values to control their position in the sidebar.

## Support

For theme-specific questions, see:
- [Just the Docs Documentation](https://just-the-docs.github.io/just-the-docs/)
- [Just the Docs GitHub](https://github.com/just-the-docs/just-the-docs)

For signatured documentation issues:
- File an issue on [GitHub](https://github.com/vgiannoul/signatured/issues)
