#!/bin/bash
# Database initialization script
# This script runs ONCE when PostgreSQL container starts for the first time
# It creates required extensions and sets up any base configuration

set -e

echo "======================================"
echo "Initializing Money Planner Database"
echo "======================================"

# Create required extensions
echo "Creating PostgreSQL extensions..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Enable UUID extension
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- Enable JSON extension
    CREATE EXTENSION IF NOT EXISTS "plpgsql";

    -- Set default schema search path
    ALTER USER "$POSTGRES_USER" SET search_path TO public;

    -- Enable query logging for debugging (optional - disable in production if needed)
    ALTER SYSTEM SET log_statement = 'all';
    ALTER SYSTEM SET log_duration = 'on';

    -- Connection limits
    ALTER SYSTEM SET max_connections = 200;
EOSQL

echo ""
echo "======================================"
echo "Database initialization complete!"
echo "======================================"
echo ""
echo "Notes:"
echo "  - Migrations will be run automatically by the Go backend on startup"
echo "  - Database URL: postgres://${POSTGRES_USER}@postgres:5432/${POSTGRES_DB}"
echo "  - Check backend logs for migration details"
echo ""
