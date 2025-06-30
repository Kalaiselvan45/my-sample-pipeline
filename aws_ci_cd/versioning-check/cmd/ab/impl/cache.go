package impl

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/aqfer/aqfer-go/lib/wp"
)

type GoCacheCmd struct {
	Get GetGoCacheCmd `cmd:"" help:"gets go cache from s3"`
	Set SetGoCacheCmd `cmd:"" help:"moves go cache to s3"`
}

type GetGoCacheCmd struct {
	Path string `help:"Optional: Specifies the path to the cache directory in s3." short:"p"`
}

type SetGoCacheCmd struct {
}

func setGoCache(cli CLI) (err error) {
	if !strings.HasSuffix(codeBuildInfo.ProjectName, "go-cache") || codeBuildInfo.ProjectName == "go-cache" {
		return fmt.Errorf(`ab go-cache set command requires '-go-cache' suffix in the project name: %s`, codeBuildInfo.ProjectName)
	}
	// cd $HOME
	if err := os.Chdir(cli.Home); err != nil {
		return fmt.Errorf("Error changing to %s: %v\n", cli.Home, err)
	}
	cmds := []string{
		"tar -czf go.tar.gz go",
		"tar -cf go-build.tar go-build",
	}
	doWork := func(cmd string, index int) (int, error) {
		slog.Info("Building tar files.", "cmd", cmd)
		goTar := exec.Command("bash", "-c", cmd)
		goTar.Stderr = os.Stderr
		if err := goTar.Run(); err != nil {
			return 0, fmt.Errorf("Error creating %s: %v\n", cmd, err)
		}
		return 0, nil
	}
	if _, err := wp.Run(2, cmds, doWork, false, false); err != nil {
		return err
	}
	s3Folder := strings.TrimSuffix(codeBuildInfo.ProjectName, "-go-cache")
	destination := fmt.Sprintf("s3://codebuild-%s-%s/%s/cache/", codeBuildInfo.Region, codeBuildInfo.AccountID,
		s3Folder)
	slog.Info("Copying go.tar.gz and go-build.tar to s3.", "destination", destination)

	return copyFiles([]CopySpec{{src: "go.tar.gz", dest: destination + "go.tar.gz"},
		{src: "go-build.tar", dest: destination + "go-build.tar"}}, 2)
}

func getGoCache(cli CLI) error {
	path := cli.GoCache.Get.Path
	if path == "" {
		path = codeBuildInfo.getStoragePath("cache")
	}
	if !strings.HasPrefix(path, "s3://") {
		return fmt.Errorf("%s is not a s3 path", path)
	}
	cmds := []string{
		fmt.Sprintf("aws s3 cp %s/go.tar.gz - | tar -C %s -xz", path, cli.Home),
		fmt.Sprintf("aws s3 cp %s/go-build.tar - | tar -C %s -x", path, cli.Home),
	}

	doWork := func(cmd string, index int) (int, error) {
		slog.Info("Executing s3 cmd", "cmd", cmd)
		cmdOut := exec.Command("bash", "-c", cmd)
		cmdOut.Stdout = os.Stdout
		cmdOut.Stderr = os.Stderr

		if err := cmdOut.Run(); err != nil {
			return 0, fmt.Errorf("task %d failed for %s: %v", index, cmd, err)
		}
		return 0, nil
	}

	if _, err := wp.Run(2, cmds, doWork, false, false); err != nil {
		return err
	}
	return nil
}
