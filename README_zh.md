<div align="center">
  <h1>GoGraph</h1>
  <p>纯 Go 编写的极简嵌入式图数据库</p>

  [![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/gograph.svg)](https://pkg.go.dev/github.com/DotNetAge/gograph)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/DotNetAge/gograph)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/DotNetAge/gograph)](https://goreportcard.com/report/github.com/DotNetAge/gograph)
  [![Docs](https://img.shields.io/badge/docs-gograph.rayainfo.cn-2094f3.svg)](https://gograph.rayainfo.cn)
		
  <p>
    <strong>简体中文</strong> | <a href="README.md">English</a>
  </p>
</div>

---

## 📖 项目概述

**GoGraph** 是一款纯 Go 编写的轻量级、零依赖的嵌入式图数据库。您可以将它理解为**“图数据库界的 SQLite”**。

它允许 Go 开发者直接在本地存储中执行 Cypher 查询语言，无需部署、维护以及连接 Neo4j 等重型外部数据库服务。

## ⚡ 快速开始 (Quick Start)

### 1. 安装命令行工具 (推荐)
探索 GoGraph 最快的方式是使用其命令行界面（CLI）。

**macOS / Linux (Homebrew):**
```bash
brew install dotnetage/tap/gograph
```

**运行交互式 Shell (TUI):**
```bash
# 不带参数直接运行，将自动打开并进入默认数据库 (default.db)
gograph
```

### 2. 作为 Go 库使用
在您的 Go 项目中引入 GoGraph：
```bash
go get github.com/DotNetAge/gograph
```

**简单示例：**
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
		fmt.Println("用户:", name)
	}
}
```

## ✨ 特色亮点

- 🚀 **纯 Go 实现**：完全没有 CGO 依赖，保证了跨平台编译的极简体验。
- 📦 **极简嵌入式**：无外部服务依赖，数据存储在本地目录中（基于 Pebble DB）。
- 🔍 **原生 Cypher 支持**：深度支持 `MATCH`, `CREATE`, `SET`, `DELETE` 等核心子集。
- 🛡️ **ACID 事务保障**：支持多版本并发控制 (MVCC)、并发读写安全及故障恢复。
- 🛠️ **内置 TUI**：自带带自动补全和 ASCII 数据表的交互式终端环境。

## 💻 CLI 命令行使用

`gograph` 二进制文件提供了一个强大的交互式 TUI 和命令行工具。

| 命令                     | 说明                                   |
| ------------------------ | -------------------------------------- |
| `gograph`                | 启动交互式 TUI (默认路径 `default.db`) |
| `gograph query <cypher>` | 执行只读查询                           |
| `gograph exec <cypher>`  | 执行数据修改指令                       |

**使用示例：**
```bash
gograph query "MATCH (n) RETURN n"
```

---

## 🧩 系统兼容性
- **系统**: macOS, Linux, Windows. 
- **架构**: `amd64`, `arm64`.

## 📚 详细文档
查阅完整的 [官方文档](https://gograph.rayainfo.cn) 或查看 [Docs 目录](./docs/README.md)。
