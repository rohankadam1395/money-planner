# Money Planner - Deployment Guide

## Overview

This guide covers deploying Money Planner to production with proper database migrations, environment configuration, and container orchestration.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Production                        │
├─────────────────────────────────────────────────────┤
│                                                      │
│  Frontend (Next.js)                                 │
│  └─ PORT: 3000                                      │
│     └─ Requests → Backend API                       │
│                                                      │
│  Backend (Go)                                       │
│  └─ PORT: 8080                                      │
│     ├─ Auto-runs migrations on startup              │
│     ├─ Health checks enabled                        │
│     └─ Connects to PostgreSQL                       │
│                                                      │
│  PostgreSQL                                         │
│  └─ PORT: 5432                                      │
│     ├─ Init script runs once (creates extensions)   │
│     ├─ Migrations run on backend startup            │
│     └─ Persistent volume: postgres_data_prod        │
│                                                      │
└─────────────────────────────────────────────────────┘
```

## Quick Start - Production Deployment

### 1. Prepare Environment

```bash
# Copy the environment template
cp .env.example .env.prod

# Edit with your production values
nano .env.prod
# OR
code .env.prod
```

**Required values to set:**
- `DB_PASSWORD`: Strong password (generate with: openssl rand -base64 32)
- `JWT_SECRET`: Strong secret (generate with: openssl rand -base64 32)
- `API_URL`: Your production domain
- `NEXT_PUBLIC_API_URL`: Frontend's view of API URL

### 2. Start All Services

```bash
# Using production docker-compose
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Watch logs
docker-compose -f docker-compose.prod.yml logs -f backend

# Check health
curl http://localhost:8080/health
```

### 3. Verify Migration Execution

```bash
# Check backend logs for migration messages
docker-compose -f docker-compose.prod.yml logs backend | grep -i migration

# Expected output:
# Starting database migrations...
# Running migration: 001_create_statements_schema.sql
# ✓ Migration applied: 001_create_statements_schema.sql
# ✓ All migrations completed successfully
```

### 4. Database Verification

```bash
# Connect to database
docker exec -it money-planner-postgres-prod psql -U postgres -d money_planner

# Check tables
\dt

# Check migrations table
SELECT * FROM schema_migrations;

# Exit
\q
```

## Migration Strategy

### How Migrations Work

1. **Init Script** (Runs once on container first start)
   - Creates PostgreSQL extensions (UUID, JSON)
   - Sets up base configuration
   - Located: `init-db.sh`

2. **Auto-Migrations** (Run on backend startup)
   - Backend connects to database
   - Reads migration files from `backend/internal/db/migrations/`
   - Creates `schema_migrations` table to track executed migrations
   - Executes any new migrations that haven't been run
   - Idempotent: Safe to restart backend multiple times

### Why This Approach?

✅ **Advantages:**
- No separate migration tool needed
- Runs automatically on deployment
- Embedded in Go binary (single deployment unit)
- Works offline (no external service calls)
- Idempotent and safe to retry
- Tracks applied migrations in database

⚠️ **Considerations:**
- Tie migrations to application version (track in git)
- Test migrations before production deployment
- Plan for backwards-compatible schema changes
- Monitor migration execution in logs

## Configuration Management

### Environment Variables

All configuration is environment-driven:

```bash
# Database
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable

# API
JWT_SECRET=your-secret-here
PORT=8080
LOG_LEVEL=info

# Frontend
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

### .env File Management

- **Local dev**: `.env.local` (in .gitignore)
- **Production**: `.env.prod` (in .gitignore)
- **Template**: `.env.example` (committed to git)

```bash
# Never commit real .env files
echo ".env.local" >> .gitignore
echo ".env.prod" >> .gitignore
```

## Container Orchestration

### Docker Compose (For simple deployments)

```bash
# Start
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Stop
docker-compose -f docker-compose.prod.yml down

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Scale services (if needed)
docker-compose -f docker-compose.prod.yml up -d --scale backend=2
```

### Kubernetes (For production at scale)

```yaml
# Example k8s deployment structure:
# deployments/
#   ├── postgres-deployment.yml
#   ├── backend-deployment.yml
#   ├── frontend-deployment.yml
#   └── configmap.yml
```

