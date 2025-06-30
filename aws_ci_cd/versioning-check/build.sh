#!/bin/bash
set -ex  # Exit on any error

PS4='+ $(date "+%Y/%m/%d %H:%M:%S")  '

# /tmp is used as only this is writable in Lambda
mkdir -p /tmp/home
chmod 700 /tmp/home
export HOME=/tmp/home
mkdir -p $HOME/downloads $HOME/build/bin $HOME/.ssh $HOME/gopath
export PATH="$HOME/build/bin:$HOME/build/go/bin:$PATH"

export GOPATH=$HOME/gopath
# Install Go
curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-${ARCH}.tar.gz" | tar -C $HOME/build -xz
go version

echo "Setting up Git SSH configuration..."
export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/id_rsa -o UserKnownHostsFile=$HOME/.ssh/known_hosts"

cat > $HOME/.ssh/known_hosts <<EOF
$CI_GITHUB_SSH_RSA
EOF

cat > $HOME/.ssh/id_rsa <<GIT
$CI_GITHUB_SSH_PRIVATE_KEY
GIT

chmod 0700 $HOME/.ssh
chmod 0600 $HOME/.ssh/known_hosts ~/.ssh/id_rsa
git config --global url."git@github.com:".insteadof "https://github.com/"
echo "Git setup completed."

(
  cd "$CODEBUILD_SRC_DIR/spark/aws-ci-cd/cost-estimator"
  go build -o "$HOME/build/bin/cost-estimator" .
)

# Install uber-ab
(
  cd "$CODEBUILD_SRC_DIR/spark/aws-ci-cd/amdp-builder/cmd/uber-ab"
  go build -o "$HOME/build/bin/uber-ab" .
  ln -s "$HOME/build/bin/uber-ab" "$HOME/build/bin/ab"
)

# eval "$(ab init --mkdirs build)"
echo $CODEBUILD_RESOLVED_SOURCE_VERSION

# Get semver version (e.g. v1.2.3)
export IMAGE_TAG="v3.0.0-$(date +%Y%m%d%H%M%S)-$(git rev-parse --short $CODEBUILD_RESOLVED_SOURCE_VERSION)"
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

# Put semver (confirm version used)
# ab semver put
