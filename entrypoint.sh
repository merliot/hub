#!/bin/sh
echo "Warming up Go compiler cache..."

# Compile the hub on different archs to preload Go packages.  This
# makes subsequent compiles of devices by the hub itself much faster.

# Print Go and TinyGo versions
go version
tinygo version

echo "Building ARM arch..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "-s -w" -tags rpi -o /dev/null ./cmd
echo "Building x86-64 arch..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -tags x86_64 -o /dev/null ./cmd

echo "Go cache warmed up. Executing: $@"
exec "$@"  # Run whatever command is passed from CMD (default is /hub)
