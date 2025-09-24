#!/bin/bash

echo "Testing Docker build locally..."

# Test if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

echo "✅ Docker is available"

# Test Docker build
echo "Building Docker image..."
if docker build -t devhive-test .; then
    echo "✅ Docker build successful"
    
    # Test if the image runs
    echo "Testing if image runs..."
    if docker run --rm devhive-test echo "Hello from container"; then
        echo "✅ Docker image runs successfully"
    else
        echo "❌ Docker image failed to run"
        exit 1
    fi
else
    echo "❌ Docker build failed"
    exit 1
fi

echo "🎉 All Docker tests passed!"
