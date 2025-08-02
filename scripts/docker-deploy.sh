#!/bin/bash

# Docker deployment scripts for go-chess

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
IMAGE_NAME="go-chess"
TAG="latest"
PORT="8080"

# Help function
show_help() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build       Build Docker image
    run         Run Docker container
    dev         Run in development mode with live reload
    stop        Stop running containers
    clean       Clean up containers and images
    logs        Show container logs
    shell       Open shell in running container
    help        Show this help

Options:
    -t, --tag TAG       Docker image tag (default: latest)
    -p, --port PORT     Host port to bind (default: 8080)
    -n, --name NAME     Container name (default: go-chess)
    --profile PROFILE   Docker compose profile (cli, server)

Examples:
    $0 build                    # Build image
    $0 run -p 9000             # Run on port 9000
    $0 dev                     # Development mode
    $0 clean                   # Clean everything

EOF
}

# Build Docker image
docker_build() {
    print_info "Building Docker image: ${IMAGE_NAME}:${TAG}"
    docker build -t "${IMAGE_NAME}:${TAG}" .
    print_info "Build completed successfully"
}

# Run Docker container
docker_run() {
    print_info "Running Docker container on port ${PORT}"
    docker run -d \
        --name "${IMAGE_NAME}" \
        -p "${PORT}:8080" \
        -e CHESS_HOST=0.0.0.0 \
        -e CHESS_PORT=8080 \
        -e CHESS_CORS_ENABLED=true \
        "${IMAGE_NAME}:${TAG}"
    print_info "Container started. Access at http://localhost:${PORT}"
    print_info "Health check: http://localhost:${PORT}/health"
}

# Development mode with live reload
docker_dev() {
    print_info "Starting development environment"
    docker-compose up --build chess-server
}

# Stop containers
docker_stop() {
    print_info "Stopping containers"
    docker-compose down
    docker stop "${IMAGE_NAME}" 2>/dev/null || true
    docker rm "${IMAGE_NAME}" 2>/dev/null || true
}

# Clean up
docker_clean() {
    print_warning "This will remove all go-chess containers and images"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker_stop
        docker rmi "${IMAGE_NAME}:${TAG}" 2>/dev/null || true
        docker system prune -f
        print_info "Cleanup completed"
    fi
}

# Show logs
docker_logs() {
    docker logs -f "${IMAGE_NAME}"
}

# Open shell
docker_shell() {
    docker exec -it "${IMAGE_NAME}" /bin/sh
}

# Parse command line arguments
COMMAND=""
while [[ $# -gt 0 ]]; do
    case $1 in
        build|run|dev|stop|clean|logs|shell|help)
            COMMAND="$1"
            shift
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --profile)
            PROFILE="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Execute command
case "${COMMAND}" in
    build)
        docker_build
        ;;
    run)
        docker_build
        docker_run
        ;;
    dev)
        docker_dev
        ;;
    stop)
        docker_stop
        ;;
    clean)
        docker_clean
        ;;
    logs)
        docker_logs
        ;;
    shell)
        docker_shell
        ;;
    help|"")
        show_help
        ;;
    *)
        print_error "Unknown command: ${COMMAND}"
        show_help
        exit 1
        ;;
esac
