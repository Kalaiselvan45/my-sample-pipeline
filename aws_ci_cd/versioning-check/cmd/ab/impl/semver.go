package impl

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/aqfer/aqfer-go/xlib/aqfs"
)

type SemverCmd struct {
	Get SemverGetCmd `cmd:"" help:"Get the current semantic version"`
	Put SemverPutCmd `cmd:"" help:"Save the current semantic version"`
}

type SemverGetCmd struct {
	Default string `help:"Default version if not found" default:"0.0.0"`
	Bump    string `help:"Bump version type" enum:"major,minor,patch,none" default:"minor"`
	Path    string `help:"Optional: Specifies the path to the semver file in s3." short:"p"`
}

type SemverPutCmd SemverGetCmd

const notFound = "operation error S3: GetObject, https response error StatusCode: 404"

func getSemverS3AndLocalPath(cmd CLI) (string, string) {
	return codeBuildInfo.getStoragePath("semver", "image.ver"), fmt.Sprintf("%s/%s/%s", cmd.Home, "semver", "image.ver")
}

func putSemver(cmd CLI) error {
	s3Path, localPath := getSemverS3AndLocalPath(cmd)
	version, err := aqfs.LoadFile(localPath)
	if err != nil {
		return err
	}
	err = aqfs.SaveFile(s3Path, version)
	if err != nil {
		return err
	}
	slog.Info("Uploaded semver file successfully.", "file", s3Path, "version", string(version))
	return nil
}

func getSemver(cmd CLI) error {
	s3Path, localPath := getSemverS3AndLocalPath(cmd)
	if cmd.Semver.Get.Path != "" {
		s3Path = cmd.Semver.Get.Path
	}
	var srcVersion string
	content, err := aqfs.LoadFile(s3Path)
	if err != nil {
		if !strings.Contains(err.Error(), notFound) {
			slog.Error("Failed to load file from s3", "s3_file", s3Path, "error", err)
			return err
		}
		content = []byte(cmd.Semver.Get.Default)
		slog.Warn("Semver file not found.", "s3_file", s3Path, "using_default_version", cmd.Semver.Get.Default)
	} else {
		srcVersion = string(content)
	}
	content, err = incrementVersion(content, cmd.Semver.Get.Bump)
	if err != nil {
		return err
	}
	slog.Info("Saving semver file locally.", "local_file", localPath, "old_version", srcVersion, "new_version", string(content))
	if err = aqfs.SaveFile(localPath, content); err != nil {
		return err
	}
	fmt.Println("v" + string(content))
	return nil
}

func incrementVersion(versionData []byte, incrementType string) ([]byte, error) {
	currentVer := strings.TrimSpace(string(versionData))
	parts := strings.Split(currentVer, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", currentVer)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	switch incrementType {
	case "none":
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return nil, fmt.Errorf("invalid increment type: %s", incrementType)
	}
	newVer := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return []byte(newVer), nil
}
