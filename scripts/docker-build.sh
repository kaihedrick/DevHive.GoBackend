#!/bin/bash

# DevHive Backend Docker Build Script
# This script helps build, run, and manage Docker containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="dev"
IMAGE_NAME="devhive-backend"
TAG="latest"
COMPOSE_FILE="docker-compose.yml"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  build       Build the Docker image"
    echo "  run         Run the application (dev mode)"
    echo "  run-prod    Run the application (production mode)"
    echo "  stop        Stop all containers"
    echo "  clean       Clean up containers and images"
    echo "  logs        Show container logs"
    echo "  shell       Access container shell"
    echo "  test        Run health checks"
    echo ""
    echo "Options:"
    echo "  -t, --tag TAG        Docker image tag (default: latest)"
    echo "  -e, --env ENV        Environment (dev/prod, default: dev)"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build                    # Build image with latest tag"
    echo "  $0 run                      # Run in development mode"
    echo "  $0 run-prod                 # Run in production mode"
    echo "  $0 stop                     # Stop all containers"
    echo "  $0 test                     # Test all endpoints"
}

# Function to build Docker image
build_image() {
    print_status "Building Docker image: $IMAGE_NAME:$TAG"
    
    # Check if .env file exists
    if [ ! -f .env ]; then
        print_warning ".env file not found. Creating from env.example..."
        if [ -f env.example ]; then
            cp env.example .env
            print_warning "Please update .env file with your configuration"
        else
            print_error "env.example file not found. Please create .env file manually."
            exit 1
        fi
    fi
    
    docker build -t $IMAGE_NAME:$TAG .
    print_success "Docker image built successfully: $IMAGE_NAME:$TAG"
}

# Function to run development environment
run_dev() {
    print_status "Starting development environment..."
    
    if [ ! -f .env ]; then
        print_error ".env file not found. Please create it first."
        exit 1
    fi
    
    docker-compose -f $COMPOSE_FILE up -d
    print_success "Development environment started"
    print_status "Application available at: http://localhost:8080"
    print_status "Swagger docs at: http://localhost:8080/swagger/"
    print_status "Health check at: http://localhost:8080/health"
}

# Function to run production environment
run_prod() {
    print_status "Starting production environment..."
    
    if [ ! -f .env ]; then
        print_error ".env file not found. Please create it first."
        exit 1
    fi
    
    COMPOSE_FILE="docker-compose.prod.yml"
    docker-compose -f $COMPOSE_FILE up -d
    print_success "Production environment started"
    print_status "Application available at: http://localhost:8080"
    print_status "Health check at: http://localhost:8080/health"
}

# Function to stop containers
stop_containers() {
    print_status "Stopping all containers..."
    docker-compose -f $COMPOSE_FILE down
    docker-compose -f docker-compose.prod.yml down 2>/dev/null || true
    print_success "All containers stopped"
}

# Function to clean up
clean_up() {
    print_status "Cleaning up containers and images..."
    docker-compose -f $COMPOSE_FILE down -v --rmi all
    docker-compose -f docker-compose.prod.yml down -v --rmi all 2>/dev/null || true
    docker system prune -f
    print_success "Cleanup completed"
}

# Function to show logs
show_logs() {
    print_status "Showing container logs..."
    docker-compose -f $COMPOSE_FILE logs -f
}

# Function to access container shell
access_shell() {
    print_status "Accessing container shell..."
    docker exec -it devhive-backend sh
}

# Function to test endpoints
test_endpoints() {
    print_status "Testing all endpoints..."
    
    # Wait for service to be ready
    print_status "Waiting for service to be ready..."
    sleep 10
    
    # Test health endpoint
    print_status "Testing health endpoint..."
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Health endpoint: OK"
    else
        print_error "Health endpoint: FAILED"
        return 1
    fi
    
    # Test Swagger endpoint
    print_status "Testing Swagger endpoint..."
    if curl -f http://localhost:8080/swagger/ > /dev/null 2>&1; then
        print_success "Swagger endpoint: OK"
    else
        print_warning "Swagger endpoint: FAILED (may not be critical)"
    fi
    
    print_success "All endpoint tests completed"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        build)
            COMMAND="build"
            shift
            ;;
        run)
            COMMAND="run"
            shift
            ;;
        run-prod)
            COMMAND="run-prod"
            shift
            ;;
        stop)
            COMMAND="stop"
            shift
            ;;
        clean)
            COMMAND="clean"
            shift
            ;;
        logs)
            COMMAND="logs"
            shift
            ;;
        shell)
            COMMAND="shell"
            shift
            ;;
        test)
            COMMAND="test"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Execute command
case $COMMAND in
    build)
        build_image
        ;;
    run)
        run_dev
        ;;
    run-prod)
        run_prod
        ;;
    stop)
        stop_containers
        ;;
    clean)
        clean_up
        ;;
    logs)
        show_logs
        ;;
    shell)
        access_shell
        ;;
    test)
        test_endpoints
        ;;
    *)
        print_error "No command specified"
        show_usage
        exit 1
        ;;
esac
