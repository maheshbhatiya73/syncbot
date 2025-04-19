package main

import (
        "fmt"
        "os"
        "path/filepath"
        "time"
        "github.com/urfave/cli/v2"
        "github.com/fatih/color"
        "gopkg.in/yaml.v2"
        "os/exec"
        "log"
)

type Config struct {
        BackupPath  string `yaml:"backup_path"`
        BackupTime  string `yaml:"backup_time"`
        Destination string `yaml:"destination"`
        IsRunning   bool   `yaml:"is_running"`
}

var config Config
var configFile = "/etc/syncbot/config.yaml"

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
        red := color.New(color.FgRed).SprintFunc()

        loadConfig()

        app := &cli.App{
                Name:  "syncbot",
                Usage: "Automated backup tool for Linux systems",
                Commands: []*cli.Command{
                        {
                                Name:  "set",
                                Usage: "Set backup configuration",
                                Flags: []cli.Flag{
                                        &cli.StringFlag{
                                                Name:  "path",
                                                Usage: "Set backup source path",
                                        },
                                        &cli.StringFlag{
                                                Name:  "dest",
                                                Usage: "Set backup destination path",
                                        },
                                        &cli.StringFlag{
                                                Name:  "time",
                                                Usage: "Set backup time (HH:MM)",
                                        },
                                },
                                Action: func(c *cli.Context) error {
                                        if c.String("path") != "" {
                                                if _, err := os.Stat(c.String("path")); os.IsNotExist(err) {
                                                        return fmt.Errorf("%s Directory does not exist: %s", red("✗"), c.String("path"))
                                                }
                                                config.BackupPath = c.String("path")
                                                fmt.Printf("%s Backup path set successfully: %s\n", green("✓"), config.BackupPath)
                                        }
                                        if c.String("dest") != "" {
                                                config.Destination = c.String("dest")
                                                fmt.Printf("%s Destination path set to: %s\n", green("✓"), config.Destination)
                                        }
                                        if c.String("time") != "" {
                                                config.BackupTime = c.String("time")
                                                fmt.Printf("%s Backup time set to: %s\n", green("✓"), config.BackupTime)
                                        }
                                        saveConfig()
                                        return nil
                                },
                        },
                        {
                                Name:  "start",
                                Usage: "Start backup service",
                                Action: func(c *cli.Context) error {
                                        if config.BackupPath == "" || config.Destination == "" {
                                                return fmt.Errorf("%s Please set backup path and destination first", red("✗"))
                                        }
                                        config.IsRunning = true
                                        saveConfig()
                                        fmt.Printf("%s Backup service started\n", green("✓"))
                                        go performBackup()
                                        return nil
                                },
                        },
                        {
                                Name:  "stop",
                                Usage: "Stop backup service",
                                Action: func(c *cli.Context) error {
                                        config.IsRunning = false
                                        saveConfig()
                                        fmt.Printf("%s Backup service stopped\n", green("✓"))
                                        return nil
                                },
                        },
                        {
                                Name:  "status",
                                Usage: "Check backup status",
                                Action: func(c *cli.Context) error {
                                        fmt.Println(blue("=== SyncBot Status ==="))
                                        fmt.Printf("Backup Path: %s\n", config.BackupPath)
                                        fmt.Printf("Destination: %s\n", config.Destination)
                                        fmt.Printf("Schedule: %s\n", config.BackupTime)
                                        fmt.Printf("Status: %s\n", map[bool]string{true: "Running", false: "Stopped"}[config.IsRunning])
                                        return nil
                                },
                        },
                },
                Action: func(c *cli.Context) error {
                        fmt.Println(green(welcomeArt))
                        fmt.Println(blue("Available commands:"))
                        fmt.Println("  set     Configure backup settings")
                        fmt.Println("  start   Start backup service")
                        fmt.Println("  stop    Stop backup service")
                        fmt.Println("  status  Show backup status")
                        fmt.Println("\nUse 'syncbot [command] --help' for more information")
                        return nil
                },
        }

        err := app.Run(os.Args)
        if err != nil {
                log.Fatal(err)
        }
}

func loadConfig() {
        os.MkdirAll("/etc/syncbot", 0755)

        data, err := os.ReadFile(configFile)
        if err != nil {
                config = Config{
                        Destination: "/backup/syncbot",
                        IsRunning:   false,
                }
                saveConfig()
                return
        }
        yaml.Unmarshal(data, &config)
}

func saveConfig() {
        data, err := yaml.Marshal(&config)
        if err != nil {
                log.Fatal(err)
        }
        err = os.WriteFile(configFile, data, 0644)
        if err != nil {
                log.Fatal(err)
        }
}

func performBackup() {
        for config.IsRunning {
                now := time.Now()
                backupTime, _ := time.Parse("15:04", config.BackupTime)

                if now.Hour() == backupTime.Hour() && now.Minute() == backupTime.Minute() {
                        backupName := fmt.Sprintf("backup_%s.tar.gz", now.Format("20060102_150405"))
                        backupPath := filepath.Join(config.Destination, backupName)

                        os.MkdirAll(config.Destination, 0755)

                        cmd := exec.Command("tar", "-zcf", backupPath, "-C", config.BackupPath, ".")
                        output, err := cmd.CombinedOutput()
                        if err != nil {
                                log.Printf("Backup failed: %v\n%s", err, string(output))
                        } else {
                                log.Printf("Backup completed successfully: %s", backupPath)
                        }
                }
				
                time.Sleep(time.Minute)
        }
}