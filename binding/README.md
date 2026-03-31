# GoGraph Binding - 跨语言绑定

[![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/gograph/binding.svg)](https://pkg.go.dev/github.com/DotNetAge/gograph/binding)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

GoGraph 的跨语言绑定层，通过 SWIG 实现与 Python、Java、C# 等语言的互操作。

## 📖 概述

`binding` 包为 GoGraph 图数据库提供了完整的 C API 和跨语言绑定。它作为适配层，将纯 Go 实现的核心功能暴露给其他编程语言。

### 核心特性

- ✅ **架构隔离**: 保持核心 Go 包纯净，无 CGO 依赖
- ✅ **单向依赖**: binding → core，无循环依赖
- ✅ **多语言支持**: Python、Java、C# 等（通过 SWIG）
- ✅ **类型安全**: 完整的类型转换和错误处理
- ✅ **内存管理**: 明确的资源分配和释放接口
- ✅ **零成本抽象**: 直接映射到 Go 核心功能，无额外开销

## 🏗️ 架构设计

```
┌─────────────────────────────────────┐
│     Python / Java / C# Code         │
└──────────────┬──────────────────────┘
               │
               ↓ SWIG Generated Bindings
┌─────────────────────────────────────┐
│     gograph.i (SWIG Interface)      │
└──────────────┬──────────────────────┘
               │
               ↓ C API
┌─────────────────────────────────────┐
│     gograph_c.h / gograph_c.go      │  ← binding 包
│     (CGO Adapter Layer)             │
└──────────────┬──────────────────────┘
               │
               ↓ Pure Go
┌─────────────────────────────────────┐
│     pkg/* (Core Go Implementation)  │  ← 核心包
│     - api/                          │
│     - cypher/                       │
│     - graph/                        │
│     - storage/                      │
│     - tx/                           │
└─────────────────────────────────────┘
```

### 目录结构

```
binding/
├── gograph_c.h          # C API 头文件
├── gograph_c.go         # Go 导出实现
├── gograph.i            # SWIG 接口文件
├── build.sh             # 一键构建脚本
├── Makefile             # Make 配置
├── BUILD_GUIDE.md       # 详细构建指南
├── README.md            # 本文件
└── examples/
    └── simple_example.py  # Python 示例
```

## 🚀 快速开始

### 1. 安装依赖

**macOS:**
```bash
brew install swig go
```

**Ubuntu/Debian:**
```bash
sudo apt-get install swig golang python3-dev
```

### 2. 构建绑定

```bash
cd gograph/binding
./build.sh python
```

或使用 Makefile:
```bash
make python
```

### 3. 运行示例

```bash
python3 examples/simple_example.py
```

## 📦 API 参考

### 数据库管理

```c
// 创建数据库
DatabaseHandle gograph_database_new(const char* db_path, ErrorInfo* error);

// 打开数据库
DatabaseHandle gograph_database_open(const char* db_path, ErrorInfo* error);

// 关闭数据库
int gograph_database_close(DatabaseHandle handle);

// 释放句柄
void gograph_database_free(DatabaseHandle handle);
```

### 事务管理

```c
// 开始事务
TransactionHandle gograph_transaction_begin(
    DatabaseHandle db_handle,
    bool read_only,
    ErrorInfo* error
);

// 提交事务
int gograph_transaction_commit(TransactionHandle tx_handle, ErrorInfo* error);

// 回滚事务
int gograph_transaction_rollback(TransactionHandle tx_handle);
```

### Cypher 查询

```c
// 执行查询
int gograph_cypher_execute(
    TransactionHandle tx_handle,
    const char* query,
    const char* params,
    QueryResult* result,
    ErrorInfo* error
);
```

### 节点操作

```c
// 创建节点
int gograph_node_create(
    TransactionHandle tx_handle,
    const char* label,
    const char* properties,
    uint64_t* node_id,
    ErrorInfo* error
);

// 查询节点
int gograph_node_get(
    TransactionHandle tx_handle,
    uint64_t node_id,
    Node* node,
    ErrorInfo* error
);

// 删除节点
int gograph_node_delete(
    TransactionHandle tx_handle,
    uint64_t node_id,
    ErrorInfo* error
);
```

### 关系操作

```c
// 创建关系
int gograph_relationship_create(
    TransactionHandle tx_handle,
    const char* rel_type,
    uint64_t start_node_id,
    uint64_t end_node_id,
    const char* properties,
    uint64_t* rel_id,
    ErrorInfo* error
);

// 查询关系
int gograph_relationship_get(
    TransactionHandle tx_handle,
    uint64_t rel_id,
    Relationship* rel,
    ErrorInfo* error
);
```

## 💡 使用示例

### Python

```python
import gograph

# 初始化错误结构
error = gograph.ErrorInfo()

# 创建数据库
db_handle = gograph.gograph_database_new(
    b"mygraph.db".decode('utf-8'), 
    error
)

if db_handle is None:
    raise Exception(f"Failed: {error.message}")

# 开始事务
tx_handle = gograph.gograph_transaction_begin(db_handle, False, error)

# 执行查询
result = gograph.QueryResult()
ret = gograph.gograph_cypher_execute(
    tx_handle,
    "MATCH (n) RETURN count(n)",
    None,
    result,
    error
)

if ret == gograph.GOGRAPH_OK:
    print(f"Query succeeded: {result.row_count} rows")

# 提交事务
gograph.gograph_transaction_commit(tx_handle, error)

# 清理资源
gograph.gograph_database_free(db_handle)
```

### Java (示例)

```java
import org.gograph.binding.*;

public class Example {
    public static void main(String[] args) {
        // 加载本地库
        System.loadLibrary("gograph_java");
        
        // 创建数据库
        gograph.ErrorInfo error = new gograph.ErrorInfo();
        gograph.DatabaseHandle db = gograph.gograph_database_new(
            "mygraph.db", 
            error
        );
        
        // 使用数据库...
        
        // 清理
        gograph.gograph_database_free(db.swigValue());
    }
}
```

### C# (示例)

```csharp
using System;
using GoGraph.Binding;

class Program
{
    static void Main()
    {
        // 创建数据库
        ErrorInfo error = new ErrorInfo();
        DatabaseHandle db = gograph.gograph_database_new(
            "mygraph.db", 
            error
        );
        
        // 使用数据库...
        
        // 清理
        gograph.gograph_database_free(db);
    }
}
```

## 🔧 构建选项

### 构建所有语言绑定

```bash
./build.sh all
```

### 只构建特定语言

```bash
# Python
./build.sh python

# Java
./build.sh java

# C#
./build.sh csharp
```

### 清理构建产物

```bash
./build.sh clean
```

## 📋 错误码

| 错误码 | 值 | 说明 |
|--------|-----|------|
| `GOGRAPH_OK` | 0 | 成功 |
| `GOGRAPH_ERROR_GENERAL` | 1 | 一般错误 |
| `GOGRAPH_ERROR_INVALID_PARAM` | 2 | 无效参数 |
| `GOGRAPH_ERROR_NOT_FOUND` | 3 | 资源未找到 |
| `GOGRAPH_ERROR_ALREADY_EXISTS` | 4 | 资源已存在 |
| `GOGRAPH_ERROR_PARSE_FAILURE` | 5 | 解析失败 |
| `GOGRAPH_ERROR_EXEC_FAILURE` | 6 | 执行失败 |
| `GOGRAPH_ERROR_STORAGE_FAILURE` | 7 | 存储故障 |
| `GOGRAPH_ERROR_MEMORY_ALLOC` | 8 | 内存分配失败 |
| `GOGRAPH_ERROR_TX_CONFLICT` | 9 | 事务冲突 |

## ⚠️ 重要提示

### 内存管理

C 侧分配的内存必须手动释放：

```python
# 正确做法
try:
    result = gograph.QueryResult()
    gograph.gograph_cypher_execute(tx, query, None, result, error)
    # 使用 result...
finally:
    gograph.gograph_query_result_free(result)
```

### 字符串编码

所有字符串必须使用 UTF-8 编码：

```python
# Python 2/3 兼容
db_path = "mygraph.db".encode('utf-8').decode('utf-8')
```

### 线程安全

- ✅ 数据库句柄：线程安全
- ✅ 事务句柄：**不**线程安全（每个线程使用独立事务）
- ✅ 注册表：内部使用互斥锁保护

## 📚 文档

- [构建指南](BUILD_GUIDE.md) - 详细的构建说明
- [Go 文档](https://pkg.go.dev/github.com/DotNetAge/gograph/binding)
- [SWIG 文档](http://www.swig.org/Doc.html)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE)

## 🔗 相关链接

- [GoGraph 主项目](..)
- [GoGraph 架构文档](../docs/architecture/)
- [Cypher 语法指南](../docs/cypher-syntax/)

---

**注意**: 此包是 GoGraph 项目的一部分，需要与核心包一起使用。
