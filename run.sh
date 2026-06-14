#!/usr/bin/env bash

mkdir -p local/{bin,etc}
GOOS=linux go build -tags "debug" -o local/bin/queue cmd/main.go 

# pushd mock_server
# GOOS=linux go build -o ../local/bin/mock cmd/mock/main.go
# mkdir -p ../local/etc/mock/
# cp configs/config.yaml ../local/etc/mock/config.yaml
# popd
docker compose build --no-cache
docker-compose up

