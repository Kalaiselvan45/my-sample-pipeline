package impl

import (
	"fmt"
	"os"
	"regexp"
)

type CodeBuildInfo struct {
	Region       string
	AccountID    string
	ProjectName  string
	BatchBuildID string
}

var (
	codeBuildARNRegex = regexp.MustCompile(`^arn:aws:codebuild:([a-z0-9-]+):(\d+):build(-batch)?/([a-zA-Z0-9_-]+):([a-f0-9-]{36})$`)
	codeBuildInfo     CodeBuildInfo
)

// ExtractInfo extracts the AWS region, account ID, project name, and BatchBuildID from the available ARN.
func ExtractInfo() (CodeBuildInfo, error) {
	arn := os.Getenv("CODEBUILD_BATCH_BUILD_ARN") // arn:aws:codebuild:us-west-2:965106989073:build-batch/tesest:b0e8cf7b-6e81-4b0a-ae2a-950b9af1b052
	if arn == "" {
		arn = os.Getenv("CODEBUILD_BUILD_ARN") // arn:aws:codebuild:us-west-2:965106989073:build/amdp-builder:1fdc9cf3-d20b-4eab-aa0a-7a52a99c547d
	}
	if arn == "" {
		return CodeBuildInfo{}, fmt.Errorf("no valid CodeBuild ARN found")
	}

	matches := codeBuildARNRegex.FindStringSubmatch(arn)
	if len(matches) != 6 {
		return CodeBuildInfo{}, fmt.Errorf("invalid ARN format")
	}

	return CodeBuildInfo{
		Region:       matches[1],
		AccountID:    matches[2],
		ProjectName:  matches[4],
		BatchBuildID: matches[5],
	}, nil
}

func (c CodeBuildInfo) getStoragePath(paths ...string) string {
	basePath := fmt.Sprintf("s3://codebuild-%s-%s/%s", c.Region, c.AccountID, c.ProjectName)
	return getPath(basePath, paths)
}

func (c CodeBuildInfo) getBatchBuildPath(paths ...string) string {
	basePath := fmt.Sprintf("s3://codebuild-%s-%s-tmp/%s/batch_build_id=%s", c.Region, c.AccountID, c.ProjectName, c.BatchBuildID)
	return getPath(basePath, paths)
}

func getPath(basePath string, paths []string) string {
	for _, path := range paths {
		basePath = fmt.Sprintf("%s/%s", basePath, path)
	}
	return basePath
}
