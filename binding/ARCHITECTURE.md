# GoGraph CGO 隔离架构设计

本文档描述了 GoGraph 项目中 CGO/SWIG绑定的架构隔离设计，确保核心包保持纯净的同时提供跨语言互操作能力。

## 📋 目录

- [设计目标](#设计目标)
- [架构原则](#架构原则)
- [目录结构](#目录结构)
- [依赖关系](#依赖关系)
- [实现细节](#实现细节)
- [最佳实践](#最佳实践)

---

## 设计目标

### 1. 保持核心包纯净

**问题**: CGO 代码会引入编译依赖，影响纯 Go 环境下的构建速度和可移植性。

**目标**: 
- ✅ `pkg/*` 目录下的核心包不包含任何 `import "C"` 语句
- ✅ 核心包可以在无 CGO 环境下正常编译和测试
- ✅ 保持 Go 代码的可读性和可维护性

### 2. 单向依赖

**问题**: 循环依赖会导致代码难以理解和维护。

**目标**:
- ✅ `binding` → `core`：允许
- ❌ `core` → `binding`：禁止
- ✅ 清晰的依赖层次结构

### 3. 构建隔离

**问题**: 不是所有用户都需要跨语言绑定。

**目标**:
- ✅ 可以选择性构建 CGO 组件
- ✅ 不构建绑定时不影响核心功能
- ✅ 独立的构建配置和脚本

---

## 架构原则

### 分层架构

```
┌────────────────────────────────────────────┐
│   External Languages (Python/Java/C#)     │  ← 外部语言层
└─────────────────┬──────────────────────────┘
                  │
                  ↓ SWIG Bindings
┌────────────────────────────────────────────┐
│   binding/ (CGO Adapter Layer)             │  ← 适配层
│   ├── gograph_c.h (C API)                 │
│   ├── gograph_c.go (Go Implementation)    │
│   └── gograph.i (SWIG Interface)          │
└─────────────────┬──────────────────────────┘
                  │
                  ↓ Pure Go
┌────────────────────────────────────────────┐
│   pkg/* (Core Implementation)              │  ← 核心层
│   ├── api/        (API 层)                │
│   ├── cypher/     (Cypher 解析)           │
│   ├── graph/      (图数据结构)            │
│   ├── storage/    (存储引擎)              │
│   └── tx/         (事务管理)              │
└────────────────────────────────────────────┘
```

### 接口契约

**binding 包职责**:
1. 定义 C API 接口（`gograph_c.h`）
2. 实现 Go-C 桥接代码（`gograph_c.go`）
3. 配置 SWIG 生成规则（`gograph.i`）
4. 管理跨语言类型转换
5. 处理内存分配和释放

**core 包职责**:
1. 实现核心业务逻辑
2. 提供纯 Go 的 API
3. 保持技术栈纯净（无 CGO）
4. 专注于性能和正确性

---

## 目录结构

### 完整的项目结构

```
gograph/
├── cmd/                          # 命令行工具
│   └── gograph/
│       ├── main.go
│       ├── commands.go
│       └── tui.go
│
├── pkg/                          # 核心包（纯 Go）
│   ├── api/                      # API 层
│   │   ├── db.go
│   │   ├── graphstore.go
│   │   └── result.go
│   │
│   ├── cypher/                   # Cypher 解析器
│   │   ├── lexer/
│   │   ├── parser/
│   │   ├── ast/
│   │   └── executor.go
│   │
│   ├── graph/                    # 图数据结构
│   │   ├── graph.go
│   │   ├── node.go
│   │   ├── relationship.go
│   │   └── property.go
│   │
│   ├── storage/                  # 存储引擎
│   │   ├── pebble.go
│   │   ├── keys.go
│   │   └── marshal.go
│   │
│   └── tx/                       # 事务管理
│       ├── manager.go
│       └── tx.go
│
├── binding/                      # CGO 绑定层
│   ├── gograph_c.h              # C API 头文件
│   ├── gograph_c.go             # Go 导出实现
│   ├── gograph.i                # SWIG 接口文件
│   ├── build.sh                 # 构建脚本
│   ├── Makefile                 # Make 配置
│   ├── BUILD_GUIDE.md           # 构建指南
│   ├── README.md                # 包文档
│   ├── ARCHITECTURE.md          # 架构设计（本文件）
│   │
│   └── examples/                # 示例代码
│       └── simple_example.py    # Python 示例
│
├── docs/                         # 项目文档
│   ├── architecture/
│   ├── cypher-syntax/
│   └── developer-guide/
│
└── scripts/                      # 辅助脚本
    └── ...
```

### binding 目录详细结构

```
binding/
├── gograph_c.h              # C 语言接口定义
│   ├── 类型定义             # ValueType, Node, Relationship 等
│   ├── 错误码定义           # GOGRAPH_ERROR_*
│   ├── 数据库管理 API       # gograph_database_*
│   ├── 事务管理 API         # gograph_transaction_*
│   ├── Cypher 查询 API      # gograph_cypher_execute
│   ├── 节点操作 API         # gograph_node_*
│   ├── 关系操作 API         # gograph_relationship_*
│   └── 内存管理 API         # gograph_*_free
│
├── gograph_c.go             # Go 实现
│   ├── 全局注册表           # databaseRegistry, transactionRegistry
│   ├── 类型转换函数         # valueToC, newError
│   ├── 数据库 API 实现      # export gograph_database_*
│   ├── 事务 API 实现        # export gograph_transaction_*
│   ├── 查询 API 实现        # export gograph_cypher_execute
│   ├── 节点 API 实现        # export gograph_node_*
│   ├── 关系 API 实现        # export gograph_relationship_*
│   └── 内存管理实现         # export gograph_*_free
│
├── gograph.i                # SWIG 配置
│   ├── 模块声明             # %module gograph
│   ├── 类型映射             # %typemap
│   ├── Python 定制          # %pythoncode
│   └── 文档字符串           # %feature("docstring")
│
├── build.sh                 # 一键构建脚本
│   ├── check_dependencies() # 检查依赖
│   ├── build_go_lib()       # 构建 Go 静态库
│   ├── generate_swig_wrapper() # 生成 SWIG 包装
│   └── build_python_module()   # 编译 Python 模块
│
├── Makefile                 # Make 配置
│   ├── all: python          # 默认目标
│   ├── libgograph_c.a       # Go 静态库
│   ├── gograph_wrap.cxx     # SWIG 包装
│   └── _gograph.so          # Python 共享库
│
└── examples/
    └── simple_example.py    # Python 使用示例
```

---

## 依赖关系

### 允许的依赖

```go
// ✅ binding/gograph_c.go
package binding

import "C"

import (
    "github.com/DotNetAge/gograph/pkg/storage"  // OK: binding → core
    "github.com/DotNetAge/gograph/pkg/tx"       // OK: binding → core
)
```

### 禁止的依赖

```go
// ❌ pkg/storage/pebble.go
package storage

import (
    "github.com/DotNetAge/gograph/binding"  // ERROR: core → binding 禁止！
)
```

### 依赖图

```
         External (Python/Java/C#)
                    ↓
              binding (CGO)
                    ↓
    ┌───────────────┴───────────────┐
    ↓               ↓               ↓
  api           cypher          graph
    ↓               ↓               ↓
  storage  ←→   tx  ←→   (其他核心包)
    
实线箭头：允许的依赖
虚线箭头：禁止的依赖
```

---

## 实现细节

### 1. 句柄管理

使用全局注册表管理 C 侧的不透明句柄：

```go
// binding/gograph_c.go
var (
    databaseRegistry   = make(map[uint64]*storage.PebbleDB)
    transactionRegistry = make(map[uint64]*tx.Transaction)
    registryMu         sync.RWMutex
    
    nextDatabaseID   uint64 = 1
    nextTransactionID uint64 = 1
)

//export gograph_database_new
func gograph_database_new(dbPath *C.char, errInfo *C.ErrorInfo) C.DatabaseHandle {
    // 1. 创建 Go 对象
    db, err := storage.NewPebbleDB(C.GoString(dbPath))
    
    // 2. 注册句柄
    registryMu.Lock()
    handle := nextDatabaseID
    nextDatabaseID++
    databaseRegistry[handle] = db
    registryMu.Unlock()
    
    // 3. 返回句柄给 C
    return C.DatabaseHandle(handle)
}

//export gograph_database_free
func gograph_database_free(handle C.DatabaseHandle) {
    registryMu.Lock()
    defer registryMu.Unlock()
    
    delete(databaseRegistry, uint64(handle))
}
```

### 2. 类型转换

#### Go → C

```go
func valueToC(goVal interface{}) C.Value {
    cVal := C.Value{type_: C.VALUE_NULL}
    
    switch v := goVal.(type) {
    case bool:
        cVal.type_ = C.VALUE_BOOL
        cVal.data.bool_val = C.bool(v)
    case int64:
        cVal.type_ = C.VALUE_INT
        cVal.data.int_val = C.int64_t(v)
    case string:
        cVal.type_ = C.VALUE_STRING
        cVal.data.string_val = C.CString(v)
    }
    
    return cVal
}
```

#### C → Go

```go
//export gograph_node_create
func gograph_node_create(
    txHandle C.TransactionHandle,
    label *C.char,
    properties *C.char,
    nodeId *C.uint64_t,
    errInfo *C.ErrorInfo,
) C.int {
    // 1. 转换 C 字符串为 Go 字符串
    labelStr := C.GoString(label)
    propsStr := C.GoString(properties)
    
    // 2. 解析 JSON 属性
    var props map[string]interface{}
    json.Unmarshal([]byte(propsStr), &props)
    
    // 3. 调用核心 Go 代码
    tx := transactionRegistry[uint64(txHandle)]
    nodeID, err := tx.CreateNode(labelStr, props)
    
    // 4. 返回结果给 C
    if err != nil {
        *errInfo = newError(C.GOGRAPH_ERROR_EXEC_FAILURE, err.Error())
        return C.GOGRAPH_ERROR_EXEC_FAILURE
    }
    
    *nodeId = C.uint64_t(nodeID)
    return C.GOGRAPH_OK
}
```

### 3. 内存管理

#### C 侧分配的内存

```c
// gograph_c.h
Value* gograph_value_new();  // C 侧分配
void gograph_value_free(Value* value);  // C 侧释放
```

#### Go 侧分配的内存

```go
// gograph_c.go
//export gograph_database_new
func gograph_database_new(...) C.DatabaseHandle {
    // Go 管理的对象，通过句柄间接管理
    db := &storage.PebbleDB{}
    // ...
}

//export gograph_database_free
func gograph_database_free(handle C.DatabaseHandle) {
    // Go GC 会自动回收，这里只需清理注册表
    delete(databaseRegistry, uint64(handle))
}
```

### 4. 错误处理

```go
func newError(code C.int, msg string) C.ErrorInfo {
    return C.ErrorInfo{
        code:    code,
        message: C.CString(msg),
    }
}

//export gograph_cypher_execute
func gograph_cypher_execute(...) C.int {
    tx := transactionRegistry[uint64(txHandle)]
    if tx == nil {
        if errInfo != nil {
            *errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
        }
        return C.GOGRAPH_ERROR_NOT_FOUND
    }
    
    // 执行查询...
    if err != nil {
        if errInfo != nil {
            *errInfo = newError(C.GOGRAPH_ERROR_EXEC_FAILURE, err.Error())
        }
        return C.GOGRAPH_ERROR_EXEC_FAILURE
    }
    
    return C.GOGRAPH_OK
}
```

---

## 最佳实践

### 1. 命名规范

**C API 命名**:
```c
// 格式：<module>_<function>_<action>
gograph_database_new()        // 创建
gograph_database_open()       // 打开
gograph_database_close()      // 关闭
gograph_database_free()       // 释放

gograph_transaction_begin()   // 开始
gograph_transaction_commit()  // 提交
gograph_transaction_rollback()// 回滚

gograph_node_create()         // 创建节点
gograph_node_get()            // 获取节点
gograph_node_delete()         // 删除节点
```

**Go 函数命名**:
```go
// export 函数必须与 C API 名称一致
//export gograph_database_new
func gograph_database_new(...) C.DatabaseHandle {
    // ...
}

// 内部辅助函数使用驼峰命名
func newError(code C.int, msg string) C.ErrorInfo {
    // ...
}
```

### 2. 注释规范

```go
// gograph_c.h
/**
 * @brief 创建新的数据库实例
 * 
 * @param db_path 数据库文件路径（例如："gograph.db"）
 * @param error 输出错误信息（可为 NULL）
 * @return DatabaseHandle 数据库句柄，失败返回 NULL
 */
DatabaseHandle gograph_database_new(const char* db_path, ErrorInfo* error);
```

### 3. 资源清理模式

```python
# Python 使用示例 - 使用 try-finally 确保资源释放
error = gograph.ErrorInfo()
db_handle = None

try:
    db_handle = gograph.gograph_database_new(b"test.db".decode(), error)
    if db_handle is None:
        raise Exception("Failed to create database")
    
    # 使用数据库...
    
finally:
    # 总是清理资源
    if db_handle is not None:
        gograph.gograph_database_free(db_handle)
```

### 4. 测试策略

**核心包测试**（无 CGO）:
```bash
cd gograph/pkg/storage
go test -v ./...
```

**binding 包测试**（需要 CGO）:
```bash
cd gograph/binding
go test -tags=cgo -v ./...
```

**集成测试**（Python）:
```bash
cd gograph/binding
./build.sh python
python3 tests/test_integration.py
```

### 5. 性能优化

**减少跨语言调用**:
```python
# ❌ 低效：多次跨语言调用
for i in range(1000):
    gograph.gograph_node_create(tx, "Label", f'{{"id": {i}}}', ...)

# ✅ 高效：批量操作
nodes_data = [{"label": "Label", "props": {"id": i}} for i in range(1000)]
gograph.gograph_node_create_batch(tx, nodes_data, ...)  # 一次调用
```

**内存池**（可选）:
```go
var valuePool = sync.Pool{
    New: func() interface{} {
        return &C.Value{}
    },
}

func getValueFromPool() *C.Value {
    return valuePool.Get().(*C.Value)
}

func returnValueToPool(val *C.Value) {
    *val = C.Value{} // 清零
    valuePool.Put(val)
}
```

---

## 总结

### 核心优势

1. **架构清晰**: binding 层和 core 层职责分离
2. **易于维护**: 单向依赖，无循环依赖
3. **灵活部署**: 可选择性构建 CGO 组件
4. **性能优良**: 直接映射到 Go 核心，无中间层开销
5. **类型安全**: 完整的类型检查和错误处理

### 适用场景

- ✅ 需要多语言互操作的 Go 项目
- ✅ 希望保持核心包纯净的库
- ✅ 需要高性能 C API 的场景
- ✅ 长期维护的大型项目

### 不适用场景

- ❌ 简单的单语言项目（过度设计）
- ❌ 快速原型开发（增加复杂度）
- ❌ 对二进制大小敏感的场景（增加体积）

---

**参考**: 此架构设计参考了 govector 项目的 CGO 隔离方案，并针对 GoGraph 的图数据库特性进行了优化。
