package impl

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Init         InitCmd          `cmd:"" help:"Initializes home, git and docker. Returns export script, which should be evaluated using eval"`
	GitCommit    CommitCmd        `cmd:"" help:"Check latest commit and skips build if commit is same as old"`
	GoCache      GoCacheCmd       `cmd:"" help:"Handle set and get of go cache"`
	Semver       SemverCmd        `cmd:"" help:"Handle semantic versions"`
	Docker       DockerCmd        `cmd:"" help:"Handle docker related actions"`
	Download     DownloadCmd      `cmd:"" help:"Downloads given url and cache them in S3"`
	WorkArea     WorkAreaCmd      `cmd:"" help:"Workarea command can be used for storing/loading files between batch builds"`
	UploadConfig UploadConfigCmd  `cmd:"" help:"Uploads the file to the given S3 path"`
	Home         string           `help:"Location to store any data" env:"HOME" required:""`
	Verbose      bool             `help:"Enable verbose logging" short:"v"`
	Version      kong.VersionFlag `help:"Show version and exit"`
}

func runCommands(v string) (err error) {
	codeBuildInfo, err = ExtractInfo()
	if err != nil {
		return err
	}
	var cli CLI
	ctx := kong.Parse(&cli, kong.Vars{"version": v})
	switch ctx.Command() {
	case "init":
		return initCmd(cli)
	case "git-commit get":
		return getCommit(cli)
	case "git-commit put":
		return putCommit(cli)
	case "download <source>":
		return download(cli)
	case "semver get":
		return getSemver(cli)
	case "semver put":
		return putSemver(cli)
	case "work-area load":
		return loadWorkArea(cli)
	case "work-area save":
		return saveWorkArea(cli)
	case "work-area delete":
		return deleteWorkArea(cli)
	case "go-cache get":
		return getGoCache(cli)
	case "go-cache set":
		return setGoCache(cli)
	case "docker create-manifest":
		return createManifest(cli)
	case "upload-config":
		return uploadConfig(cli)
	default:
		return fmt.Errorf("unknown command: %v", ctx.Command())
	}
}

func Main(v string) {
	if err := runCommands(v); err != nil {
		log.Fatal(err)
	}
}

func run(args ...string) error {
	slog.Info("Running command...", "command", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	// Redirect stdout and stderr to the parent process
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run the command
	return cmd.Run()
}
