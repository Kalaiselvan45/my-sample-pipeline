#!/bin/bash
set -ex  # Exit immediately on error and print each command

PS4='+ $(date "+%Y/%m/%d %H:%M:%S")  '

# Use /tmp as writable space for HOME in CodeBuild or Lambda
mkdir -p /tmp/home
chmod 700 /tmp/home
export HOME=/tmp/home
mkdir -p "$HOME/downloads" "$HOME/build/bin" "$HOME/.ssh" "$HOME/gopath"
export PATH="$HOME/build/bin:$HOME/build/go/bin:$PATH"
export GOPATH=$HOME/gopath

# Install Go
curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-${ARCH}.tar.gz" | tar -C "$HOME/build" -xz
go version

# Print Git commit hash
echo "Git commit: $CODEBUILD_RESOLVED_SOURCE_VERSION"

# Login to ECR
aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$ECR_REPO"

# Build the Docker image (this includes copying `ab` into /bin)
docker build \
  -f "$CODEBUILD_SRC_DIR/aws_ci_cd/versioning-check/dockerfile" \
  -t "$ECR_REPO:latest" \
  "$CODEBUILD_SRC_DIR"

# Run `ab semver get` from inside the image to generate the tag
export IMAGE_TAG=$(ab semver get | tail -n1)
echo "Using image tag: $IMAGE_TAG"

# Tag the image with semantic version
docker tag "$ECR_REPO:latest" "$ECR_REPO:$IMAGE_TAG"

# Push both latest and semver-tagged images
docker push "$ECR_REPO:$IMAGE_TAG"
docker push "$ECR_REPO:latest"

# Optional: Save version info back
# docker run --rm "$ECR_REPO:$IMAGE_TAG" ab semver put
