#!/bin/bash

if [[ "$CODEBUILD_BUILD_SUCCEEDING" -eq 1 ]]; then
    echo "Build succeeded. No need to send a notification."
    exit 0
fi

# Sample environment variables
# CODEBUILD_BUILD_ARN="arn:aws:codebuild:us-west-2:965106989073:build/aut-build-daily-report:0f90b898-8bfe-4480-9c2b-bafc40b55825"
# CODEBUILD_BATCH_BUILD_IDENTIFIER="certify"
# CODEBUILD_BUILD_NUMBER="1234"
# SLACK_WEBHOOK_URL="https://hooks.slack.com/services/XXXX/XXXX/XXXX"
if [[ -z "$SLACK_WEBHOOK_URL" ]]; then
    echo "SLACK_WEBHOOK_URL is not set. No notification will be sent."
    exit 0
fi

for var in CODEBUILD_BUILD_ARN CODEBUILD_BUILD_NUMBER; do
    if [[ -z "${!var}" ]]; then
        echo "Error: $var is not set."
        exit 1
    fi
done

AWS_REGION=$(echo "$CODEBUILD_BUILD_ARN" | awk -F':' '{print $4}')
AWS_ACCOUNT_ID=$(echo "$CODEBUILD_BUILD_ARN" | awk -F':' '{print $5}')
PROJECT_NAME=$(echo "$CODEBUILD_BUILD_ARN" | awk -F'[:/]' '{print $7}')
BUILD_IDENTIFIER=$(echo "$CODEBUILD_BUILD_ARN" | awk -F'[:/]' '{print $7":"$8}')
ENCODED_BUILD_IDENTIFIER=$(echo "$BUILD_IDENTIFIER" | sed 's/:/%3A/')
BUILD_URL="https://${AWS_REGION}.console.aws.amazon.com/codesuite/codebuild/${AWS_ACCOUNT_ID}/projects/${PROJECT_NAME}/build/${ENCODED_BUILD_IDENTIFIER}?region=${AWS_REGION}"

# Construct title based on whether it is batch build or normal one.
TITLE="${PROJECT_NAME}"
[[ -n "$CODEBUILD_BATCH_BUILD_IDENTIFIER" ]] && TITLE+="/${CODEBUILD_BATCH_BUILD_IDENTIFIER}"
TITLE+=" #${CODEBUILD_BUILD_NUMBER} - Failed"


curl -X POST -H "Content-type: application/json" \
    --data '{
    "attachments": [
        {
            "color": "#D00000",
            "fields": [
                { "title": "'"${TITLE}"'", "short": false },
            ],
            "footer": "<'"${BUILD_URL}"'|Click here to view build details (requires login to AWS Account ID: '"${AWS_ACCOUNT_ID}"')>"
        }
    ]
}' "$SLACK_WEBHOOK_URL"