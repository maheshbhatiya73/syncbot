# 🔄 SyncBot

> **Automated Backup Tool for Linux Systems**  
> Backup your important files and directories with scheduled, reliable, and easy-to-use automation.

![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Linux-orange)

---

## 🧩 Features

- 🕒 **Scheduled Backups**: Configure backups to run at your desired time daily.
- 📁 **Custom Paths**: Set custom source and destination directories.
- 💾 **Compressed Archives**: Creates `.tar.gz` backups to save space.
- 🔧 **CLI Tool**: Intuitive CLI interface using [urfave/cli](https://github.com/urfave/cli).
- 🛠 **Persistent Config**: Settings saved in YAML (`/etc/syncbot/config.yaml`).

---

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/syncbot.git
cd syncbot
