# Money Planner - Local Development Setup

## Quick Start (5 minutes)

### 1. Start Database

```bash
# From project root
docker-compose up -d

# Verify it's running
docker-compose ps
# Should show: money-planner-postgres-dev (running)
#            money-planner-pgadmin (running)
```

### 2. Set Environment Variable

```bash
# PowerShell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/money_planner?sslmode=disable"
$env:JWT_SECRET = "local-dev-secret"

# Or add to .env.local and load it in your IDE
```

### 3. Start Backend

```bash
cd backend
go run ./cmd/statement-import-api

# Expected output:
# ✓ All migrations completed successfully
# Starting statement import API server on port 8080
# [GIN-debug] Listening and serving HTTP on :8080
```

### 4. Start Frontend

```bash
# In another terminal
cd frontend
npm run dev

# Expected output:
# - ready started server on 0.0.0.0:3000, url: http://localhost:3000
```

### 5. Open in Browser

- **Frontend**: http://localhost:3000
- **Backend**: http://localhost:8080/health
- **pgAdmin**: http://localhost:5050 (admin@example.com / admin)

---

## Local Development Services

### PostgreSQL
- **URL**: `postgres://localhost:5432/money_planner`
- **User**: postgres
- **Password**: postgres
- **Database**: money_planner
- **Port**: 5432
- **Container**: money-planner-postgres-dev

### pgAdmin (Database UI)
- **URL**: http://localhost:5050
- **Email**: admin@example.com
- **Password**: admin
- **Purpose**: Browse tables, run SQL queries, manage schema

### Backend API
- **URL**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Port**: 8080
- **Language**: Go 1.21+
- **Auto-migrations**: Runs on startup

### Frontend
- **URL**: http://localhost:3000
- **Framework**: Next.js with React
- **Node**: 18+
- **Hot reload**: Yes, automatic on file save

---

## Workflow for Development

### 1. Backend Development

```bash
cd backend

# Environment must be set
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/money_planner?sslmode=disable"
export JWT_SECRET="local-dev-secret"

# Run with hot-reload using air (if installed)
go install github.com/cosmtrek/air@latest
air

# Or simple run
go run ./cmd/statement-import-api
```

**Making changes:**
- Edit Go files
- Backend restarts automatically (if using air) or manually restart
- Test at http://localhost:8080/health

### 2. Frontend Development

```bash
cd frontend

# Install dependencies (first time only)
npm install

# Run with hot-reload
npm run dev

# Expected: http://localhost:3000 with live reload on save
```

**Making changes:**
- Edit React/TypeScript files
- Browser auto-reloads via Next.js dev server
- Check browser console for errors

### 3. Database Development

#### Option A: pgAdmin UI
1. Go to http://localhost:5050
2. Login with admin@example.com / admin
3. Add server: hostname=postgres, user=postgres, password=postgres
4. Browse tables and run queries

#### Option B: Command Line
```bash
# Connect to database
docker exec -it money-planner-postgres-dev psql -U postgres -d money_planner

# List tables
\dt

# View migrations
SELECT * FROM schema_migrations;

# Run a query
SELECT COUNT(*) FROM transactions;

# Exit
\q
```

#### Option C: VS Code Extension
- Install: PostgreSQL Explorer
- Configure connection to localhost:5432
- Browse and query directly from editor

---

## Common Development Tasks

### Add a New Migration

1. Create new migration file:
   ```bash
   touch backend/internal/db/migrations/002_add_new_table.sql
   ```

2. Write your SQL:
   ```sql
   -- backend/internal/db/migrations/002_add_new_table.sql
   CREATE TABLE new_table (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

3. Restart backend:
   ```bash
   # Backend will auto-detect and run the new migration
   go run ./cmd/statement-import-api
   ```

4. Verify:
   ```bash
   docker exec money-planner-postgres-dev psql -U postgres -d money_planner -c "\dt"
   ```

### Clear Database and Start Fresh

```bash
# Option 1: Delete volumes and restart
docker-compose down -v
docker-compose up -d

# Option 2: Just truncate tables
docker exec money-planner-postgres-dev psql -U postgres -d money_planner <<EOF
TRUNCATE TABLE statements CASCADE;
TRUNCATE TABLE transactions CASCADE;
TRUNCATE TABLE import_jobs CASCADE;
DELETE FROM schema_migrations;
EOF
```

### Test an API Endpoint

```bash
# Get JWT token (test endpoint)
curl -X POST http://localhost:8080/api/auth/login

# Upload statement (requires auth)
curl -X POST http://localhost:8080/api/statements/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@sample.csv" \
  -F "bank_code=HDFC"

