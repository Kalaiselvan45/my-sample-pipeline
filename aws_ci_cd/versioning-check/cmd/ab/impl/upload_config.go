package impl

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/aqfer/aqfer-go/xlib/aqfs"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type UploadConfigCmd struct {
	SourceFilePath string `name:"src" help:"source file path to upload"`
	DestPrefix     string `name:"dst" help:"S3 prefix to upload the file under (e.g. s3://com.aqfer.preprod.config/config/analytics-engine/preprod)"`
	Gzip           bool   `name:"gzip" help:"If true, gzip the file before uploading"`
	Versioned      bool   `name:"versioned" help:"If true, hashes each file and uploads under versioned path"`
}

func md5File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func uploadToS3(srcPath, dstPath string) error {
	cmd := exec.Command("aws", "s3", "cp", srcPath, dstPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func uploadConfig(cli CLI) error {
	var (
		version string
		err     error
	)

	if cli.UploadConfig.Versioned {
		version, err = md5File(cli.UploadConfig.SourceFilePath)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %v", cli.UploadConfig.SourceFilePath, err)
		}
	}

	srcPath := cli.UploadConfig.SourceFilePath
	dstPath := aqfs.MultiPathJoin(false, cli.UploadConfig.DestPrefix, version, filepath.Base(cli.UploadConfig.SourceFilePath))

	if cli.UploadConfig.Gzip {
		gzipCmd := exec.Command("gzip", "-f", cli.UploadConfig.SourceFilePath)
		gzipCmd.Stdout = os.Stdout
		gzipCmd.Stderr = os.Stderr
		if err := gzipCmd.Run(); err != nil {
			return err
		}
		srcPath += ".gz"
		dstPath += ".gz"
	}

	if err := uploadToS3(srcPath, dstPath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf(`{"version": "%s"}`, version)
	return nil
}
