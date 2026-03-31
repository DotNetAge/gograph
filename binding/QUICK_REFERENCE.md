# GoGraph Binding 快速参考

## 🚀 快速开始

```bash
# 1. 安装依赖
brew install swig go  # macOS
sudo apt-get install swig golang python3-dev  # Ubuntu

# 2. 构建绑定
cd gograph/binding
./build.sh python

# 3. 运行示例
python3 examples/simple_example.py
```

## 📦 核心 API

### 数据库管理

```python
# 创建
db = gograph.gograph_database_new(b"path.db".decode(), error)

# 打开
db = gograph.gograph_database_open(b"path.db".decode(), error)

# 关闭
gograph.gograph_database_close(db)

# 释放
gograph.gograph_database_free(db)
```

### 事务管理

```python
# 开始
tx = gograph.gograph_transaction_begin(db, False, error)

# 提交
gograph.gograph_transaction_commit(tx, error)

# 回滚
gograph.gograph_transaction_rollback(tx)

# 释放
gograph.gograph_transaction_free(tx)
```

### Cypher 查询

```python
result = gograph.QueryResult()
ret = gograph.gograph_cypher_execute(
    tx, 
    "MATCH (n) RETURN n", 
    None, 
    result, 
    error
)

if ret == gograph.GOGRAPH_OK:
    print(f"Rows: {result.row_count}")

# 清理
gograph.gograph_query_result_free(result)
```

## 🔢 错误码

```python
GOGRAPH_OK                    = 0   # 成功
GOGRAPH_ERROR_GENERAL         = 1   # 一般错误
GOGRAPH_ERROR_INVALID_PARAM   = 2   # 无效参数
GOGRAPH_ERROR_NOT_FOUND       = 3   # 未找到
GOGRAPH_ERROR_ALREADY_EXISTS  = 4   # 已存在
GOGRAPH_ERROR_PARSE_FAILURE   = 5   # 解析失败
GOGRAPH_ERROR_EXEC_FAILURE    = 6   # 执行失败
GOGRAPH_ERROR_STORAGE_FAILURE = 7   # 存储故障
GOGRAPH_ERROR_MEMORY_ALLOC    = 8   # 内存分配失败
GOGRAPH_ERROR_TX_CONFLICT     = 9   # 事务冲突
```

## 💡 使用模式

### 完整的资源管理

```python
error = gograph.ErrorInfo()
db = None
tx = None

try:
    # 1. 创建数据库
    db = gograph.gograph_database_new(b"test.db".decode(), error)
    
    # 2. 开始事务
    tx = gograph.gograph_transaction_begin(db, False, error)
    
    # 3. 执行操作
    result = gograph.QueryResult()
    ret = gograph.gograph_cypher_execute(
        tx, "MATCH (n) RETURN count(n)", None, result, error
    )
    
    # 4. 提交
    gograph.gograph_transaction_commit(tx, error)
    
finally:
    # 5. 清理（总是执行）
    if tx is not None:
        gograph.gograph_transaction_free(tx)
    if db is not None:
        gograph.gograph_database_free(db)
```

### 字符串编码

```python
# 正确方式
path = "mygraph.db".encode('utf-8').decode('utf-8')

# 错误方式 ❌
path = "mygraph.db"  # 可能导致编码问题
```

## 🛠️ 构建命令速查

```bash
# Python
./build.sh python
make python

# Java
./build.sh java
make java

# C#
./build.sh csharp
make csharp

# 全部
./build.sh all
make all

# 清理
./build.sh clean
make clean
```

## 📁 文件结构

```
binding/
├── gograph_c.h          # C API 头文件
├── gograph_c.go         # Go 实现
├── gograph.i            # SWIG 接口
├── build.sh             # 构建脚本
├── Makefile             # Make 配置
├── BUILD_GUIDE.md       # 构建指南
├── README.md            # API 文档
├── ARCHITECTURE.md      # 架构设计
└── examples/
    └── simple_example.py  # Python 示例
```

## ⚠️ 常见问题

### Q: IDE 显示 CGO 错误？
**A**: 忽略它们，以命令行构建为准
```bash
./build.sh python  # 能成功就无误
```

### Q: 找不到 Python.h？
**A**: 安装 Python 开发包
```bash
# macOS
brew install python@3.9

# Ubuntu
sudo apt-get install python3-dev
```

### Q: SWIG 未找到？
**A**: 安装 SWIG
```bash
# macOS
brew install swig

# Ubuntu
sudo apt-get install swig
```

### Q: 运行时找不到库？
**A**: 设置库路径
```bash
# Linux
export LD_LIBRARY_PATH=$PWD:$LD_LIBRARY_PATH

# macOS
export DYLD_LIBRARY_PATH=$PWD:$DYLD_LIBRARY_PATH
```

## 📊 性能提示

### 批量操作（推荐）

```python
# ✅ 高效：一次调用
nodes = [{"label": "Person", "props": {"id": i}} for i in range(1000)]
gograph.gograph_node_create_batch(tx, nodes, ...)

# ❌ 低效：多次调用
for i in range(1000):
    gograph.gograph_node_create(tx, "Person", f'{{"id":{i}}}', ...)
```

### 连接复用

```python
# ✅ 复用连接
db = gograph.gograph_database_new(...)
for query in queries:
    tx = gograph.gograph_transaction_begin(db, False, error)
    # 执行查询
    gograph.gograph_transaction_commit(tx, error)
gograph.gograph_database_free(db)

# ❌ 频繁创建销毁
for query in queries:
    db = gograph.gograph_database_new(...)
    tx = gograph.gograph_transaction_begin(db, False, error)
    # ...
    gograph.gograph_database_free(db)
```

## 🔗 相关文档

- [完整构建指南](BUILD_GUIDE.md)
- [API 参考](README.md)
- [架构设计](ARCHITECTURE.md)
- [完成总结](COMPLETION_SUMMARY.md)

---

**快速参考版本**: v1.0  
**更新日期**: 2026-03-31
