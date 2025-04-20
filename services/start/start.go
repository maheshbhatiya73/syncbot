package start

import (
	"fmt"
	"time"
	"github.com/urfave/cli/v2"
	"github.com/fatih/color"
	"syncbot/config"
	"syncbot/services/process"
	"log"
)

func StartService(c *cli.Context) error {
	cfg := config.GetConfig()
	if cfg.BackupPath == "" || cfg.Destination == "" {
		return fmt.Errorf("%s Please set backup path and destination first", color.New(color.FgRed).SprintFunc()("✗"))
	}
	cfg.IsRunning = true
	config.SaveConfig()
	fmt.Printf("%s Backup service started\n", color.New(color.FgGreen).SprintFunc()("✓"))
	go performBackup()
	return nil
}

func StopService(c *cli.Context) error {
	cfg := config.GetConfig()
	cfg.IsRunning = false
	config.SaveConfig()
	fmt.Printf("%s Backup service stopped\n", color.New(color.FgGreen).SprintFunc()("✓"))
	return nil
}

func performBackup() {
	cfg := config.GetConfig()
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		config.LogFatal(fmt.Errorf("failed to load time zone %s: %v", cfg.TimeZone, err))
	}

	for cfg.IsRunning {
		now := time.Now().In(loc)
		backupTime, err := time.ParseInLocation("15:04", cfg.BackupTime, loc)
		if err != nil {
			config.LogFatal(fmt.Errorf("failed to parse backup time %s: %v", cfg.BackupTime, err))
		}

		if now.Hour() == backupTime.Hour() && now.Minute() == backupTime.Minute() {
			if err := process.RunBackup(); err != nil {
				log.Printf("Backup failed: %v", err)
			} else {
				process.CleanupOldBackups()
			}
		}

		time.Sleep(time.Minute)
	}
}