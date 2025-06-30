package impl

import (
	"fmt"
	"os"

	"github.com/aqfer/aqfer-go/xlib/aqfs"
)

type InitCmd struct {
	KnownHosts string   `help:"CI_GITHUB_SSH_RSA used for known hosts" env:"CI_GITHUB_SSH_RSA" required:""`
	IdDsa      string   `help:"CI_GITHUB_SSH_PRIVATE_KEY key used accessing the github" env:"CI_GITHUB_SSH_PRIVATE_KEY" required:""`
	Mkdirs     []string `name:"mkdirs" help:"List of directories to create."`
}

var envVars = map[string]string{
	"PS4":           `'+ $(date "+%Y/%m/%d %H:%M:%S.%6N") '`,
	"DOCKER_CONFIG": `$HOME/.docker`,
	"GOPATH":        `$HOME/go`,
	"GOCACHE":       `$HOME/go-build`,
	"GOEXPERIMENT":  `nocoverageredesign`,
	"GOOS":          `linux`,
	"CGO_ENABLED":   `0`,
	"GOPRIVATE":	 `github.com/aqfer/*`,
}

const gitConfigFileContent = `[url "git@github.com:"]
    insteadof = https://github.com/
`

func createMultipleDirs(basePath string, dirs []string, perm os.FileMode) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(fmt.Sprintf("%s/%s", basePath, dir), perm); err != nil {
			return err
		}
	}
	return nil
}

func printEnvVars() {
	for key, value := range envVars {
		fmt.Printf("export %s=%s\n", key, value)
	}
}

func initCmd(cli CLI) error {
	err := os.Mkdir(cli.Home, 0o700)
	if err != nil {
		return err
	}
	if err = createMultipleDirs(cli.Home, cli.Init.Mkdirs, 0o700); err != nil {
		return err
	}
	if err = gitInit(cli); err != nil {
		return err
	}
	envVars["GIT_SSH_COMMAND"] = fmt.Sprintf(`"ssh -i %s/.ssh/id_dsa -o UserKnownHostsFile=%s/.ssh/known_hosts"`, cli.Home, cli.Home)
	printEnvVars()
	return nil
}

func gitInit(cli CLI) error {
	sshDir := fmt.Sprintf("%s/.ssh", cli.Home)
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return err
	}

	khFile := fmt.Sprintf("%s/.ssh/known_hosts", cli.Home)
	if err := aqfs.SaveFile(khFile, []byte(cli.Init.KnownHosts+"\n")); err != nil {
		return err
	}
	if err := os.Chmod(khFile, 0o600); err != nil {
		return err
	}

	idFile := fmt.Sprintf("%s/.ssh/id_dsa", cli.Home)
	if err := aqfs.SaveFile(idFile, []byte(cli.Init.IdDsa+"\n")); err != nil {
		return err
	}
	if err := os.Chmod(idFile, 0o600); err != nil {
		return err
	}

	gitConfigFile := fmt.Sprintf("%s/.gitconfig", cli.Home)
	return aqfs.SaveFile(gitConfigFile, []byte(gitConfigFileContent))
}
