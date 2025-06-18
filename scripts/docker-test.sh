#!/bin/bash

# Docker test script for PullPoet
set -e

echo "🐳 Testing PullPoet Docker image..."

# Build the image
echo "📦 Building Docker image..."
docker build -t pullpoet .

# Test basic functionality
echo "🧪 Testing basic functionality..."
docker run --rm pullpoet --help

echo "✅ Docker image test completed successfully!"
echo ""
echo "Usage examples:"
echo "  docker run --rm pullpoet --help"
echo "  docker run --rm pullpoet --repo https://github.com/example/repo.git --source feature --target main --provider openai --model gpt-3.5-turbo --api-key YOUR_KEY"