Would require:
- StatefulSet for PostgreSQL (persistent data)
- Deployment for backend (multiple replicas)
- Deployment for frontend (multiple replicas)
- Service for database access
- ConfigMap for configuration
- Secrets for sensitive data

## Security Best Practices

### ✅ Implemented

- [x] Non-root user in containers
- [x] Health checks configured
- [x] Resource limits set
- [x] Environment-based secrets
- [x] No hardcoded credentials in code
- [x] HTTPS ready (use reverse proxy)

### ⚠️ Recommended for Production

- [ ] Use managed database (RDS, Cloud SQL) instead of container
- [ ] Implement secret management (Vault, AWS Secrets Manager)
- [ ] Use reverse proxy (nginx, Caddy) for HTTPS
- [ ] Enable database encryption at rest
- [ ] Set up monitoring and alerting
- [ ] Configure backup strategy
- [ ] Use container registry authentication
- [ ] Implement rate limiting on API

## Monitoring & Logs

### Health Checks

```bash
# API health
curl http://localhost:8080/health

# Database health (via docker-compose)
docker-compose -f docker-compose.prod.yml ps
# All services should show "healthy"
```

### Logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs

# Specific service
docker-compose -f docker-compose.prod.yml logs backend
docker-compose -f docker-compose.prod.yml logs postgres

# Live tail
docker-compose -f docker-compose.prod.yml logs -f backend

# Filter by keyword
docker-compose -f docker-compose.prod.yml logs | grep -i error
```

## Troubleshooting

### Backend won't start

```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs backend

# Common issues:
# 1. DATABASE_URL not set or incorrect
# 2. PostgreSQL not healthy yet
# 3. Port already in use

# Wait for database to be healthy
docker-compose -f docker-compose.prod.yml logs postgres | grep "database system is ready"
```

### Database connection fails

```bash
# Check PostgreSQL is running
docker-compose -f docker-compose.prod.yml ps postgres

# Test connection
docker exec money-planner-postgres-prod psql -U postgres -d money_planner -c "SELECT 1;"

# Check DATABASE_URL format
# postgres://user:password@host:5432/dbname?sslmode=disable
```

### Migrations not running

```bash
# Check backend logs for migration messages
docker-compose -f docker-compose.prod.yml logs backend | grep -i migration

# If migrations stuck, check schema_migrations table
docker exec money-planner-postgres-prod psql -U postgres -d money_planner \
  -c "SELECT * FROM schema_migrations;"
```

## Rollback Strategy

### Database Schema Rollback

If a migration fails:

1. **Fix the migration file** (`backend/internal/db/migrations/XXX.sql`)
2. **Remove the failed migration record** from `schema_migrations` table:
   ```sql
   DELETE FROM schema_migrations WHERE name = '001_create_statements_schema.sql';
   ```
3. **Restart backend** - migration will re-run

### Application Rollback

If deployment has issues:

```bash
# Stop all services
docker-compose -f docker-compose.prod.yml down

# Revert to previous version (if using version tags)
docker pull your-registry/money-planner-backend:v1.0.0

# Start again with previous version
docker-compose -f docker-compose.prod.yml up -d
```

## Deployment Checklist

- [ ] Environment variables configured (.env.prod)
- [ ] Database password is strong
- [ ] JWT_SECRET is generated (not default)
- [ ] API_URL matches actual domain
- [ ] Docker images built and available
- [ ] Persistent volume configured for database
- [ ] Health checks passing
- [ ] Migrations executed successfully
- [ ] API responding to requests
- [ ] Frontend can reach backend
- [ ] Logs are being collected
- [ ] Backup strategy in place
- [ ] Monitoring alerts configured

## Next Steps

1. **Test locally first** with production config
2. **Deploy to staging** before production
3. **Monitor logs** during initial deployment
4. **Verify transactions** are being extracted
5. **Set up automated backups**
6. **Configure monitoring/alerting**

## Support

For issues with deployments:
1. Check logs: `docker-compose logs`
2. Verify environment variables: `docker inspect`
3. Test database connection directly
4. Review this guide's troubleshooting section
