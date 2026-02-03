#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting all services...${NC}"

# Start PostgreSQL
echo -e "${YELLOW}Starting PostgreSQL...${NC}"
docker run -d --name splitwise-db \
  -p 5433:5432 \
  -e POSTGRES_DB=splitwise \
  -e POSTGRES_USER=splitwise \
  -e POSTGRES_PASSWORD=splitwise \
  postgres:15-alpine 2>/dev/null || docker start splitwise-db

sleep 2

# Start User Service
echo -e "${YELLOW}Starting User Service on :50051...${NC}"
bazel run //services/user:user_service &
USER_PID=$!

sleep 3

# Start Expense Service  
echo -e "${YELLOW}Starting Expense Service on :50052...${NC}"
APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service &
EXPENSE_PID=$!

sleep 3

# Start API Gateway
echo -e "${YELLOW}Starting API Gateway on :8080...${NC}"
go run ./api-gateway/main.go &
GATEWAY_PID=$!

sleep 2

# Start Web Frontend
echo -e "${YELLOW}Starting Web Frontend on :5173...${NC}"
cd web && npm run dev &
WEB_PID=$!
cd ..

sleep 2

echo -e "${GREEN}✓ All services started!${NC}"
echo -e "  ${YELLOW}User Service:${NC} localhost:50051"
echo -e "  ${YELLOW}Expense Service:${NC} localhost:50052"
echo -e "  ${YELLOW}API Gateway:${NC} http://localhost:8080"
echo -e "  ${YELLOW}Web Frontend:${NC} http://localhost:5173"
echo ""
echo -e "${YELLOW}Test the API:${NC}"
echo -e "  curl http://localhost:8080/health"
echo ""
echo -e "${YELLOW}Open the app:${NC}"
echo -e "  http://localhost:5173"
echo ""
echo -e "${RED}Press Ctrl+C to stop all services${NC}"

# Wait for Ctrl+C
trap "echo -e '\n${YELLOW}Stopping services...${NC}'; kill $USER_PID $EXPENSE_PID $GATEWAY_PID $WEB_PID 2>/dev/null; docker stop splitwise-db 2>/dev/null; exit" INT

wait
