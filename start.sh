#!/bin/bash

# This file is for quick starting on linux to avoid manually switching architcture
export GOOS=linux
export GOARCH=amd64
go run main.go "$@"


