#!/bin/bash
CURRENT_BUILD_ID="$CODEBUILD_BUILD_ID"
PROJECT_NAME="samplePipeline"

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

  echo "Build $ID has status: $STATUS"

  if [ "$STATUS" = "IN_PROGRESS" ]; then
    echo "Another build ($ID) is still in progress. Exiting..."
    exit 0
  fi
done

echo "No other builds in progress — continuing."
echo "Waiting for other builds to finish..."
sleep 60  
