# API 参考 (API Reference)

该文档按代码的模块和包（Package）进行分类，列举核心结构与导出方法。

## 1. `pkg/api`

该包包含针对外部提供的高阶通用图库接口抽象。

- **`func Open(path string, opts ...DBOption) (*DB, error)`**
  打开路径所在的数据库文件。若不存在则初始化为 Pebble KV 库结构。
  
- **`type DBOption func(*DB)`**
  函数选项模式的配置定义。

- **`func WithObservability(o *cypher.Observability) DBOption`**
  创建一个允许注入配置好日志、追踪和监控计数的 `DBOption`。

- **`type DB struct`**
  核心的线程安全图库实例，控制与底层引擎的交互：
  - `func (db *DB) Exec(ctx context.Context, cypherQuery string, args ...interface{}) (Result, error)`
  - `func (db *DB) Query(ctx context.Context, cypherQuery string, args ...interface{}) (*Rows, error)`
  - `func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error)`
  - `func (db *DB) Close() error`
  - `func (db *DB) IsClosed() bool`

- **`type Result = cypher.Result`**
  包含了对图做修改操作后的状态及计数（别名）：
  - `AffectedNodes int`：影响的节点数量。
  - `AffectedRels int`：影响的关系数量。
  - `Rows []map[string]interface{}`：内部匹配的数据行（用于链式执行）。
  - `Columns []string`：内部列名。

- **`type Rows struct`**
  图引擎的迭代游标：
  - `func (r *Rows) Next() bool`
  - `func (r *Rows) Scan(dest ...interface{}) error`
  - `func (r *Rows) Columns() []string`
  - `func (r *Rows) Close() error`

- **`type Tx struct`**
  安全的并发多版本控制事务包装：
  - `func (tx *Tx) Exec(cypherQuery string, args ...interface{}) (Result, error)`
  - `func (tx *Tx) Query(cypherQuery string, args ...interface{}) (*Rows, error)`
  - `func (tx *Tx) Commit() error`
  - `func (tx *Tx) Rollback() error`

## 2. `pkg/cypher`

处理 Cypher 查询语句及其执行环境的包。

- **`type Executor struct`**
  执行抽象语法树的上下文对象。通过组合 `Creator`, `Matcher` 和 `Modifier`。
  
- **`func NewExecutor(store *storage.DB, opts ...ExecutorOption) *Executor`**
  创建执行器对象实例。
  
- **`type Observability struct`**
  由 `Logger`, `Tracer` 和 `Meter` 三大能力接口组成。
  
- **`func NewObservability(opts ...ObservabilityOption) *Observability`**
  生成具有默认空实现的观察者，可通过 `WithLogger`, `WithTracer`, `WithMeter` 替换。

## 3. `pkg/graph`

描述图数据库核心模型及序列化元数据。

- **`type Mutator interface`**
  定义了事务/存储的抽象修改能力，能够让实体参与到底层事务中：
  - `Put(key, value []byte) error`
  - `Delete(key []byte) error`

- **`type Index struct`**
  管理节点标签和属性的高效索引构建与查询。

- **`type AdjacencyList struct`**
  图的邻接表结构，支持 $O(1)$ 的关系双向查询与维护。

- **`type Node struct`**
  - `ID         string`
  - `Labels     []string`
  - `Properties map[string]PropertyValue`
  
  提供 `GetProperty`, `SetProperty`, `HasLabel`, `RemoveLabel` 等快捷方法。

- **`type Relationship struct`**
  - `ID          string`
  - `StartNodeID string`
  - `EndNodeID   string`
  - `Type        string`
  - `Properties  map[string]PropertyValue`

- **`type PropertyValue struct`**
  包装基础数据类型：
  - `Type() PropertyType`
  - `StringValue() string`
  - `IntValue() int64`
  - `FloatValue() float64`
  - `BoolValue() bool`

## 4. `pkg/storage`

包含对 Pebble 底层的高效交互及自定义图数据库的二进制键值对定义。

- **`func Open(path string) (*DB, error)`**
  打开 KV 持久化层。

- **`func (db *DB) Put(key, value []byte) error`**
- **`func (db *DB) Get(key []byte) ([]byte, error)`**
- **`func (db *DB) Delete(key []byte) error`**
- **`func Marshal(v interface{}) ([]byte, error)`**
- **`func Unmarshal(data []byte, v interface{}) error`**
