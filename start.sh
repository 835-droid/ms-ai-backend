#!/bin/bash

# MS-AI Backend Start Script
# This script builds and runs the Go backend server

set -e

echo "🚀 Starting MS-AI Backend..."

# Check for DEV_NO_DB flag
DEV_NO_DB_FLAG=""
if [ "$1" = "--no-db" ] || [ "$DEV_NO_DB" = "1" ]; then
    DEV_NO_DB_FLAG="DEV_NO_DB=1"
    echo "📋 Running in development mode without database..."
fi

# Navigate to the project directory
cd "$(dirname "$0")"

# Check if .env exists, if not copy from .env.example
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        echo "📋 Copying .env.example to .env..."
        cp .env.example .env
        echo "⚠️  Please edit .env file with your configuration before running again."
        exit 1
    else
        echo "❌ Neither .env nor .env.example found!"
        exit 1
    fi
fi

# Build the Go server
echo "🔨 Building Go server..."
go build -o bin/server ./cmd/server

# Run the server in background
echo "🌐 Starting server on http://localhost:8080..."
if [ -n "$DEV_NO_DB_FLAG" ]; then
    $DEV_NO_DB_FLAG ./bin/server &
else
    ./bin/server &
fi
SERVER_PID=$!

# Save PID for later stopping
echo $SERVER_PID > server.pid

echo "✅ Server started successfully!"
if [ -n "$DEV_NO_DB_FLAG" ]; then
    echo "🌍 Access the web interface at: http://localhost:8080/web/index.html"
    echo "📋 Note: API endpoints will not work without database"
else
    echo "🌍 Access the application at: http://localhost:8080/web/index.html"
fi
echo "🛑 To stop the server, run: kill $SERVER_PID"

# Wait for server to keep script running
wait $SERVER_PID