#!/bin/bash

docker compose up -d && go test -run TestPostgresInitialized github.com/spagettikod/migrator; docker compose down
docker compose up -d && go test -run TestPostgresInit github.com/spagettikod/migrator; docker compose down
docker compose up -d && go test -run TestPostgresVersion github.com/spagettikod/migrator; docker compose down
docker compose up -d && go test -run TestPostgresMigrate github.com/spagettikod/migrator; docker compose down
