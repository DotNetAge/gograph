# GoGraph 性能分析与优化方案

## 当前架构分析

### 存储层 (PebbleDB LSM-Tree)

当前使用 PebbleDB 作为底层存储，采用 LSM-Tree 结构。数据以以下键格式存储：

| 数据类型 | 键格式 | 示例 |
|----------|--------|------|
| 节点 | `node:<nodeID>` | `node:01KQF...` |
| 关系 | `rel:<relID>` | `rel:01KQF...` |
| 标签索引 | `label:<label>:<nodeID>` | `label:Person:node:01KQF...` |
| 属性索引 | `prop:<label>:<prop>:<value>` | `prop:Person:name:Alice` |
| 邻接表 | `adj:<nodeID>:<type>:<dir>:<relID>` | `adj:node:01KQF...:KNOWS:out:rel:...` |

### 查询执行流程

1. **MATCH 节点**: 通过标签索引或全表扫描获取候选节点
2. **过滤**: 遍历候选节点，应用 WHERE 条件
3. **关系遍历**: 通过邻接表查找连接节点
4. **返回**: 组装结果集

## 性能瓶颈识别

### 1. 索引冗余 (高优先级)

**问题**: `BuildPropertyIndex` 为节点的**每个标签**都创建属性索引。

```go
// 当前实现 - 冗余索引
func (idx *Index) BuildPropertyIndex(m Mutator, node *Node) error {
    for _, label := range node.Labels {  // 每个标签都建索引
        for propName, propValue := range node.Properties {
            key := storage.PropertyKey(label, propName, encodedValue)
            m.Put(key, []byte(node.ID))
        }
    }
}
```

**影响**: 一个带有 3 个标签、10 个属性的节点会创建 30 条索引记录。

**优化方案**: 只为第一个标签（主标签）创建属性索引，或采用复合索引策略。

### 2. 全表扫描风险 (高优先级)

**问题**: `MATCH (n)` 没有标签过滤时会进行全表扫描。

```go
// matchers/match.go
func (m *Matcher) findNodes(pattern *ast.NodePattern) ([]*graph.Node, error) {
    if len(pattern.Labels) > 0 {
        // 使用索引
        return m.findNodesByLabel(pattern.Labels[0])
    }
    // 无标签时全表扫描！
    return m.findAllNodes()
}
```

**优化方案**: 
- 添加节点类型统计信息，小表优先扫描
- 实现代价估算器

### 3. 迭代器频繁创建 (中优先级)

**问题**: 每次查询都创建新的迭代器，没有连接池或迭代器复用。

```go
// 每次查询都创建新迭代器
iter, err := idx.store.NewIter(nil)
```

**优化方案**: 使用迭代器池或快照读取。

### 4. 数据反序列化开销 (中优先级)

**问题**: 每次从存储读取节点都需要完整的 JSON/protobuf 反序列化。

```go
// storage/marshal.go
func Unmarshal(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)  // 完整反序列化
}
```

**优化方案**: 
- 使用更高效的序列化格式（MessagePack、FlatBuffers）
- 实现列式存储，只读取需要的属性

### 5. 邻接表遍历效率 (中优先级)

**当前实现**: 邻接表存储为独立的 KV 对，遍历需要多次存储访问。

```
adj:node:A:KNOWS:out:rel:1 -> node:B
adj:node:A:KNOWS:out:rel:2 -> node:C
adj:node:A:KNOWS:out:rel:3 -> node:D
```

**优化方案**: 将同一节点的邻接关系合并存储，减少存储访问次数。

## 优化实施计划

### Phase 1: 索引优化 (预期提升 30-50%)

1. **主标签索引策略**
   - 只为第一个标签创建属性索引
   - 查询时优先使用主标签过滤

2. **索引压缩**
   - 使用短标签编码（如 `P` 代替 `Person`）
   - 共享属性值字典

### Phase 2: 存储格式优化 (预期提升 20-40%)

