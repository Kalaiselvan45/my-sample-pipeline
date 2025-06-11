#!/bin/bash
set -e
export AWS_REGION=us-west-2
export ECR_REPOSITORY=123456789012.dkr.ecr.us-west-2.amazonaws.com/my-sample-image
export IMAGE_TAG=v1.0.0-$(date +%Y%m%d%H%M%S)
echo "Starting build with IMAGE_TAG=$IMAGE_TAG"