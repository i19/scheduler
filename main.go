package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"scheduler/cmd/server"
	"scheduler/cmd/swagger"
)

func main() {
	app := cli.NewApp()
	app.Usage = "Handle requests about custom linting rules  from scheduler"
	app.Commands = []*cli.Command{
		server.Server,
		swagger.Server,
	}
	if err := app.Run(os.Args); err != nil {
		panic(err.Error())
	}
}
