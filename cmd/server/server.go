package server

import (
	"github.com/urfave/cli/v2"

	"scheduler/internal/api/router"
	"scheduler/internal/config"
	"scheduler/internal/platform/database"
	"scheduler/internal/platform/etcd"
	"scheduler/internal/platform/worker"
	"scheduler/pkg/scheduler"
)

var Server = &cli.Command{
	Name:  "server",
	Usage: "start scheduler server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "conf, c",
			Usage: "path to config file",
			Value: "./config.yaml",
		},
	},
	Action: func(cCtx *cli.Context) error {
		config.Init(cCtx.String("conf"))
		database.Init(
			config.Config.Database.MySQLHost,
			config.Config.Database.MySQLPort,
			config.Config.Database.MySQLDB,
			config.Config.Database.MySQLUser,
			config.Config.Database.MySQLPassword,
		)
		etcd.Init(config.Config.EtcdHosts)
		worker.Init(config.Config.EtcdWorkerPrefix)
		scheduler.Init(config.Config.Kinds, config.Config.Priorities)
		router.Run(config.Config.Port)
		return nil
	},
}
