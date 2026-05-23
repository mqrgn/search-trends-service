#!/bin/bash

# Получаем путь к GOPATH и убираем возможные символы возврата каретки
GOPATH_RAW=$(go env GOPATH)
GOPATH_DIR=$(echo $GOPATH_RAW | tr -d '\r')
HEY_PATH="$GOPATH_DIR/bin/hey"

echo "=== Load Testing Search Trends Service ==="
echo ""

if [ ! -f "$HEY_PATH.exe" ]; then
    echo "Installing hey..."
    go install github.com/rakyll/hey@latest
fi

echo "Starting load test..."
echo ""

echo "Test 1: GET /api/v1/top (100k requests, 100 concurrent)"
"$HEY_PATH" -n 100000 -c 100 -m GET http://localhost:8080/api/v1/top?limit=10

echo ""
echo "Test 2: GET /api/v1/top (high concurrency - 500 concurrent)"
"$HEY_PATH" -n 50000 -c 500 -m GET http://localhost:8080/api/v1/top?limit=10

echo ""
echo "Test 3: POST /api/v1/stoplist (write operations)"
"$HEY_PATH" -n 10000 -c 50 -m POST -H "Content-Type: application/json" -d '{"query":"test"}' http://localhost:8080/api/v1/stoplist

echo ""
echo "=== Load testing complete ==="