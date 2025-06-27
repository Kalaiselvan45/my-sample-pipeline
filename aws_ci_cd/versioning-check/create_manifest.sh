#!/bin/bash
set -e
echo "Creating manifest..."
COMMIT_ID=$(echo $CODEBUILD_RESOLVED_SOURCE_VERSION | cut -c1-7)
TIMESTAMP=$(date +%Y%m%d%H%M%S)
cat <<EOF > manifest.json
{
  "tool": "my-go-tool",
  "version": "v1.0.0-$COMMIT_ID-$TIMESTAMP",
  "builds": ["amd64", "arm64"],
  "built_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
cat manifest.json
