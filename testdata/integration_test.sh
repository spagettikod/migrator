#!/bin/bash

go test -v -run TestSQLite

export INTEGRATION=1
docker compose --progress quiet up -d
#echo "Waiting 2 seconds for database to start..."
sleep 2
go test -v -run TestPostgresUtil
docker compose --progress quiet down

docker compose --progress quiet up -d
#echo "Waiting 2 seconds for database to start..."
sleep 2
go test -v -run TestPostgresMigrate
docker compose --progress quiet down
