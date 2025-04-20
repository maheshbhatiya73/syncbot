package status

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/fatih/color"
	"syncbot/config"
)

func ShowStatus(c *cli.Context) error {
	cfg := config.GetConfig()
	fmt.Println(color.New(color.FgBlue).SprintFunc()("=== SyncBot Status ==="))
	fmt.Printf("Backup Path: %s\n", cfg.BackupPath)
	fmt.Printf("Destination: %s\n", cfg.Destination)
	fmt.Printf("Schedule: %s\n", cfg.BackupTime)
	fmt.Printf("Retention: %d days\n", cfg.RetentionDays)
	fmt.Printf("Compression: %s\n", cfg.Compression)
	fmt.Printf("Max Parallel: %d\n", cfg.MaxParallel)
	fmt.Printf("Exclude Patterns: %v\n", cfg.ExcludePatterns)
	fmt.Printf("Time Zone: %s\n", cfg.TimeZone)
	fmt.Printf("Status: %s\n", map[bool]string{true: "Running", false: "Stopped"}[cfg.IsRunning])
	return nil
}