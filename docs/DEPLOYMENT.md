# Deployment Guide

This guide covers production deployment, configuration, and troubleshooting for the MS-AI Manga Platform.

## Production Recommendations

### Infrastructure
- **Database**: Use MongoDB replica set (rs0) for transactions and high availability
- **Secrets Management**: Use secret manager (Vault, AWS Secrets Manager, etc.) for JWT secrets and DB credentials
- **Containerization**: Use Docker images built from the provided Dockerfile
- **Orchestration**: Kubernetes or docker-compose for staging/production

### Security
- **CORS Configuration**: Never set `CORS_ORIGINS` to `*` in production
- **Environment Variables**: Use per-environment allowlists (comma-separated)
- **Example Production CORS**:
  ```
  CORS_ORIGINS=https://manga.example.com,https://admin.manga.example.com
  ```

### Monitoring
- **Health Checks**:
  - `/api/health` - General service health
  - `/livez` - Liveness probe
  - `/readyz` - Readiness probe (DB connectivity)
- **Logging**: Structured logging with appropriate log levels
- **Metrics**: Monitor API response times, error rates, and database performance

## Environment Configuration

### Required Environment Variables

```env
# Server Configuration
PORT=8080
ENVIRONMENT=production

# Database
MONGODB_URI=mongodb://username:password@host:port/msai?replicaSet=rs0

# Authentication
JWT_SECRET=your-256-bit-secret-here
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# CORS (Critical for security)
CORS_ORIGINS=https://yourdomain.com,https://admin.yourdomain.com

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=info
```

### Environment Validation

The application validates critical settings at startup:
- Rejects `CORS_ORIGINS=*` in production
- Validates MongoDB connection
- Checks JWT secret length
- Verifies required environment variables

## Docker Deployment

### Build Image

```bash
# Build production image
docker build -t ms-ai:latest .

# Or use the provided Makefile
make docker
```

### Run Container

```bash
# Using docker run
docker run -d \
  --name ms-ai \
  -p 8080:8080 \
  --env-file .env \
  --restart unless-stopped \
  ms-ai:latest

# Using docker-compose
docker-compose up -d
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  ms-ai:
    image: ms-ai:latest
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=mongodb://mongo:27017/msai
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGINS=https://yourdomain.com
    depends_on:
      - mongo
    restart: unless-stopped

  mongo:
    image: mongo:5.0
    volumes:
      - mongo_data:/data/db
    restart: unless-stopped

volumes:
  mongo_data:
```

## Kubernetes Deployment

### Sample Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ms-ai
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ms-ai
  template:
    metadata:
      labels:
        app: ms-ai
    spec:
      containers:
      - name: ms-ai
        image: ms-ai:latest
        ports:
        - containerPort: 8080
        env:
        - name: MONGODB_URI
          valueFrom:
            secretKeyRef:
              name: ms-ai-secrets
              key: mongodb-uri
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: ms-ai-secrets
              key: jwt-secret
        - name: CORS_ORIGINS
          value: "https://yourdomain.com"
        livenessProbe:
          httpGet:
            path: /livez
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Troubleshooting

### Database Connection Issues

**Symptoms:**
- Application fails to start
- "connection refused" errors
- Health check failures

**Solutions:**
1. Verify MongoDB URI format
2. Check network connectivity
3. Ensure MongoDB is running and accessible
4. Validate replica set configuration

### CORS Issues

**Symptoms:**
- Frontend cannot connect to API
- "CORS error" in browser console
- Authentication failures

**Solutions:**
1. Check `CORS_ORIGINS` matches your frontend domain
2. Include protocol (http/https) in CORS settings
3. For development: `http://localhost:8080,http://127.0.0.1:8080`
4. For production: `https://yourdomain.com`

### Authentication Issues

**Symptoms:**
- Login fails
- API returns 401 errors
- JWT token errors

**Solutions:**
1. Verify `JWT_SECRET` is set and matches between deployments
2. Check token expiration settings
3. Validate `CORS_ORIGINS` allows your domain
4. Check server logs for JWT validation errors

### Performance Issues

**Symptoms:**
- Slow API responses
- High memory usage
- Database timeouts

**Solutions:**
1. Check MongoDB indexes (see DATABASE.md)
2. Monitor database connection pool
3. Adjust rate limiting settings
4. Check server resources (CPU, memory)
5. Review application logs for bottlenecks

### Static File Issues

**Symptoms:**
- Web interface not loading
- 404 errors for CSS/JS files
- Blank pages

**Solutions:**
1. Verify static file serving configuration
2. Check file permissions in container
3. Ensure web assets are included in Docker image
4. Validate base URL configuration

## Backup and Recovery

### Database Backup

```bash
# MongoDB backup
mongodump --db msai --out /backup/$(date +%Y%m%d)

# Docker container backup
docker exec ms-ai-mongo mongodump --db msai --out /backup
```

### Recovery

```bash
# Restore from backup
mongorestore --db msai /backup/msai
```

## Monitoring

### Key Metrics to Monitor

- API response times
- Error rates by endpoint
- Database connection pool usage
- JWT token validation failures
- Rate limit hits
- Memory and CPU usage

### Log Aggregation

Use structured logging for better monitoring:
- Request IDs for tracing
- Error context and stack traces
- Performance metrics
- Security events

## Scaling Considerations

### Horizontal Scaling

- Stateless design supports multiple instances
- Use load balancer for distribution
- Shared MongoDB replica set
- Session storage in database (not memory)

### Database Scaling

- Use MongoDB sharding for large datasets
- Implement read replicas for read-heavy workloads
- Monitor index usage and query performance
- Consider archiving old data

### CDN Integration

For better performance with manga images:
- Use CDN for cover images and chapter pages
- Implement image optimization
- Cache static assets
- Consider regional CDN deployment
   ```
   # Development
   CORS_ORIGINS=http://localhost:8080

   # Production (single domain)
   CORS_ORIGINS=https://example.com

   # Production (multiple domains)
   CORS_ORIGINS=https://example.com,https://www.example.com,https://app.example.com
   ```

   ⚠️ WARNING: Never use `CORS_ORIGINS=*` in production

5. **Additional Settings to Check**
   - JWT token validity and expiration
   - Server port matches client configuration
   - If using a proxy:
     - WebSocket headers are properly forwarded
     - Timeout settings are appropriate for WebSocket
     - SSL/TLS termination is properly configured
