package impl

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aqfer/aqfer-go/xlib/aqfs"
)

type DownloadCmd struct {
	Source      string `arg:"" help:"Source url to download the artifact" required:""`
	Destination string `help:"Destination to download files from $HOME directory" default:"downloads/"`
	Ttl         string `help:"TTL to be applied on the cache file in case of non versioned files" default:"0s"`
}

//const downloadPath = "downloads"

func download(cli CLI) error {
	downloadPath := cli.Download.Destination
	outputName := ""
	if !strings.HasSuffix(downloadPath, "/") {
		outputName = filepath.Base(downloadPath)
		downloadPath = strings.TrimSuffix(downloadPath, "/"+outputName)
	} else {
		downloadPath = strings.TrimSuffix(downloadPath, "/")
	}
	s3DownloadPath := codeBuildInfo.getStoragePath(downloadPath)
	localBasePath := fmt.Sprintf("%s/%s", cli.Home, downloadPath)
	s3path, localPath, err := GeneratePaths(cli.Download.Source, localBasePath, outputName, s3DownloadPath)
	slog.Info("Checking file in cache...", "url", cli.Download.Source)
	if err != nil {
		return err
	}
	found, err := isCacheAvailable(s3path, cli.Download.Ttl)
	if err != nil {
		return err
	}
	if found {
		slog.Info("File in cache. Downloading from s3 cache...", "s3_path", s3path)
		if err = aqfs.DoCopyFile(s3path, localPath, false); err != nil {
			return err
		}
		slog.Info("Download completed.", "local_file", localPath)
		return nil
	}
	slog.Info("File not in cache or expired. Downloading file from url...", "url", cli.Download.Source)
	if err = DownloadFile(cli.Download.Source, localPath); err != nil {
		return err
	}
	slog.Info("Uploading to s3 cache...", "s3_path", s3path)
	if err = aqfs.DoCopyFile(localPath, s3path, false); err != nil {
		return err
	}
	slog.Info("Download completed.", "local_file", localPath)
	return nil
}

func isCacheAvailable(s3path, ttl string) (bool, error) {
	ch, err := aqfs.CheckFile(s3path)
	if err != nil || ch == nil {
		return false, err
	}
	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return false, err
	}
	return duration == 0 || ch.Time.Add(duration).After(time.Now()), nil
}

// GeneratePaths returns the S3 path and local path for a given URL
func GeneratePaths(inputURL, localBasePath, localFileName, s3DownloadPath string) (string, string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	// Retain domain and full path (without http/https)
	domainAndPath := parsedURL.Host + parsedURL.Path

	// Construct S3 path
	s3Path := fmt.Sprintf("%s/%s", s3DownloadPath, domainAndPath)

	// Extract only the filename for local storage
	if localFileName == "" {
		localFileName = filepath.Base(parsedURL.Path)
	}

	// Construct local path
	localPath := fmt.Sprintf("%s/%s", localBasePath, localFileName)

	return s3Path, localPath, nil
}

// DownloadFile downloads a file from a given URL and saves it to a local file
func DownloadFile(url, outputPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response: %d %s", resp.StatusCode, resp.Status)
	}

	// Create the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(outFile)

	// Copy the response body to the file
	if _, err = io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
