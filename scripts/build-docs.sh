#!/bin/bash

# Build and serve MkDocs documentation locally

set -e

echo "Building ProfileToMetrics Connector Documentation..."

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is required but not installed."
    exit 1
fi

# Check if pip is installed
if ! command -v pip3 &> /dev/null; then
    echo "Error: pip3 is required but not installed."
    exit 1
fi

# Install dependencies if requirements.txt exists
if [ -f "requirements.txt" ]; then
    echo "Installing Python dependencies..."
    pip3 install -r requirements.txt
fi

# Check if mkdocs is installed
if ! command -v mkdocs &> /dev/null; then
    echo "Error: MkDocs is not installed. Please install it with: pip3 install mkdocs mkdocs-material"
    exit 1
fi

# Build documentation
echo "Building documentation..."
mkdocs build --strict

echo "Documentation built successfully!"
echo "You can now:"
echo "1. Serve locally: mkdocs serve"
echo "2. View the site directory: open site/index.html"
echo "3. Deploy to GitHub Pages: mkdocs gh-deploy"
