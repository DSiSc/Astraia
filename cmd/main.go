package main

import (
	"fmt"
	local "github.com/DSiSc/astraia/utils"
	"github.com/DSiSc/wallet/cmd"
	"github.com/DSiSc/wallet/utils"
	"github.com/urfave/cli"
	"os"
	"sort"
)

const (
	clientIdentifier = "geth" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = local.NewApp(gitCommit, "the ligntClient command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.PasswordFileFlag,
		utils.LightKDFFlag,
	}

	rpcFlags = []cli.Flag{ }
	whisperFlags = []cli.Flag{ }
	metricsFlags = []cli.Flag{ }

)

func init() {
	app.Action = astraia
	app.HideVersion = true
	app.Copyright = "Copyright 2018-2023 The justitia Authors"

	app.Commands = []cli.Command{
		consoleCommand,
	}
	app.Commands = append(app.Commands, cmd.AccountCommand)
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)

	app.Before = func(ctx *cli.Context) error {
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func astraia(ctx *cli.Context) error {
	fmt.Println("***Usage")
	fmt.Println("***astraia account new/update/import/list --datadir --keystore")
	return nil
}

