#!/bin/bash

mkdir -p ./arch/pi/bin
env GOOS=linux GOARCH=arm GOARM=5 go build -o ./arch/pi/bin ./cmd/...
