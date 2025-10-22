# Documentation

This directory contains the MkDocs documentation for the ProfileToMetrics Connector.

## Structure

- `index.md` - Homepage
- `getting-started/` - Getting started guides
- `configuration/` - Configuration documentation
- `deployment/` - Deployment guides
- `testing/` - Testing documentation
- `development/` - Development guides
- `api/` - API reference

## Building Locally

### Prerequisites

- Python 3.11+
- pip

### Installation

```bash
# Install dependencies
pip install -r requirements.txt

# Or install manually
pip install mkdocs mkdocs-material pymdownx-extensions
```

### Build and Serve

```bash
# Build documentation
mkdocs build

# Serve locally (with live reload)
mkdocs serve

# Deploy to GitHub Pages
mkdocs gh-deploy
```

### Using the build script

```bash
# Build documentation
./scripts/build-docs.sh

# Serve locally
mkdocs serve
```

## Configuration

The documentation is configured in `mkdocs.yml` with:

- **Material Theme**: Modern, responsive design
- **Navigation**: Tabbed navigation with sections
- **Search**: Built-in search functionality
- **Code Highlighting**: Syntax highlighting for code blocks
- **Mermaid Diagrams**: Support for Mermaid diagrams
- **Emoji Support**: Emoji support in markdown

## Features

- **Responsive Design**: Works on desktop and mobile
- **Dark Mode**: Automatic dark/light mode switching
- **Search**: Full-text search across all pages
- **Code Copy**: Copy code blocks with one click
- **Table of Contents**: Automatic TOC generation
- **Breadcrumbs**: Navigation breadcrumbs
- **Social Links**: GitHub and Twitter links

## Contributing

When adding new documentation:

1. Create the appropriate file in the correct directory
2. Update `mkdocs.yml` navigation if needed
3. Test locally with `mkdocs serve`
4. Submit a pull request

## Deployment

The documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch.

- **Live Site**: https://henrikrexed.github.io/profiletoMetrics
- **Source**: This repository
- **Build**: GitHub Actions workflow
