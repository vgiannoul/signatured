# signatured Documentation Site

This directory contains the GitHub Pages documentation site for signatured.

## Structure

```
docs/
├── _config.yml          # Jekyll configuration
├── index.md             # Main documentation (from README.md)
├── assets/              # Static assets (logo, images)
└── README.md            # This file
```

## Enabling GitHub Pages

To enable the documentation site on GitHub:

1. Go to your repository settings: https://github.com/vgiannoul/signatured/settings/pages
2. Under "Build and deployment":
   - Source: Select "GitHub Actions"
3. The site will automatically deploy when changes are pushed to the `main` branch

## Local Development

To preview the documentation site locally:

```bash
# Install Jekyll
gem install bundler jekyll

# Create Gemfile in docs/
cat > docs/Gemfile <<EOF
source "https://rubygems.org"
gem "github-pages", group: :jekyll_plugins
gem "jekyll-theme-cayman"
EOF

# Install dependencies
cd docs
bundle install

# Serve locally
bundle exec jekyll serve

# Visit http://localhost:4000
```

## Updating Documentation

The documentation is automatically synced from README.md. To update:

1. Edit the main [README.md](../README.md) file
2. Copy changes to [index.md](index.md):
   ```bash
   cp README.md docs/index.md
   ```
3. Commit and push to trigger deployment

## Theme

The site uses the [Cayman theme](https://github.com/pages-themes/cayman) which provides:
- Clean, modern design
- Responsive layout
- Syntax highlighting for code blocks
- GitHub integration

## Customization

To customize the theme, you can:
- Modify `_config.yml` for site settings
- Add custom CSS in `assets/css/style.scss`
- Override layouts by creating `_layouts/` directory
