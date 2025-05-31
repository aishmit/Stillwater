# 📂 Stillwater

A small Go-based daemon that helps you **organize, classify, and clean up your Downloads folder** over time. Designed to run continuously and support long-term file hygiene with user-defined rules.

---

## 🔧 Features

- 🕵️‍♂️ **Watches your Downloads folder** continuously
- 🧠 **Tracks file changes** and resets classification if files are modified
- 📌 **Prompts classification** of new files:
  - `0` — Unclassified
  - `1` — Never archive
  - `2` — Archive after 60 days
  - `3` — Delete after 60 days
- 📅 **Stores file metadata** in a local database (PostgreSQL or SQLite)
- 🧹 **Scheduled cleanup**: deletes or archives files automatically

---

## 🚀 Getting Started

### 1. Clone the Repo

```bash
git clone https://github.com/your-username/downloads-lifecycle.git
cd downloads-lifecycle
```

### 2. Install Dependencies

Make sure Go is installed (≥ 1.18), then initialize modules:

```bash
go mod tidy
```

### 3. Build and Run

```bash
go run main.go
```

## 📌 Roadmap

- [ ] Interactive CLI prompts to classify new files

- [ ] Command-line downloadsctl tool for manual classification

- [ ] Archive directory configuration

- [ ] Integration with macOS LaunchAgents or systemd

- [ ] .downloadsignore support to exclude certain files

- [ ] GUI (TUI or Electron)
