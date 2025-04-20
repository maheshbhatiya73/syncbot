package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/urfave/cli/v2"
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
	"log"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

type Config struct {
	BackupPath      string   `yaml:"backup_path"`
	BackupTime      string   `yaml:"backup_time"`
	Destination     string   `yaml:"destination"`
	IsRunning       bool     `yaml:"is_running"`
	RetentionDays   int      `yaml:"retention_days"`
	Compression     string   `yaml:"compression"`
	MaxParallel     int      `yaml:"max_parallel"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	TimeZone        string   `yaml:"time_zone"`
}

var config Config
var configFile = "/etc/syncbot/config.yaml"
var logFile = "/var/log/syncbot/backup.log"

var availableTimeZones = []string{
	"UTC",
	"America/New_York",
	"America/Los_Angeles",
	"America/Chicago",
	"America/Denver",
	"America/Phoenix",
	"America/Anchorage",
	"America/Honolulu",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Moscow",
	"Asia/Tokyo",
	"Asia/Shanghai",
	"Asia/Kolkata",
	"Asia/Dubai",
	"Asia/Singapore",
	"Australia/Sydney",
	"Australia/Melbourne",
	"Pacific/Auckland",
}

func SetupLogging() {
	logPath := filepath.Dir(logFile)
	if err := os.MkdirAll(logPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	writer, err := rotatelogs.New(
		logFile+".%Y%m%d",
		rotatelogs.WithLinkName(logFile),
		rotatelogs.WithMaxAge(30*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func getSystemTimeZone() string {
	if loc := time.Local.String(); loc != "Local" {
		return loc
	}
	if tz, err := os.Readlink("/etc/localtime"); err == nil {
		if strings.HasPrefix(tz, "/usr/share/zoneinfo/") {
			return strings.TrimPrefix(tz, "/usr/share/zoneinfo/")
		}
	}
	return "UTC"
}

func LoadConfig() {
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		config = Config{
			Destination:     "/backup/syncbot",
			IsRunning:       false,
			RetentionDays:   7,
			Compression:     "gzip",
			MaxParallel:     1,
			ExcludePatterns: []string{},
			TimeZone:        getSystemTimeZone(),
		}
		SaveConfig()
		return
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	if config.TimeZone == "" {
		config.TimeZone = getSystemTimeZone()
		SaveConfig()
	}
}

func SaveConfig() {
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Fatalf("Failed to write config: %v", err)
	}
}

func GetConfig() *Config {
	return &config
}

func LogFatal(err error) {
	log.Fatal(err)
}

func SetFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "path", Usage: "Set backup source path"},
		&cli.StringFlag{Name: "dest", Usage: "Set backup destination path"},
		&cli.StringFlag{Name: "time", Usage: "Set backup time (HH:MM)"},
		&cli.IntFlag{Name: "retention", Usage: "Set retention days"},
		&cli.StringFlag{Name: "compression", Usage: "Set compression (gzip, bzip2, xz)"},
		&cli.IntFlag{Name: "parallel", Usage: "Set max parallel backups"},
		&cli.StringSliceFlag{Name: "exclude", Usage: "Exclude patterns"},
		&cli.StringFlag{Name: "timezone", Usage: "Set time zone (e.g., America/New_York)"},
	}
}

func SetConfig(c *cli.Context) error {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// Track if any changes were made
	changed := false

	if c.String("path") != "" {
		if _, err := os.Stat(c.String("path")); os.IsNotExist(err) {
			return fmt.Errorf("%s Directory does not exist: %s", red("✗"), c.String("path"))
		}
		config.BackupPath = c.String("path")
		log.Printf("Backup path set: %s", config.BackupPath)
		changed = true
	}
	if c.String("dest") != "" {
		config.Destination = c.String("dest")
		log.Printf("Destination set: %s", config.Destination)
		changed = true
	}
	if c.String("time") != "" {
		if _, err := time.Parse("15:04", c.String("time")); err != nil {
			return fmt.Errorf("%s Invalid time format (use HH:MM): %s", red("✗"), c.String("time"))
		}
		config.BackupTime = c.String("time")
		log.Printf("Backup time set: %s", config.BackupTime)
		changed = true
	}
	if c.Int("retention") > 0 {
		config.RetentionDays = c.Int("retention")
		log.Printf("Retention set: %d days", config.RetentionDays)
		changed = true
	}
	if c.String("compression") != "" {
		if !isValidCompression(c.String("compression")) {
			return fmt.Errorf("%s Invalid compression (use gzip, bzip2, xz): %s", red("✗"), c.String("compression"))
		}
		config.Compression = c.String("compression")
		log.Printf("Compression set: %s", config.Compression)
		changed = true
	}
	if c.Int("parallel") > 0 {
		config.MaxParallel = c.Int("parallel")
		log.Printf("Max parallel set: %d", config.MaxParallel)
		changed = true
	}
	if len(c.StringSlice("exclude")) > 0 {
		config.ExcludePatterns = c.StringSlice("exclude")
		log.Printf("Exclude patterns set: %v", config.ExcludePatterns)
		changed = true
	}
	if c.String("timezone") != "" {
		if _, err := time.LoadLocation(c.String("timezone")); err != nil {
			return fmt.Errorf("%s Invalid time zone: %s", red("✗"), c.String("timezone"))
		}
		config.TimeZone = c.String("timezone")
		log.Printf("Time zone set: %s", config.TimeZone)
		changed = true
	}

	if changed {
		SaveConfig()
		fmt.Printf("%s Backup configuration updated successfully\n", green("✓"))
		fmt.Println(color.New(color.FgBlue).SprintFunc()("=== Current Configuration ==="))
		fmt.Printf("Backup Path: %s\n", config.BackupPath)
		fmt.Printf("Destination: %s\n", config.Destination)
		fmt.Printf("Schedule: %s\n", config.BackupTime)
		fmt.Printf("Retention: %d days\n", config.RetentionDays)
		fmt.Printf("Compression: %s\n", config.Compression)
		fmt.Printf("Max Parallel: %d\n", config.MaxParallel)
		fmt.Printf("Exclude Patterns: %v\n", config.ExcludePatterns)
		fmt.Printf("Time Zone: %s\n", config.TimeZone)
	} else {
		fmt.Printf("%s No changes made to configuration\n", red("✗"))
	}

	return nil
}

func ListTimeZones(c *cli.Context) error {
	filter := strings.ToLower(c.String("filter"))
	fmt.Println("Available time zones:")
	for _, tz := range availableTimeZones {
		if filter == "" || strings.Contains(strings.ToLower(tz), filter) {
			fmt.Printf("  %s\n", tz)
		}
	}
	return nil
}

func isValidCompression(comp string) bool {
	return comp == "gzip" || comp == "bzip2" || comp == "xz"
}