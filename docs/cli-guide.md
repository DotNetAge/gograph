# 命令行工具 (CLI & TUI) 指南

GoGraph 提供了一个功能丰富且开箱即用的命令行工具（CLI），您可以利用它快速连接到本地图数据库、执行 Cypher 语句，并以直观的表格形式查看数据结果。

该工具不仅支持一次性的命令执行，还内置了包含语法高亮和自动补全的交互式终端用户界面（TUI）。

## 1. 编译与安装

在项目根目录下，执行以下命令即可编译生成 `gograph` 可执行文件：

```bash
go build -o gograph ./cmd/gograph
```

编译完成后，您可以通过 `./gograph -h` 查看基础帮助信息。

## 2. 基础命令结构

CLI 的所有命令均遵循以下标准调用格式：

```bash
gograph <command> <database_path> [cypher_query]
```

- `<command>`：子命令（`exec`, `query`, `tui`），与 GoGraph API 核心语义保持严格一致。
- `<database_path>`：目标数据库文件路径（例如 `./test.db`）。如果该路径下没有数据库文件，系统将自动创建。
- `[cypher_query]`：具体的 Cypher 语句（使用双引号包裹）。

### 2.1 数据修改 (`exec`)

**功能描述**：用于执行会改变图状态的语句（如 `CREATE`, `SET`, `DELETE`, `REMOVE`）。等价于 API 中的 `db.Exec()`。

**使用示例**：
```bash
./gograph exec ./test.db "CREATE (alice:User {name: 'Alice', age: 30})-[:KNOWS]->(bob:User {name: 'Bob'})"
```

**输出**：成功后将打印影响的节点与关系数量（高亮绿色）。
```text
Query executed successfully.
Affected Nodes: 2
Affected Rels:  1
```

### 2.2 数据查询 (`query`)

**功能描述**：用于执行数据检索类语句（如 `MATCH ... RETURN`）。等价于 API 中的 `db.Query()`。

**使用示例**：
```bash
./gograph query ./test.db "MATCH (n:User)-[r:KNOWS]->(m:User) RETURN n.name, m.name"
```

**输出**：结果将自动被格式化为美观的 ASCII 数据表。
```text
+--------+--------+
| N NAME | M NAME |
+--------+--------+
| Alice  | Bob    |
+--------+--------+

Returned 1 rows.
```

---

## 3. 交互式终端界面 (TUI 模式)

TUI 是日常调试、查询探索的最佳工具，它接管了终端输入并提供 REPL（Read-Eval-Print Loop）环境。

**启动 TUI**：
```bash
./gograph tui ./test.db
```

### 3.1 核心特性

- **彩色输出控制**：不同的输出类型（提示、成功、错误、数据表）使用标准 ANSI 色彩进行区分（支持降级渲染），极大增强视觉体验。
- **命令自动补全**：按下 `TAB` 键，即可触发内置关键字（如 `MATCH`, `CREATE`, `RETURN`, `WHERE` 等）和斜杠指令的自动补全。
- **历史记录导航**：支持使用键盘 `Up` / `Down`（上下方向键）在会话中翻查曾执行过的历史命令。

### 3.2 TUI 内部指令

进入 TUI 后，输入提示符会变为 `gograph> `。您可以通过键入 Cypher 直接执行操作，系统会智能推断您的意图：
- 以 `CREATE`, `SET`, `DELETE`, `REMOVE` 开头的语句，将自动走 `Exec` 数据修改通道。
- 其他语句（如 `MATCH`）将自动走 `Query` 查询检索通道。

如果需要显式指定行为，可以采用以下斜杠命令：

| 指令 | 说明 | 示例 |
|---|---|---|
| `/help` | 打印当前可用的指令与帮助说明 | `gograph> /help` |
| `/exec` | 强制以变更模式执行后续语句 | `gograph> /exec CREATE (n)` |
| `/query` | 强制以查询模式执行后续语句 | `gograph> /query MATCH (n) RETURN n` |
| `/exit` 或 `/quit` | 安全断开与数据库的连接，退出 TUI | `gograph> /exit` |

### 3.3 数据自动格式化

在 TUI 中进行查询时，无论 `RETURN` 子句返回的是基础属性还是完整的图结构实体，CLI 都具备优雅打印的能力：

```cypher
gograph> MATCH (n) RETURN n
```
```text
+-------------------------------------+
|                  N                  |
+-------------------------------------+
| (node:id1:User {name:Alice, age:30})|
+-------------------------------------+
(1 rows)
```
无论是 `*graph.Node` 还是 `*graph.Relationship` 都能被格式化为直观的 Cypher 风格文本 `(ID:Label {Props})` 或 `[ID:Type {Props}]`，方便快速查看图结构全貌。
