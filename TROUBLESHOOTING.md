# Troubleshooting Guide

## Cannot GET /web/control.html

### Problem
After successful login, you see a white page with "Cannot GET /web/control.html" error.

### Root Cause
The backend switches to static fallback mode when database connection fails, and the fallback mode was using HTTP standard library instead of Gin framework for serving static files.

### Solutions

#### Option 1: Run without database (for testing UI only)
```bash
# Using start script
./start.sh --no-db

# Or directly
DEV_NO_DB=1 ./bin/server
```

#### Option 2: Fix database connection
1. Ensure MongoDB is running locally on port 27017
2. Or update `.env` with correct MongoDB Atlas URI
3. Or use Docker Compose to start MongoDB:
```bash
docker-compose up -d mongodb
```

#### Option 3: Check static files
Ensure web files exist in `cmd/web/` directory:
```bash
ls -la cmd/web/
```

### Verification
- Health check: `curl http://localhost:8080/health`
- Web interface: `http://localhost:8080/web/index.html`
- API (requires database): `curl http://localhost:8080/api/v1/auth/login`

## Database Connection Issues

### MongoDB Connection Failed
```
Error: mongo init failed and no working fallback
```

**Solutions:**
1. Start local MongoDB: `mongod`
2. Use Docker: `docker run -d -p 27017:27017 mongo:6.0`
3. Use MongoDB Atlas and update `MONGO_URI` in `.env`

### PostgreSQL Connection Failed
```
Error: postgres init failed and no working fallback
```

**Solutions:**
1. Start local PostgreSQL
2. Update `POSTGRES_DSN` in `.env`
3. Use Docker: `docker run -d -p 5432:5432 postgres:15`

## Build Issues

### go build fails
```bash
go mod tidy
go build -o bin/server ./cmd/server
```

### Permission denied
```bash
chmod +x bin/server
```

## Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080
# Kill the process
kill -9 <PID>
```