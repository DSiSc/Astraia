package main

import (
	"fmt"
	local "github.com/DSiSc/lightClient/utils"
	"github.com/DSiSc/wallet/cmd"
	"github.com/DSiSc/wallet/utils"
	"github.com/urfave/cli"
	"os"
	"sort"
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
	app.Action = lightClient
	app.HideVersion = true
	app.Copyright = "Copyright 2018-2023 The justitia Authors"
	//accountCmdType := reflect.TypeOf(cmd.AccountCommand)
	//fmt.Println(accountCmdType)

	app.Commands = append(app.Commands, cmd.AccountCommand)

	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)

	app.Before = func(ctx *cli.Context) error {
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		//debug.Exit()
		//console.Stdin.Close()
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func lightClient(ctx *cli.Context) error {
	fmt.Println("***Usage")
	fmt.Println("***lightClient account new/update/import/list --datadir --keystore")
	return nil
}

