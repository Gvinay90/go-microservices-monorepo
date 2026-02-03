#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function print_usage {
    echo -e "${BLUE}Usage: ./run.sh [COMMAND]${NC}"
    echo ""
    echo "Commands:"
    echo "  build          - Build all services with Bazel"
    echo "  build-docker   - Build Docker images using Bazel"
    echo "  run            - Run services with Docker Compose"
    echo "  watch          - Run services with Docker Compose Watch (Hot Reload)"
    echo "  run-local      - Run services locally with Bazel (no Docker)"
    echo "  stop           - Stop all services"
    echo "  clean          - Clean Bazel cache and Docker containers"
    echo "  logs           - View service logs"
    echo "  test           - Run tests"
    echo "  health         - Check service health"
    echo ""
    echo "Examples:"
    echo "  ./run.sh build         # Build with Bazel"
    echo "  ./run.sh run           # Run in Docker"
    echo "  ./run.sh watch         # Run with Hot Reload"
    echo "  ./run.sh run-local     # Run locally without Docker"
}

# Check if bazel is available
if ! command -v bazel &> /dev/null; then
    echo -e "${RED}Error: bazel is not installed${NC}"
    echo "Install from: https://bazel.build/install"
    exit 1
fi

# Check for docker compose availability (only for docker commands)
function check_docker {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Error: docker is not installed${NC}"
        exit 1
    fi

    if command -v docker-compose &> /dev/null; then
        DOCKER_COMPOSE="docker-compose"
    elif docker compose version &> /dev/null; then
        DOCKER_COMPOSE="docker compose"
    else
        echo -e "${RED}Error: docker-compose (or 'docker compose') is not installed${NC}"
        exit 1
    fi
}

COMMAND=$1

case $COMMAND in
    build)
        echo -e "${GREEN}Building with Bazel...${NC}"
        # Skip api-gateway due to quic-go/chacha20 dependency issues
        bazel build //proto/... //services/... //internal/... //pkg/...
        echo -e "${GREEN}✓ Build complete!${NC}"
        ;;
    
    build-docker)
        check_docker
        echo -e "${GREEN}Building Docker images (Bazel builds inside container)...${NC}"
        $DOCKER_COMPOSE build
        echo -e "${GREEN}✓ Docker images built!${NC}"
        ;;
    
    run)
        check_docker
        echo -e "${GREEN}Starting all services with Docker Compose...${NC}"
        $DOCKER_COMPOSE up -d
        
        echo -e "${YELLOW}Waiting for services to be ready...${NC}"
        sleep 10
        
        echo -e "${GREEN}✓ All services are running!${NC}"
        echo ""
        echo -e "${BLUE}Services:${NC}"
        echo -e "  ${YELLOW}PostgreSQL:${NC} localhost:5433"
        echo -e "  ${YELLOW}User Service:${NC} localhost:50051 (gRPC)"
        echo -e "  ${YELLOW}Expense Service:${NC} localhost:50052 (gRPC)"
        echo -e "  ${YELLOW}API Gateway:${NC} http://localhost:8081"
        echo -e "  ${YELLOW}Web Frontend:${NC} http://localhost:3000"
        echo -e "  ${YELLOW}Keycloak:${NC} http://localhost:8080 (admin/admin)"
        echo -e "  ${YELLOW}Mailhog:${NC}  http://localhost:8025"
        echo ""
        echo -e "${GREEN}🚀 Open your browser:${NC} ${BLUE}http://localhost:3000${NC}"
        echo ""
        echo -e "Run ${YELLOW}./run.sh logs${NC} to view logs"
        echo -e "Run ${YELLOW}./run.sh stop${NC} to stop services"
        ;;

    watch)
        check_docker
        echo -e "${GREEN}Starting services with Docker Compose Watch...${NC}"
        echo -e "${YELLOW}This enables hot-reloading for Go backend and React frontend${NC}"
        $DOCKER_COMPOSE watch
        ;;
    
    run-local)
        echo -e "${GREEN}Starting services locally with Bazel...${NC}"
        echo -e "${YELLOW}Starting User Service on :50051...${NC}"
        bazel run //services/user:user_service &
        USER_PID=$!
        
        echo -e "${YELLOW}Starting Expense Service on :50052...${NC}"
        APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service &
        EXPENSE_PID=$!
        
        echo -e "${GREEN}Services started!${NC}"
        echo -e "  ${BLUE}User Service PID:${NC} $USER_PID"
        echo -e "  ${BLUE}Expense Service PID:${NC} $EXPENSE_PID"
        echo ""
        echo -e "${YELLOW}Press Ctrl+C to stop services${NC}"
        
        # Wait for Ctrl+C
        trap "echo -e '\n${YELLOW}Stopping services...${NC}'; kill $USER_PID $EXPENSE_PID 2>/dev/null; exit 0" INT
        wait
        ;;
    
    stop)
        check_docker
        echo -e "${GREEN}Stopping services...${NC}"
        $DOCKER_COMPOSE down
        echo -e "${GREEN}✓ Services stopped${NC}"
        ;;
    
    clean)
        check_docker
        echo -e "${GREEN}Cleaning Bazel cache...${NC}"
        bazel clean --expunge
        echo -e "${GREEN}Cleaning Docker...${NC}"
        $DOCKER_COMPOSE down --rmi all --volumes
        echo -e "${GREEN}✓ Clean complete!${NC}"
        ;;
    
    logs)
        check_docker
        $DOCKER_COMPOSE logs -f
        ;;
    
    test)
        echo -e "${GREEN}Running tests...${NC}"
        bazel test //...
        ;;
    
    health)
        echo -e "${GREEN}Checking service health...${NC}"
        if command -v grpcurl &> /dev/null; then
            grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check && \
                echo -e "${GREEN}✓ User service healthy${NC}" || \
                echo -e "${RED}✗ User service not responding${NC}"
            
            grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check && \
                echo -e "${GREEN}✓ Expense service healthy${NC}" || \
                echo -e "${RED}✗ Expense service not responding${NC}"
        else
            echo -e "${YELLOW}grpcurl not installed, checking ports...${NC}"
            nc -zv localhost 50051 2>&1 | grep -q succeeded && \
                echo -e "${GREEN}✓ User service port open${NC}" || \
                echo -e "${RED}✗ User service port closed${NC}"
            
            nc -zv localhost 50052 2>&1 | grep -q succeeded && \
                echo -e "${GREEN}✓ Expense service port open${NC}" || \
                echo -e "${RED}✗ Expense service port closed${NC}"
        fi
        ;;
    
    "")
        print_usage
        ;;
    
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        echo ""
        print_usage
        exit 1
        ;;
esac
