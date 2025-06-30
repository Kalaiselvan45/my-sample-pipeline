package impl

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aqfer/aqfer-go/xlib/aqfs"
)

type CommitCmd struct {
	Get CommitGetCmd `cmd:"" help:"Get the latest commit"`
	Put CommitPutCmd `cmd:"" help:"Save the latest commit"`
}

type CommitGetCmd struct {
	Default string `help:"Default commit if not found" default:"dummy-hash"`
}

type CommitPutCmd struct {
	CurrentCommit string `help:"CODEBUILD_RESOLVED_SOURCE_VERSION used for writing the current commit" env:"CODEBUILD_RESOLVED_SOURCE_VERSION" required:""`
}

func getCommitS3Path() string {
	return codeBuildInfo.getStoragePath("commit", "commit.txt")
}

func putCommit(cmd CLI) error {
	s3Path := getCommitS3Path()
	err := aqfs.SaveFile(s3Path, []byte(cmd.GitCommit.Put.CurrentCommit))
	if err != nil {
		return err
	}
	slog.Info("Uploaded commit file successfully.", "file", s3Path, "commit", cmd.GitCommit.Put.CurrentCommit)
	return nil
}

func getCommit(cmd CLI) error {
	s3Path := getCommitS3Path()
	content, err := aqfs.LoadFile(s3Path)
	if err != nil {
		if !strings.Contains(err.Error(), notFound) {
			slog.Error("Failed to load file from s3", "s3_file", s3Path, "error", err)
			return err
		}
		content = []byte(cmd.GitCommit.Get.Default)
		slog.Warn("Commit file not found.", "s3_file", s3Path, "using_default_version", cmd.GitCommit.Get.Default)
	}
	fmt.Println(string(content))
	return nil
}
