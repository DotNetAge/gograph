<div align="center">
  <h1>GoGraph</h1>
  <p>The Minimalist Embedded Graph Database in Pure Go</p>
  <p>
    <a href="README.md">中文</a> | <strong>English</strong>
  </p>
</div>

---

## 📖 Project Overview

**GoGraph** is a lightweight, zero-dependency, embedded graph database written entirely in Go. Think of it as **"SQLite for Graph Databases"**. 

It allows Go developers to execute Cypher queries (the standard graph query language) and manage local graph data—nodes, relationships, and properties—without the overhead of deploying, managing, or connecting to external heavy database services like Neo4j or NebulaGraph.

With GoGraph, you get the expressive power of a graph data model combined with the simplicity and performance of local single-file storage.

## ✨ Key Features

- 🚀 **Pure Go Implementation**: No CGO dependencies required, ensuring seamless cross-platform compilation and deployment.
- 📦 **Minimalist & Embedded**: Runs locally via a single directory (powered by Pebble DB). Simply `go get` and start querying.
- 🔍 **Native Cypher Support**: First-class support for core Cypher subset (`MATCH`, `CREATE`, `SET`, `DELETE`, `REMOVE`).
- 🛡️ **ACID Transactions**: Backed by CockroachDB's Pebble, featuring Multiversion Concurrency Control (MVCC), thread safety, and Write-Ahead Log (WAL) crash recovery.
- 🛠️ **Built-in CLI & TUI**: Comes with an interactive terminal interface featuring syntax highlighting, auto-completion, and ASCII table rendering.
- 📊 **Observability First**: Native support for injecting custom Loggers, Tracers, and Metrics collectors via the Option pattern.

---

## ⚡ Quick Start

You can embed GoGraph into your application and start querying in under 5 minutes.

### 1. Requirements
- **Go 1.18 or higher** (Go 1.21+ recommended for best performance).
- Supported OS: Linux, macOS, Windows.

### 2. Installation
Get the GoGraph package in your Go module:

```bash
go get github.com/DotNetAge/gograph
```

### 3. Usage Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DotNetAge/gograph/pkg/api"
)

func main() {
	// 1. Open or create the embedded database
	db, err := api.Open("./data/mygraph.db")
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 2. Create nodes and relationships using Cypher
	createCypher := `
		CREATE (alice:User {name: "Alice", age: 28})
		CREATE (bob:User {name: "Bob", age: 30})
		CREATE (alice)-[:KNOWS {since: 2022}]->(bob)
	`
	result, err := db.Exec(ctx, createCypher)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Created %d nodes, %d relationships.\n", result.AffectedNodes, result.AffectedRels)

	// 3. Query the graph
	queryCypher := `MATCH (u:User)-[r:KNOWS]->(friend:User) RETURN u.name, friend.name, r.since`
	rows, err := db.Query(ctx, queryCypher)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// 4. Iterate and read results
	for rows.Next() {
		var uName, fName string
		var since int64
		if err := rows.Scan(&uName, &fName, &since); err == nil {
			fmt.Printf("%s knows %s since %d\n", uName, fName, since)
		}
	}
}
```

---

## 💻 CLI & TUI Guide

GoGraph includes a fully-functional command-line interface and an interactive Terminal User Interface (TUI) for debugging and exploring your graphs.

### Build and Install the CLI

Run the following command in the project root:
```bash
go build -o gograph ./cmd/gograph
```

### Available Commands

The CLI structure is `gograph <command> <database_path> [cypher_query]`.

#### 1. Execute Data Modification (`exec`)
Used for statements that change the graph state (`CREATE`, `SET`, `DELETE`, `REMOVE`).
```bash
./gograph exec ./test.db "CREATE (n:Person {name: 'Alice'})"
```

#### 2. Query Data Retrieval (`query`)
Used for reading data (`MATCH ... RETURN`).
```bash
./gograph query ./test.db "MATCH (n:Person) RETURN n.name"
```
*Output is automatically rendered into an elegant ASCII table.*

#### 3. Interactive Shell (`tui`)
Launch a REPL (Read-Eval-Print Loop) environment equipped with syntax completion, history navigation, and colored outputs.
```bash
./gograph tui ./test.db
```

**Inside the TUI:**
- **Auto-routing**: Simply type `CREATE (n)` or `MATCH (n) RETURN n`, and the TUI will dynamically route it to the correct execution engine.
- **Tab Completion**: Press `TAB` to auto-complete Cypher keywords.
- **Internal Directives**: 
  - `/help` — Display help instructions.
  - `/exit` or `/quit` — Safely close the database and exit.
  - `/exec <cypher>` — Force the execution pipeline.
  - `/query <cypher>` — Force the query pipeline.

---

## 🧩 System Compatibility
- **Architecture**: `amd64`, `arm64`, `386`.
- **OS**: macOS (Darwin), Linux, Windows. 
- **Dependencies**: 100% Pure Go. Relies on `github.com/cockroachdb/pebble` for robust persistent storage, completely bypassing CGO requirements (unlike SQLite/RocksDB).

---

## ❓ FAQ (Frequently Asked Questions)

**1. Why use Pebble DB instead of standard LevelDB/RocksDB?**
Pebble is a pure Go key-value store crafted by CockroachDB. It entirely sidesteps the notorious CGO cross-compilation headaches associated with RocksDB while providing comparable performance, native MVCC, and robust WAL recovery mechanisms.

**2. Is GoGraph thread-safe?**
Yes. The `api.DB` instance handles concurrent queries safely. It utilizes Pebble's internal batching and snapshot capabilities to provide Multi-Version Concurrency Control (MVCC), ensuring readers are never blocked by writers.

**3. What property data types are supported?**
Currently, GoGraph's property engine natively serializes `string`, `int64` (and `int`), `float64`, and `bool`. 

---

## 📚 Documentation
For detailed architecture designs and deeper technical insights, please check out our [Docs Folder](./docs/README.md).
