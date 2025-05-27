#!/bin/bash

go test -v ./...

export INTEGRATION=1
docker compose --file testdata/docker-compose.yml --progress quiet up -d
# Waiting 2 seconds for database to start...
sleep 2
go test -v -run TestPostgresUtil
docker compose --file testdata/docker-compose.yml --progress quiet down

docker compose --file testdata/docker-compose.yml --progress quiet up -d
# Waiting 2 seconds for database to start...
sleep 2
go test -v -run TestPostgresMigrate
docker compose --file testdata/docker-compose.yml --progress quiet down
