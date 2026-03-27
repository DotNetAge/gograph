<div align="center">
  <h1>GoGraph</h1>
  <p>纯 Go 编写的极简嵌入式图数据库</p>
  <p>
    <strong>中文</strong> | <a href="README_en.md">English</a>
  </p>
</div>

---

## 📖 项目概述

**GoGraph** 是一款纯 Go 编写的轻量级、零依赖的嵌入式图数据库。您可以将它理解为**“图数据库界的 SQLite”**。

它允许 Go 开发者直接在本地存储中执行 Cypher 查询语言，并轻松管理节点、关系和属性等复杂图数据。GoGraph 彻底免去了部署、维护以及连接诸如 Neo4j 或 NebulaGraph 等重型外部数据库服务的烦恼。

借助 GoGraph，您不仅能获得图数据模型强大的表达能力，还能兼享单文件本地存储的简单性与极致性能。

## ✨ 特色亮点

- 🚀 **纯 Go 实现**：完全没有 CGO 依赖，保证了跨平台编译的极简体验。
- 📦 **极简嵌入式**：无外部服务依赖，所有数据存储在本地单目录中（基于 Pebble DB），只需 `go get` 即可接入查询。
- 🔍 **原生 Cypher 支持**：深度支持 Cypher 核心子集（`MATCH`, `CREATE`, `SET`, `DELETE`, `REMOVE` 等）。
- 🛡️ **ACID 事务保障**：底层存储由 CockroachDB 出品的 Pebble 驱动，原生支持多版本并发控制 (MVCC)、并发读写安全以及预写日志 (WAL) 崩溃恢复能力。
- 🛠️ **内置交互式 CLI & TUI**：自带命令行工具与交互式终端（TUI），支持语法高亮、自动补全以及炫酷的 ASCII 数据表结果展示。
- 📊 **可观测性优先**：支持通过 Option 模式原生注入自定义的日志 (Logger)、链路追踪 (Tracer) 和指标收集 (Meter) 组件。

---

## ⚡ 快速开始 (Quick Start)

只需 3 分钟，即可将 GoGraph 嵌入到您的应用中并开启 Cypher 查询。

### 1. 环境要求
- **Go 1.18 及以上版本**（推荐 Go 1.21+ 以获得最佳性能表现）。
- 兼容的操作系统：Linux, macOS, Windows。

### 2. 下载安装
在您的 Go Module 项目中获取依赖：

```bash
go get github.com/DotNetAge/gograph
```

### 3. 代码示例

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DotNetAge/gograph/pkg/api"
)

func main() {
	// 1. 打开或创建嵌入式图数据库
	db, err := api.Open("./data/mygraph.db")
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	// 切记释放资源
	defer db.Close()

	ctx := context.Background()

	// 2. 使用 Cypher 写入节点和关系
	createCypher := `
		CREATE (alice:User {name: "Alice", age: 28})
		CREATE (bob:User {name: "Bob", age: 30})
		CREATE (alice)-[:KNOWS {since: 2022}]->(bob)
	`
	result, err := db.Exec(ctx, createCypher)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("成功创建了 %d 个节点，%d 条关系。\n", result.AffectedNodes, result.AffectedRels)

	// 3. 查询图数据
	queryCypher := `MATCH (u:User)-[r:KNOWS]->(friend:User) RETURN u.name, friend.name, r.since`
	rows, err := db.Query(ctx, queryCypher)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// 4. 遍历并读取返回的结果集
	for rows.Next() {
		var uName, fName string
		var since int64
		if err := rows.Scan(&uName, &fName, &since); err == nil {
			fmt.Printf("%s 认识 %s 始于 %d 年\n", uName, fName, since)
		}
	}
}
```

---

## 💻 CLI 命令行使用指南

除了作为内嵌代码库，GoGraph 还附带了一个功能完整的命令行界面与交互式 TUI，方便您直接调试和浏览图数据。

### 编译 CLI 工具

在项目根目录下执行编译：
```bash
go build -o gograph ./cmd/gograph
```

### 所有可用命令

全局的命令行调用格式为：`gograph <command> <database_path> [cypher_query]`。

#### 1. 执行数据修改 (`exec`)
用于会修改图内部状态的语句（例如 `CREATE`, `SET`, `DELETE`, `REMOVE`）。
```bash
./gograph exec ./test.db "CREATE (n:Person {name: 'Alice'})"
```

#### 2. 执行数据检索 (`query`)
用于不改变内部数据的只读语句（如 `MATCH ... RETURN`）。
```bash
./gograph query ./test.db "MATCH (n:Person) RETURN n.name"
```
*返回的结果集将被自动渲染为整洁的 ASCII 表格形式。*

#### 3. 启动交互式界面 (`tui`)
开启内置 REPL 的交互式用户终端（TUI），享受代码自动补全、历史翻页和彩色输出。
```bash
./gograph tui ./test.db
```

**TUI 内部操作秘籍：**
- **智能路由**：直接在界面中敲入 `CREATE (n)` 或 `MATCH (n) RETURN n`，TUI 会自动嗅探您的意图并分发给对应执行器。
- **自动补全**：按下 `TAB` 键可以补全 Cypher 关键字和斜杠命令。
- **强制指令**：
  - `/help` — 打印内部指令帮助。
  - `/exit` 或 `/quit` — 安全断开连接并退出 TUI。
  - `/exec <cypher>` — 显式声明这条指令为修改操作。
  - `/query <cypher>` — 显式声明这条指令为查询操作。

---

## 🧩 系统兼容性
- **系统架构支持**：`amd64`, `arm64`, `386` 等 Go 原生支持的架构。
- **操作系统**：macOS (Darwin), Linux, Windows。
- **库依赖情况**：100% 纯 Go 实现。底层的单文件持久化引擎选用了 `github.com/cockroachdb/pebble`，从而彻底绕过了如同 SQLite 或 RocksDB 带来的跨系统 CGO 编译梦魇。

---

## ❓ 常见问题解答 (FAQ)

**1. 为什么采用 Pebble DB 而非传统的 LevelDB/RocksDB？**
Pebble 是由 CockroachDB 团队为了替代 RocksDB 而研发的纯 Go 键值存储引擎。它完美避开了跨环境的 CGO 编译依赖，同时又能提供相当出色的读写性能、原生级别的 MVCC 和强大的预写日志（WAL）故障恢复机制，是当前构建 Go 内嵌数据库的不二之选。

**2. GoGraph 是线程安全（Thread-Safe）的吗？**
是的。对外暴露的 `api.DB` 实例不仅是并发安全的，底层 Pebble 也利用读写批处理（Batch）和快照技术（Snapshot）实现了多版本并发控制（MVCC）。这意味着在同一个数据库文件中并发进行读取请求时，永远不会被正在写入的操作所阻塞。

**3. GoGraph 节点和关系目前支持哪些属性数据类型？**
当前 GoGraph 引擎能原生地序列化这四种基本类型：`string`, `int64` (亦兼容传入 `int`), `float64`, 和 `bool`。

---

## 📚 详细文档
关于 GoGraph 的核心接口规范、详细架构设计图解和开发者参与指南，请前往 [文档目录 (Docs Folder)](./docs/README.md) 查阅。
