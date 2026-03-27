# 开发者指南 (Developer Guide)

欢迎参与 GoGraph 相关的扩展与使用。本文档汇总了实用开发技巧、最佳实践与常见问题的解决方案。

## 1. 最佳实践 (Best Practices)

### 安全的参数化查询

为防止 Cypher 注入风险，并且由于预编译参数可提高执行器性能，**始终推荐使用参数化查询**：

```go
// 错误示例（易造成注入和性能低下）：
// cypher := fmt.Sprintf("MATCH (n:User {name: '%s'}) RETURN n", userName)

// 推荐做法：
ctx := context.Background()
cypherQuery := "MATCH (n:User {name: $name}) RETURN n"
args := map[string]interface{}{
    "name": "Alice",
}
rows, err := db.Query(ctx, cypherQuery, args)
```

### 资源释放习惯

游标与事务的泄露会导致系统资源的枯竭和写入锁死。请时刻利用 `defer`：

```go
rows, err := db.Query(ctx, "MATCH (n) RETURN n")
if err != nil {
    return err
}
defer rows.Close() // 关键：释放迭代器内部资源

tx, _ := db.BeginTx(ctx, nil)
defer tx.Rollback() // 关键：即便产生 panic，也会回滚释放 MVCC
// ...
tx.Commit() // 显式提交
```

### 可观测性注入 (Observability)

GoGraph 原生支持日志 (`Logger`)、链路追踪 (`Tracer`) 及指标收集 (`Meter`) 的依赖注入，这极大地方便了在微服务环境中的监控：

```go
import "github.com/DotNetAge/gograph/pkg/cypher"

// 注入你的自定义 Logger (必须实现 cypher.Logger 接口)
obs := cypher.NewObservability(
    cypher.WithLogger(myCustomLogger),
    cypher.WithTracer(myOpenTelemetryTracer),
)

// 通过 Option 模式注册给 DB 实例
db, err := api.Open("./data.db", api.WithObservability(obs))
```

## 2. 常见问题 (FAQ)

### 为什么选择 Pebble 作为存储引擎？
Pebble 是 CockroachDB 团队为了替代 RocksDB 和 LevelDB 研发的纯 Go KV 库。它避免了 CGO 的编译跨平台难题，同时支持 WAL 预写日志，拥有出色的写性能和原生的 MVCC（多版本并发控制）支持。非常适合作为内嵌式数据库的引擎。

### MVCC 的锁机制是怎样的？
GoGraph 采用了“读写分离”设计，依靠底层的 Pebble DB 实现的 `NewBatch`：
- **读写事务**可以并行读取，但当提交写操作时如果发生竞争（尽管现在是由应用级 `sync.RWMutex` 或底层保证），未提交的数据对其他会话不可见。
- 脏读被彻底避免，读操作绝不会被写操作所阻塞。

### 支持哪些类型的属性值？
GoGraph 中节点和关系的属性 `PropertyValue` 核心支持四种类型：
- `string`
- `int64`（可兼容 `int`，内部存储为 `int64`）
- `float64`
- `bool`
若传入其他类型对象，内部会调用 `fmt.Sprintf("%v", v)` 自动将其转化为字符串，这可能对数据结构分析产生不一致的副作用，请尽量只传入支持的基础类型。

## 3. 代码与编码规范

1. **Option 模式**: 任何组件的扩展配置（如 `cypher.NewExecutor`, `api.Open`）必须遵循 Functional Options 范式，以确保 API 的向后兼容性。
2. **错误处理**: GoGraph 的错误设计明确（如 `api.ErrDBClosed`, `api.ErrNoMoreRows`）。如果底层返回 `pebble.ErrNotFound`，请在上层包装为 `cypher`/`api` 的领域相关错误。
3. **高内聚的充血模型**: 在最近的系统重构中，底层所有的 KV 拼接和检索操作均已收拢回 `pkg/graph` 领域层（如 `Index` 和 `AdjacencyList`）。`Creator`、`Matcher` 等调度层不应再越权直接操纵裸存储字节。未来新增索引和关系特性时，请遵循 `graph.Mutator` 的设计模式将修改委托回对应的实体层。

## 4. 性能优化 (Performance Tips)

### 利用索引扫描与图遍历 (Index & Graph Traversal)
为获得极致的数据检索性能，在编写您的 Cypher 查询时请充分利用新引擎的特性：
- **指定 Label 避开全表扫**：对于 `MATCH (n:User)`，引擎将走 **Index Scan**；如果不带 Label 直接 `MATCH (n)`，系统只能退化执行全库的 Node Key 顺序扫描，在海量数据下会引起明显延迟。
- **顺藤摸瓜利用邻接表**：GoGraph 的关联关系 `AdjacencyList` 已完美支持 $O(1)$ 指针跳跃。当起点的 ID 被定位后，查询 `MATCH (n)-[r]->(m)` 只会沿关联边拉取 `m`，而不会执行全库连接（JOIN），使得复杂深度遍历维持极低耗时。
