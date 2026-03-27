# 快速开始 (Quick Start)

通过这篇指南，您将在 30 分钟内完成开发环境配置，并在您的 Go 项目中成功嵌入、运行 GoGraph 图数据库。

## 1. 环境准备

**要求：**
- **Go 1.18 及以上版本**（推荐 Go 1.21+，以获得最佳性能）。
- CGO：不需要（纯 Go 实现，跨平台支持）。

## 2. 安装指南

在您的 Go Module 项目目录下执行以下命令获取依赖：

```bash
go get github.com/DotNetAge/gograph
```

## 3. 基础使用示例

下面演示一个完整的 `main.go` 示例，包含打开数据库、创建节点、创建关系以及执行查询的过程。

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DotNetAge/gograph/pkg/api"
)

func main() {
	// 1. 打开或创建一个图数据库实例，采用单文件存储形式（目录路径）
	db, err := api.Open("./data/mygraph.db")
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	// 确保程序退出时关闭数据库，释放资源
	defer db.Close()

	ctx := context.Background()

	// 2. 执行数据写入 (CREATE)
	// 支持一次性创建带属性的节点，以及节点间的关系
	cypherCreate := `
		CREATE (alice:User:VIP {name: "Alice", age: 28})
		CREATE (bob:User {name: "Bob", age: 30})
		CREATE (alice)-[:KNOWS {since: 2022}]->(bob)
	`
	result, err := db.Exec(ctx, cypherCreate)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Created %d nodes and %d relationships.\n", result.AffectedNodes, result.AffectedRels)

	// 3. 数据修改 (SET/REMOVE)
	_, err = db.Exec(ctx, "MATCH (n:User {name: 'Bob'}) SET n.age = 31 REMOVE n.age")
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	// 4. 数据查询 (MATCH ... RETURN)
	// 查询 User 节点和它们之间的 KNOWS 关系，并支持简单的条件过滤
	cypherQuery := `MATCH (u:User)-[r:KNOWS]->(friend:User) WHERE u.age > 20 RETURN u.name, friend.name, r.since`
	
	rows, err := db.Query(ctx, cypherQuery)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close() // 养成关闭 Rows 的好习惯

	// 5. 遍历结果集并映射数据
	fmt.Println("\n--- Query Results ---")
	for rows.Next() {
		var uName, fName string
		var since int64

		// 依照 RETURN 语句的顺序传入指针接收数据
		if err := rows.Scan(&uName, &fName, &since); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		fmt.Printf("%s knows %s since %d\n", uName, fName, since)
	}
}
```

## 4. 事务与并发

GoGraph 提供了读写安全的多版本并发控制 (MVCC)。当需要原子地执行多个 Cypher 语句时，可使用显式事务。

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    log.Fatal(err)
}

// 发生错误时回滚事务
defer tx.Rollback()

// 在事务中执行语句
_, err = tx.Exec("CREATE (n:Log {event: 'Tx Started'})")
if err != nil {
    return // 将自动触发 Rollback
}

// 提交变更
if err := tx.Commit(); err != nil {
    log.Fatal("Commit failed:", err)
}
```

您已成功掌握了 GoGraph 的核心功能操作。接下来，您可以前往查阅 [核心接口说明](core-interfaces.md) 或 [系统架构设计](architecture.md)。
