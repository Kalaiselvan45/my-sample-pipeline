#!/bin/bash
set -e

echo "Checking if another $PROJECT_NAME build is running..."

BUILD_IDS=$(aws codebuild list-builds-for-project \
  --project-name "$PROJECT_NAME" \
  --query 'ids[?@ != `'"$CURRENT_BUILD_ID"'`] | [:5]' \
  --output text)

for ID in $BUILD_IDS; do
  STATUS=$(aws codebuild batch-get-builds \
    --ids "$ID" \
    --query 'builds[0].buildStatus' \
    --output text)

  if [ "$STATUS" = "IN_PROGRESS" ]; then
    echo "Another build ($ID) is still in progress."
    exit 0
  fi
done

echo "No other builds in progress — continuing."
