#!/bin/bash
set -e
ARCH=$1
echo "Building for $ARCH..."
GOARCH=$ARCH go build -o bin/tool-$ARCH ./cmd/tool
