#!/bin/bash

mkdir -p ./arch/windows/bin
env GOOS=windows GOARCH=amd64 go build -o ./arch/windows/bin ./cmd/...
