#!/bin/bash
set -euo pipefail

export HOME=/tmp/home
mkdir -p "$HOME"

# Load work area and init tools
eval "$(ab init --mkdirs build)"
ab go-cache get

# Get semver version (e.g. v1.2.3)
export IMAGE_TAG=$(ab semver get | tail -n1)
echo "Using image tag: $IMAGE_TAG"

# Build Go binary
echo "Building Go binary..."
go build -o myself

# Docker login
aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$ECR_REPO"

# Build and tag Docker image
docker build -t "$ECR_REPO:$IMAGE_TAG" -t "$ECR_REPO:latest" .
docker push "$ECR_REPO:$IMAGE_TAG"
docker push "$ECR_REPO:latest"

# Put semver (confirm version used)
ab semver put
