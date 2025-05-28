#!/bin/bash

go test ./...

# PostgreSQL integration tests by running a Docker container with Postgres
export INTEGRATION=1
docker compose --file testdata/docker-compose.yml --progress quiet up -d
# Waiting 2 seconds for database to start...
sleep 2
go test -run TestPostgresUtil
docker compose --file testdata/docker-compose.yml --progress quiet down

# New iteration of PostgreSQL integration tests, restarting container to
# start fresh
docker compose --file testdata/docker-compose.yml --progress quiet up -d
# Waiting 2 seconds for database to start...
sleep 2
go test -run TestPostgresMigrate
docker compose --file testdata/docker-compose.yml --progress quiet down
