package main

import (
	"fmt"
	"os"
	"github.com/urfave/cli/v2"
	"github.com/fatih/color"
	"syncbot/config"
	"syncbot/services/process"
	"syncbot/services/start"
	"syncbot/services/status"
)

const welcomeArt = `
   _____                  ____        _       
  / ____|                |  _ \      | |      
 | (___  _   _ _ __   ___| |_) | ___ | |_ 
  \___ \| | | | '_ \ / __|  _ < / _ \| __|
  ____) | |_| | | | | (__| |_) | (_) | |_
 |_____/ \__, |_| |_| \___|____/ \___/ \__|
          __/ |                            
         |___/

Welcome to SyncBot - Your Automated Backup Solution
`

func main() {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	config.SetupLogging()
	config.LoadConfig()

	app := &cli.App{
		Name:  "syncbot",
		Usage: "Advanced backup tool for Linux systems",
		Commands: []*cli.Command{
			{
				Name:  "set",
				Usage: "Set backup configuration",
				Flags: config.SetFlags(),
				Subcommands: []*cli.Command{
					{
						Name:  "list-timezones",
						Usage: "List available time zones (filter with keywords)",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "filter", Usage: "Filter time zones by keyword (e.g., America, Asia)"},
						},
						Action: config.ListTimeZones,
					},
				},
				Action: config.SetConfig,
			},
			{
				Name:   "start",
				Usage:  "Start backup service",
				Action: start.StartService,
			},
			{
				Name:   "stop",
				Usage:  "Stop backup service",
				Action: start.StopService,
			},
			{
				Name:   "status",
				Usage:  "Check backup status",
				Action: status.ShowStatus,
			},
			{
				Name:   "backup",
				Usage:  "Run backup immediately",
				Action: func(c *cli.Context) error {
					return process.RunBackup()
				},
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println(green(welcomeArt))
			fmt.Println(blue("Available commands:"))
			fmt.Println("  set                Configure backup settings")
			fmt.Println("  set list-timezones List available time zones")
			fmt.Println("  start              Start backup service")
			fmt.Println("  stop               Stop backup service")
			fmt.Println("  status             Show backup status")
			fmt.Println("  backup             Run backup immediately")
			fmt.Println("\nUse 'syncbot [command] --help' for more information")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		config.LogFatal(err)
	}
}