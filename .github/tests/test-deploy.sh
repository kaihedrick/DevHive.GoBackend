#!/bin/bash
# Bash script to test GitHub Actions deploy workflow locally using Act
# Usage: ./test-deploy.sh [--workflow WORKFLOW] [--dryrun] [--job JOB]

set -e

WORKFLOW="deploy.yml"
DRYRUN=false
JOB=""
BUILD_ONLY=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --workflow)
            WORKFLOW="$2"
            shift 2
            ;;
        --dryrun)
            DRYRUN=true
            shift
            ;;
        --job)
            JOB="$2"
            shift 2
            ;;
        --build-only)
            BUILD_ONLY=true
            shift
            ;;
        --help|-h)
            echo "GitHub Actions Local Testing Script"
            echo ""
            echo "Usage:"
            echo "    ./test-deploy.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "    --workflow WORKFLOW    Workflow file to test (default: deploy.yml)"
            echo "    --dryrun              Show what would run without executing"
            echo "    --job JOB             Run specific job only"
            echo "    --build-only          Test build-only workflow instead"
            echo "    --help, -h            Show this help message"
            echo ""
            echo "Examples:"
            echo "    ./test-deploy.sh"
            echo "    ./test-deploy.sh --workflow deploy.yml --dryrun"
            echo "    ./test-deploy.sh --job build-and-deploy"
            echo "    ./test-deploy.sh --build-only"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if Act is installed
if ! command -v act &> /dev/null; then
    echo "❌ Act is not installed. Please install it first:"
    echo "   brew install act"
    echo "   Or see: https://github.com/nektos/act#installation"
    exit 1
fi

# Check if Docker is running
if ! docker ps &> /dev/null; then
    echo "❌ Docker is not running. Please start Docker."
    exit 1
fi

echo "🚀 Starting GitHub Actions local test..."
echo ""

# Set working directory to tests folder
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Determine workflow file
if [ "$BUILD_ONLY" = true ]; then
    WORKFLOW_FILE="test-build-only.yml"
    echo "📋 Testing build-only workflow..."
else
    WORKFLOW_FILE="../workflows/$WORKFLOW"
    echo "📋 Testing workflow: $WORKFLOW"
fi

# Check if workflow file exists
if [ ! -f "$WORKFLOW_FILE" ]; then
    echo "❌ Workflow file not found: $WORKFLOW_FILE"
    exit 1
fi

# Build Act command
ACT_CMD="act -W $WORKFLOW_FILE"

# Add secrets file if it exists
if [ -f ".secrets" ]; then
    ACT_CMD="$ACT_CMD --secret-file .secrets"
    echo "🔐 Using secrets from .secrets file"
else
    echo "⚠️  No .secrets file found. Using environment variables only."
    echo "   Copy .secrets.example to .secrets and add your test secrets."
fi

# Add dry run flag
if [ "$DRYRUN" = true ]; then
    ACT_CMD="$ACT_CMD --dryrun"
    echo "🔍 Dry run mode - no actions will be executed"
fi

# Add job filter
if [ -n "$JOB" ]; then
    ACT_CMD="$ACT_CMD -j $JOB"
    echo "🎯 Running specific job: $JOB"
fi

echo ""
echo "Command: $ACT_CMD"
echo ""

# Execute Act
eval $ACT_CMD

EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo ""
    echo "✅ Test completed successfully!"
else
    echo ""
    echo "❌ Test failed with exit code: $EXIT_CODE"
fi

exit $EXIT_CODE