# Get preview
curl -X GET "http://localhost:8080/api/statements/{id}/preview" \
  -H "Authorization: Bearer $TOKEN"
```

### Debug Mode

```bash
# Enable verbose logging
export LOG_LEVEL=debug
go run ./cmd/statement-import-api

# Enable SQL query logging
# Already enabled in docker-compose.yml
docker-compose logs -f postgres | grep "statement:"
```

---

## Troubleshooting

### "Connection refused" on backend startup

```bash
# Check database is running
docker-compose ps

# If not running
docker-compose up -d

# Check it's healthy
docker-compose ps postgres
# Status should show: healthy
```

### "Database does not exist" error

```bash
# Create database
docker exec money-planner-postgres-dev psql -U postgres -c "CREATE DATABASE money_planner;"

# Or restart (init script creates it)
docker-compose down -v
docker-compose up -d
```

### Migrations not running

```bash
# Check logs
docker-compose logs backend | grep -i migration

# If stuck, manually delete migration record
docker exec money-planner-postgres-dev psql -U postgres -d money_planner <<EOF
DELETE FROM schema_migrations WHERE name = '001_create_statements_schema.sql';
EOF

# Restart backend
```

### Port already in use

```bash
# PostgreSQL (5432)
sudo lsof -i :5432
# Kill process if needed: kill -9 <PID>

# Backend (8080)
sudo lsof -i :8080

# Frontend (3000)
sudo lsof -i :3000

# pgAdmin (5050)
sudo lsof -i :5050

# Or change ports in docker-compose.yml
```

### pgAdmin can't connect to database

```bash
# Add server in pgAdmin:
# Host: postgres (NOT localhost!)
# Port: 5432
# User: postgres
# Password: postgres
# Database: money_planner
```

---

## VS Code Setup (Recommended)

### Extensions to Install

```json
{
  "go.linterTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.useLanguageServer": true,
  "typescript.tsdk": "node_modules/typescript/lib"
}
```

**Recommended Extensions:**
- Go (golang.go)
- PostgreSQL Explorer (cweijan.vscode-postgresql-client2)
- Thunder Client (rangav.vscode-thunder-client) - for API testing
- REST Client (humao.rest-client) - for API testing

### Launch Configuration (.vscode/launch.json)

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Backend (Go)",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/statement-import-api",
      "env": {
        "DATABASE_URL": "postgres://postgres:postgres@localhost:5432/money_planner?sslmode=disable",
        "JWT_SECRET": "local-dev-secret"
      },
      "args": [],
      "showLog": "true"
    }
  ]
}
```

---

## Git Workflow for Development

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes, commit regularly
git add .
git commit -m "feat: description of change"

# Test locally
npm run test  # frontend
go test ./... # backend

# Push to remote
git push -u origin feature/your-feature

# Create PR on GitHub
```

---

## Dependencies Management

### Backend (Go)

```bash
cd backend

# Add dependency
go get github.com/some/package

# Update dependencies
go mod tidy

# View dependency tree
go mod graph
```

### Frontend (Node)

```bash
cd frontend

# Add dependency
npm install package-name

# Update dependencies
npm update

# Check for security vulnerabilities
npm audit
npm audit fix
```

---

## Performance Tips

### Fast Rebuild

```bash
# Use air for live reload (faster than restarting)
go install github.com/cosmtrek/air@latest
cd backend && air
```

### Database Optimization

- Use indexes: already included in migrations
- Monitor slow queries: `LOG_LEVEL=debug`
- Use pgAdmin to analyze query performance

### Frontend Hot Reload

- Enabled by default in Next.js dev server
- Changes reflect instantly without full reload
- Check browser console for errors

---

## When Ready to Deploy

1. **Test locally first**
   - ✅ Backend running
   - ✅ Frontend running
   - ✅ Database connected
   - ✅ Migrations applied
   - ✅ API responding
   - ✅ Transactions extracting

2. **Run tests**
   ```bash
   go test ./backend/...
   npm test --frontend
   ```

3. **Build Docker images**
   ```bash
   docker build -t money-planner-backend:latest backend/
   docker build -t money-planner-frontend:latest frontend/
   ```

4. **Use production compose**
   ```bash
   docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
   ```

---

## Additional Resources

- **Backend**: See `backend/README.md`
- **Frontend**: See `frontend/README.md`
- **Deployment**: See `DEPLOYMENT.md`
- **API Docs**: See `backend/docs/api.md`

## Support

Having issues? Check:
1. Logs: `docker-compose logs -f`
2. Troubleshooting section above
3. Verify all services are running: `docker-compose ps`
4. Check environment variables are set
5. Make sure ports aren't in use on your machine
