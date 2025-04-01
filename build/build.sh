#!/bin/bash

cd ../
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GIN_MODE=release go build -o build/server_arm
CGO_ENABLED=0 GIN_MODE=release go build -o build/server_linux
