#!/bin/bash
set -e

# Create the money_planner database for pgAdmin health checks
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
    CREATE DATABASE money_planner;
EOSQL
