#!/bin/bash
set -ex

export HOME=/tmp/home
mkdir -p "$HOME"

# # Load work area and init tools
eval "$(ab init --mkdirs build)"

# Get semver version (e.g. v1.2.3)
export IMAGE_TAG=$(ab semver get | tail -n1)
echo "Using image tag: $IMAGE_TAG"


# Docker login
aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$ECR_REPO"

# Build and tag Docker image
docker build \
  -f "$CODEBUILD_SRC_DIR/aws_ci_cd/versioning-check/dockerfile" \
  -t "$ECR_REPO:$IMAGE_TAG" \
  -t "$ECR_REPO:latest" \
  "$CODEBUILD_SRC_DIR"
docker push "$ECR_REPO:$IMAGE_TAG"
docker push "$ECR_REPO:latest"

Put semver (confirm version used)
ab semver put
