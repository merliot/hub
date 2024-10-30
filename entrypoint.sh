#!/bin/sh
echo "Warming up Go compiler cache..."

# Compile a dummy Go program to preload Go packages
echo 'package main; func main() {}' > /tmp/dummy.go
go build -o /dev/null /tmp/dummy.go && rm /tmp/dummy.go

echo "Go cache warmed up. Executing: $@"
exec "$@"  # Run whatever command is passed from CMD (default is /hub)
