<div align="center">
  <h1>CypherDB</h1>
  <p>The Minimalist Embedded Graph Database in Pure Go with Python bindings</p>

  [![PyPI version](https://badge.fury.io/py/cypherdb.svg)](https://pypi.org/project/cypherdb/)
  [![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/gograph.svg)](https://pkg.go.dev/github.com/DotNetAge/gograph)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/DotNetAge/gograph)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/DotNetAge/gograph)](https://goreportcard.com/report/github.com/DotNetAge/gograph)

</div>

---

## 📖 Project Overview

**CypherDB** (formerly GoGraph) is a lightweight, zero-dependency, embedded graph database written entirely in Go. Think of it as **"SQLite for Graph Databases"**.

It allows Go and Python developers to execute Cypher queries (the standard graph query language) and manage local graph data—nodes, relationships, and properties—without the overhead of external heavy database services like Neo4j.

## ⚡ Quick Start

### 1. Install Python Package

```bash
pip install cypherdb
```

**Basic Example:**

```python
import cypherdb

# Create database
db = cypherdb.Database("my_graph.db")

# Use transaction with context manager
with db.transaction() as tx:
    node_id = tx.create_node("Person", {"name": "Alice"})
    print(f"Created node: {node_id}")
    # Auto-commit

# Manual transaction
tx = db.transaction()
try:
    tx.create_node("Person", {"name": "Bob"})
    tx.commit()
except:
    tx.rollback()
```

### 2. Install CLI (Go)

The fastest way to explore CypherDB is via the CLI.

**macOS / Linux (Homebrew):**

```bash
brew tap dotnetage/gograph
brew install gograph
```

**Run TUI (Interactive Shell):**

```bash
# Simply run without arguments to open default.db in interactive mode
gograph
```

## ✨ Key Features

- 🚀 **Pure Go**: No CGO, seamless cross-platform support.
- 🐍 **Python Bindings**: Easy integration with Python applications.
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
- **Python**: 3.9+

## 📚 Documentation
Check out the full [Documentation](https://gograph.rayainfo.cn) or the [Docs Folder](./docs/README.md).
