package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"sort"

	"github.com/urfave/cli/v2"

	ab "github.com/aqfer/aqfer/spark/aws-ci-cd/amdp-builder/cmd/ab/impl"
)

var Version string // to be initialized by the linker

var branchMap = map[string]func(version string){
	"ab": ab.Main,
}

func setVersion(app *cli.App, version string) {
	if version == "" {
		version = "unknown"
	}
	slog.Info("Command line details", "argument", os.Args[0], "version", version)
	app.Version = version
}

func startApp(app *cli.App) func(version string) {
	return func(version string) {
		setVersion(app, version)
		if err := app.Run(os.Args); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			os.Exit(2)
		}
	}
}

func main() {
	var action string
	if action == "uber-builder" {
		os.Args = os.Args[1:]
		action = os.Args[0]
	} else {
		action = path.Base(os.Args[0])
	}
	if f, found := branchMap[action]; found {
		f(Version)
	} else {
		actions := make([]string, 0, len(branchMap))
		for b := range branchMap {
			actions = append(actions, b)
		}
		sort.Slice(actions, func(i, j int) bool {
			return actions[i] < actions[j]
		})
		_, _ = fmt.Fprintf(os.Stderr, "%s is not a valid action. (valid actions are: %+v)\n", action, actions)
		os.Exit(1)
	}
}
