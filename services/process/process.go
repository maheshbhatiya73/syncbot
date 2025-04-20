package process

import (
	"io"
	"os"
	"path/filepath"
	"time"
	"os/exec"
	"github.com/pkg/errors"
	"log"
	"syncbot/config"
	"fmt"
)

func RunBackup() error {
	cfg := config.GetConfig()
	backupName := fmt.Sprintf("backup_%s.tar.%s", time.Now().Format("20060102_150405"), cfg.Compression)
	backupPath := filepath.Join(cfg.Destination, backupName)

	if err := os.MkdirAll(cfg.Destination, 0755); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	args := []string{"-c"}
	switch cfg.Compression {
	case "gzip":
		args = append(args, "-z")
	case "bzip2":
		args = append(args, "-j")
	case "xz":
		args = append(args, "-J")
	}

	args = append(args, "-f", backupPath, "-C", cfg.BackupPath, ".")

	for _, pattern := range cfg.ExcludePatterns {
		args = append(args, "--exclude", pattern)
	}

	cmd := exec.Command("tar", args...)
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start backup")
	}

	errOutput, _ := io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		return errors.Wrapf(err, "backup failed: %s", string(errOutput))
	}

	log.Printf("Backup completed successfully: %s", backupPath)
	return nil
}

func CleanupOldBackups() {
	cfg := config.GetConfig()
	cutoff := time.Now().AddDate(0, 0, -cfg.RetentionDays)
	files, err := os.ReadDir(cfg.Destination)
	if err != nil {
		log.Printf("Failed to read backup directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(cfg.Destination, file.Name())); err != nil {
				log.Printf("Failed to remove old backup %s: %v", file.Name(), err)
			} else {
				log.Printf("Removed old backup: %s", file.Name())
			}
		}
	}
}