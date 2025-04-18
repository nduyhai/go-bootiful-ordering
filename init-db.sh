#!/bin/bash
set -e

# This script is executed when the PostgreSQL container is started
# It creates the necessary databases for the application

# The default database (order) is already created by PostgreSQL
# We need to create the products database

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE products;
    GRANT ALL PRIVILEGES ON DATABASE products TO $POSTGRES_USER;
EOSQL

echo "Databases initialized successfully"