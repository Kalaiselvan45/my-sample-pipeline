package impl

import (
	"log/slog"
	"strings"

	"github.com/aqfer/aqfer-go/lib/wp"
	"github.com/aqfer/aqfer-go/xlib/aqfs"
)

type WorkAreaCmd struct {
	Load       WorkAreaLoadCmd   `cmd:"" help:"Load saved content from work area"`
	Save       WorkAreaSaveCmd   `cmd:"" help:"Save build folder into work area"`
	Delete     WorkAreaDeleteCmd `cmd:"" help:"Delete work area folder in S3 at the end"`
	NumWorkers int               `help:"Number of workers to be used for saving/loading files" default:"10"`
}

type WorkAreaLoadCmd struct {
	Paths []string `required:"" help:"Paths to save/load, relative to home path"`
}

type WorkAreaSaveCmd struct {
	Paths []string `required:"" help:"Paths to save/load, relative to home path"`
}

type WorkAreaDeleteCmd struct{}

type CopySpec struct {
	src  string
	dest string
}

func deleteWorkArea(cli CLI) error {
	folder, err := aqfs.RemoveFolder(codeBuildInfo.getBatchBuildPath(), cli.WorkArea.NumWorkers)
	if err != nil {
		return err
	}
	slog.Info("Deleted files.", "num_files", folder.FileCount)
	return nil
}

func saveWorkArea(cli CLI) error {
	var cs []CopySpec
	removePrefix := cli.Home + "/"
	for _, path := range cli.WorkArea.Save.Paths {
		srcPath := cli.Home + "/" + path
		files, err := aqfs.ListFiles(srcPath)
		if err != nil {
			return err
		}
		for _, file := range files {
			filePath := strings.TrimPrefix(file.Name, removePrefix)
			destPath := codeBuildInfo.getBatchBuildPath(filePath)
			cs = append(cs, CopySpec{src: file.Name, dest: destPath})
		}
	}
	return copyFiles(cs, cli.WorkArea.NumWorkers)
}

func loadWorkArea(cli CLI) error {
	var cs []CopySpec
	removePrefix := codeBuildInfo.getBatchBuildPath() + "/"
	for _, path := range cli.WorkArea.Load.Paths {
		srcPath := codeBuildInfo.getBatchBuildPath(path)
		files, err := aqfs.ListFiles(srcPath)
		if err != nil {
			return err
		}
		for _, file := range files {
			filePath := strings.TrimPrefix(file.Name, removePrefix)
			destPath := cli.Home + "/" + filePath
			cs = append(cs, CopySpec{src: file.Name, dest: destPath})
		}
	}
	return copyFiles(cs, cli.WorkArea.NumWorkers)
}

func copyFiles(files []CopySpec, numWorkers int) error {
	workFn := func(cp CopySpec, _ int) (int, error) {
		return 0, aqfs.DoCopyFile(cp.src, cp.dest, false)
	}
	_, err := wp.Run[CopySpec, int](numWorkers, files, workFn, false, true)
	return err
}