1. **MessagePack 替代 JSON**
   - 更小的序列化体积
   - 更快的编解码速度

2. **邻接表合并存储**
   ```
   // 优化前: 每个关系一个 KV
   adj:node:A:KNOWS:out:rel:1 -> node:B
   adj:node:A:KNOWS:out:rel:2 -> node:C
   
   // 优化后: 一个 KV 存储所有邻居
   adj:node:A:KNOWS:out -> [rel:1:node:B, rel:2:node:C]
   ```

### Phase 3: 查询优化器 (预期提升 50-200%)

1. **统计信息收集**
   - 节点/关系数量统计
   - 标签分布直方图
   - 属性值选择性分析

2. **基于代价的优化**
   - 选择最优的扫描顺序
   - 下推过滤条件
   - 索引选择策略

## 基准测试建议

```go
// benchmark_test.go
func BenchmarkNodeCreation(b *testing.B) {
    db, _ := api.Open(":memory:")
    defer db.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        db.Exec(ctx, "CREATE (n:Person {name: 'Alice', age: 30})")
    }
}

func BenchmarkMatchByLabel(b *testing.B) {
    // 准备 10000 个节点
    db, _ := api.Open(":memory:")
    defer db.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        db.Query(ctx, "MATCH (p:Person) WHERE p.age > 25 RETURN p.name")
    }
}

func BenchmarkRelationshipTraversal(b *testing.B) {
    // 准备深度为 5 的图
    db, _ := api.Open(":memory:")
    defer db.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        db.Query(ctx, "MATCH (a)-[:KNOWS*3]->(b) RETURN b.name")
    }
}
```

## 与竞品对比目标

| 数据库 | 节点创建 (ops/s) | 标签查询 (ops/s) | 深度遍历 (ops/s) |
|--------|------------------|------------------|------------------|
| Neo4j (嵌入式) | ~50,000 | ~20,000 | ~5,000 |
| Dgraph | ~100,000 | ~50,000 | ~10,000 |
| **GoGraph 目标** | **>200,000** | **>100,000** | **>20,000** |

## 免索引邻接 (Index-Free Adjacency) 分析

### 概念
免索引邻接是指每个节点直接存储指向邻居节点的指针（物理地址或 ID），无需通过全局索引查找关系。

### 当前实现评估

GoGraph 的邻接表实现已经接近免索引邻接：

```
adj:node:A:KNOWS:out:rel:1 -> node:B
```

**优点**:
- 给定节点 ID，可以直接定位邻接表（O(1)）
- 无需全局关系索引

**差距**:
- 仍然使用 KV 存储的 Seek 操作，不是真正的内存指针
- 需要反序列化完整的邻接记录

### 实现方案

**方案 A: 内存映射邻接表 (激进)**
- 启动时将邻接表加载到内存
- 使用 `map[nodeID][]AdjEntry` 结构
- 写入时同步更新内存和磁盘

**方案 B: 节点内联邻接 (中等)**
- 修改节点存储格式，在节点记录内包含邻接 ID 列表
- 减少存储访问次数

```
// 优化前
node:01KQF... -> {name: "Alice", labels: ["Person"], properties: {...}}
adj:node:01KQF...:KNOWS:out:rel:1 -> node:B

// 优化后
node:01KQF... -> {name: "Alice", labels: ["Person"], properties: {...}, 
                  adj_out: [{type: "KNOWS", rel: "rel:1", target: "node:B"}]}
```

**方案 C: 缓存层 (保守)**
- 在 AdjacencyList 上添加 LRU 缓存
- 缓存热点节点的邻接关系

### 推荐方案

建议采用 **方案 C (缓存层)** 作为第一步，因为：
1. 改动最小，风险最低
2. 对现有存储结构无侵入
3. 可以显著改善热点查询性能

后续可以考虑 **方案 B (节点内联邻接)** 来进一步减少存储访问。
