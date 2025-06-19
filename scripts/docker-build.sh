#!/bin/bash

# Docker Build Script for PullPoet
# Automatically extracts version from git tags and builds Docker image

set -e

# Get version from git tags
VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 'dev')
IMAGE_NAME=${1:-pullpoet}

echo "🏗️  Building Docker image..."
echo "📦 Image name: $IMAGE_NAME"
echo "🏷️  Version: $VERSION"
echo

# Build Docker image with version
docker build --build-arg VERSION=$VERSION -t $IMAGE_NAME .

echo
echo "✅ Build completed successfully!"
echo "🧪 Test version: docker run --rm $IMAGE_NAME -v"
echo "🚀 Run: docker run --rm $IMAGE_NAME [options]"
