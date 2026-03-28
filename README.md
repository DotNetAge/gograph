<div align="center">
  <h1>GoGraph</h1>
  <p>The Minimalist Embedded Graph Database in Pure Go</p>

  [![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/gograph.svg)](https://pkg.go.dev/github.com/DotNetAge/gograph)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/DotNetAge/gograph)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/DotNetAge/gograph)](https://goreportcard.com/report/github.com/DotNetAge/gograph)
  [![Docs](https://img.shields.io/badge/docs-gograph.rayainfo.cn-2094f3.svg)](https://gograph.rayainfo.cn)

  <p>
    <a href="README_zh.md">简体中文</a> | <strong>English</strong>
  </p>
</div>

---

## 📖 Project Overview

**GoGraph** is a lightweight, zero-dependency, embedded graph database written entirely in Go. Think of it as **"SQLite for Graph Databases"**. 

It allows Go developers to execute Cypher queries (the standard graph query language) and manage local graph data—nodes, relationships, and properties—without the overhead of external heavy database services like Neo4j.

## ⚡ Quick Start

### 1. Install CLI (Recommended)
The fastest way to explore GoGraph is via the CLI.

**macOS / Linux (Homebrew):**
```bash
brew install dotnetage/tap/gograph
```

**Run TUI (Interactive Shell):**
```bash
# Simply run without arguments to open default.db in interactive mode
gograph
```

### 2. Use as a Go Library
Add GoGraph to your project:
```bash
go get github.com/DotNetAge/gograph
```

**Basic Example:**
```go
package main

import (
	"context"
	"fmt"
	"github.com/DotNetAge/gograph/pkg/api"
)

func main() {
	db, _ := api.Open("default.db")
	defer db.Close()

	ctx := context.Background()
	db.Exec(ctx, "CREATE (a:User {name: 'Alice'})-[:KNOWS]->(b:User {name: 'Bob'})")
	
	rows, _ := db.Query(ctx, "MATCH (u:User) RETURN u.name")
	defer rows.Close()
	
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Println("User:", name)
	}
}
```

## ✨ Key Features

- 🚀 **Pure Go**: No CGO, seamless cross-platform support.
- 📦 **Embedded**: Zero-config, single-directory storage (Pebble DB).
- 🔍 **Cypher Support**: Native `MATCH`, `CREATE`, `SET`, `DELETE`.
- 🛡️ **ACID**: MVCC, thread-safety, and WAL recovery.
- 🛠️ **TUI Included**: Interactive shell with auto-completion and ASCII tables.

## 💻 CLI Usage

The `gograph` binary provides a powerful TUI and command-line utilities.

| Command                  | Description                                      |
| ------------------------ | ------------------------------------------------ |
| `gograph`                | Launch Interactive TUI (default to `default.db`) |
| `gograph query <cypher>` | Run a read-only query                            |
| `gograph exec <cypher>`  | Run a data modification command                  |

**Example:**
```bash
gograph query "MATCH (n) RETURN n"
```

---

## 🧩 System Compatibility
- **OS**: macOS, Linux, Windows. 
- **Arch**: `amd64`, `arm64`.

## 📚 Documentation
Check out the full [Documentation](https://gograph.rayainfo.cn) or the [Docs Folder](./docs/README.md).
