# 核心接口说明 (Core Interfaces)

GoGraph 的对外 API 风格借鉴了 Go 标准库 `database/sql` 的设计理念，旨在为 Go 开发者提供直观、原生的调用体验。

所有对外部暴露的核心结构与方法均在 `pkg/api` 包内。

## 1. 数据库实例 (`api.DB`)

### `Open`

**功能描述**：打开并在指定路径初始化一个单文件的图数据库实例。如果文件不存在则自动创建，内部自动配置并开启 Pebble DB 的 WAL 以保证崩溃恢复。

**接口签名**：
```go
func Open(path string, opts ...DBOption) (*DB, error)
```

**参数**：
| 参数名 | 类型 | 是否必填 | 说明 | 默认值 |
|---|---|---|---|---|
| `path` | `string` | 是 | 数据库文件存放的本地目录路径 | 无 |
| `opts` | `...DBOption` | 否 | 功能选项，如配置可观测性机制 | 默认内置选项 |

**返回值**：
- `*api.DB`：数据库实例指针。
- `error`：初始化失败时返回的具体错误原因。

**使用示例**：
```go
// 启用 Tracing 和 Logging 选项注入
obs := cypher.NewObservability(cypher.WithLogger(myLogger))
db, err := api.Open("./test.db", api.WithObservability(obs))
if err != nil {
    panic(err)
}
defer db.Close()
```

---

### `Exec`

**功能描述**：执行**非查询类**的数据变更操作（如：`CREATE`, `SET`, `DELETE`, `REMOVE`），并返回变更所影响的节点和关系计数。支持安全的参数化传入。

**接口签名**：
```go
func (db *DB) Exec(ctx context.Context, cypherQuery string, args ...interface{}) (Result, error)
```

**参数**：
| 参数名 | 类型 | 是否必填 | 说明 |
|---|---|---|---|
| `ctx` | `context.Context` | 是 | 请求上下文，用于超时控制或链路追踪传递 |
| `cypherQuery` | `string` | 是 | 核心 Cypher 执行语句 |
| `args` | `...interface{}`| 否 | 参数化映射或按顺序传入的参数。可直接传入 `map[string]interface{}` 或键值对对齐的变量 |

**返回值**：
- `api.Result`：它是 `cypher.Result` 的类型别名，其中包含 `AffectedNodes`（修改的节点数）和 `AffectedRels`（修改的关系数）。另外还保存了 `Rows` 与 `Columns` 供级联执行子句消费。
- `error`：若为只读事务或执行失败，将返回错误。

**使用示例**：
```go
// 参数化创建（防止 Cypher 注入）
result, err := db.Exec(ctx, "CREATE (n:User {name: $name})", map[string]interface{}{"name": "Alice"})
```

---

### `Query`

**功能描述**：执行**查询类**操作（如：`MATCH ... RETURN`），并返回可迭代的结果集 `Rows`。

**接口签名**：
```go
func (db *DB) Query(ctx context.Context, cypherQuery string, args ...interface{}) (*Rows, error)
```

**返回值**：
- `*api.Rows`：结果集迭代器指针。
- `error`：查询失败（例如语法错误、类型不匹配）时返回的错误信息。

**使用示例**：
```go
rows, err := db.Query(ctx, "MATCH (n:User) RETURN n.name, n.age")
```

---

### `BeginTx`

**功能描述**：开启一个事务（支持 MVCC）。事务提供 ACID 保证。

**接口签名**：
```go
func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error)
```

## 2. 事务操作 (`api.Tx`)

在通过 `BeginTx` 获取事务实例后，您可以进行原子提交或回滚。

### `Commit` 和 `Rollback`

**功能描述**：事务成功时提交持久化（`Commit`）；否则取消操作（`Rollback`）。

**接口签名**：
```go
func (tx *Tx) Commit() error
func (tx *Tx) Rollback() error
```

**说明**：与 `DB` 类似，`Tx` 本身也暴露了 `Exec` 和 `Query` 方法，用法一致。

## 3. 结果集 (`api.Rows`)

`Rows` 提供对查询结果（通常由 `RETURN` 指定的项）进行流式迭代的能力。

### `Next`

**功能描述**：迭代游标到下一行。若返回 `false`，则表明已到达结果集末尾。

**接口签名**：
```go
func (r *Rows) Next() bool
```

### `Scan`

**功能描述**：将当前游标指向的行数据按顺序写入目标指针中。

**接口签名**：
```go
func (r *Rows) Scan(dest ...interface{}) error
```

**参数**：传入所需映射字段的指针列表（如 `*string`, `*int64`, `*bool`, `*float64` 等）。顺序必须与 `RETURN` 声明完全一致。

**返回值**：若越界或类型不匹配会返回 `api.ErrNoMoreRows` 等异常。

### `Close`

**功能描述**：安全关闭游标并释放内部占用。务必在 `Query` 之后搭配 `defer rows.Close()` 妥善关闭。

## 4. 常见错误枚举 (`Error`)

在开发过程中，您可能会遇到以下常量定义的错误码：
- `api.ErrDBClosed`：针对已调用过 `Close()` 的 `DB` 对象进行 `Query`/`Exec` 操作。
- `api.ErrNoMoreRows`：对已到底的迭代器继续调用 `Scan()` 时抛出。
