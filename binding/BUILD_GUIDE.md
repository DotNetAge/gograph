# GoGraph Binding 构建指南

本指南将帮助您为 GoGraph 图数据库构建跨语言绑定（Python、Java、C# 等）。

## 📋 目录

- [环境准备](#环境准备)
- [依赖安装](#依赖安装)
- [构建绑定](#构建绑定)
- [运行示例](#运行示例)
- [故障排除](#故障排除)

---

## 环境准备

### macOS

```bash
# 安装 Xcode Command Line Tools
xcode-select --install

# 安装 Homebrew（如果尚未安装）
/bin/bash -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"

# 安装 SWIG
brew install swig

# 安装 Python 3（如果需要）
brew install python@3.9

# 获取 Python 头文件路径
python3-config --includes
```

### Ubuntu/Debian

```bash
# 更新包列表
sudo apt-get update

# 安装构建工具
sudo apt-get install -y build-essential

# 安装 SWIG
sudo apt-get install -y swig

# 安装 Python 开发包
sudo apt-get install -y python3-dev python3-pip

# 验证安装
swig -version
python3-config --includes
```

### Windows (WSL)

```bash
# 在 WSL 中安装
sudo apt-get update
sudo apt-get install -y build-essential swig python3-dev

# 或者使用 MinGW
# 下载并安装 MSYS2: https://www.msys2.org/
# 在 MSYS2 中安装：pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-swig
```

---

## 依赖安装

### 检查依赖

运行以下命令检查是否已安装所有必需的依赖：

```bash
cd gograph/binding
./build.sh help
```

### 安装 Python 依赖（可选）

如果您只需要 Python 绑定，确保安装了 Python 3.7+：

```bash
python3 --version  # 应该 >= 3.7
pip3 install --upgrade pip
```

---

## 构建绑定

### 方法一：使用构建脚本（推荐）

#### 一键构建所有绑定

```bash
cd gograph/binding
./build.sh all
```

#### 只构建 Python 绑定

```bash
./build.sh python
```

#### 构建 Java 绑定

```bash
./build.sh java
```

#### 构建 C# 绑定

```bash
./build.sh csharp
```

#### 清理构建产物

```bash
./build.sh clean
```

### 方法二：使用 Makefile

#### 构建所有绑定

```bash
cd gograph/binding
make
```

#### 只构建 Python 绑定

```bash
make python
```

#### 构建 Java 绑定

```bash
make java
```

#### 构建 C# 绑定

```bash
make csharp
```

#### 清理

```bash
make clean
```

### 方法三：手动构建

#### 1. 构建 Go 静态库

```bash
cd gograph/binding
go build -buildmode=c-archive -o libgograph_c.a .
```

这将生成：
- `libgograph_c.a` - Go 静态库
- `libgograph_c.h` - C 头文件（自动生成）

#### 2. 生成 SWIG 包装

```bash
# Python
swig -cgo -python -py3 -o gograph_wrap.cxx gograph.i

# Java
swig -cgo -java -package org.gograph.binding -o gograph_wrap_java.cxx gograph.i

# C#
swig -cgo -csharp -namespace GoGraph.Binding -o gograph_wrap_csharp.cxx gograph.i
```

#### 3. 编译共享库

**Python:**

```bash
g++ -fPIC -shared \
    gograph_wrap.cxx \
    -o _gograph.so \
    $(python3-config --includes) \
    -L. -lgograph_c \
    $(python3-config --ldflags) \
    -Wl,-rpath,./
```

**Java:**

```bash
g++ -fPIC -shared \
    gograph_wrap_java.cxx \
    -o libgograph_java.so \
    -L. -lgograph_c \
    -ljvm
```

---

## 运行示例

### Python 示例

```bash
# 确保已构建 Python 绑定
cd gograph/binding
./build.sh python

# 运行示例
python3 examples/simple_example.py
```

预期输出：

```
============================================================
GoGraph Python Binding Example
============================================================

Step 1: Creating database...
✓ Database created at: gograph_example.db

Step 2: Beginning transaction...
✓ Transaction started (read-write)

Step 3: Executing Cypher query...
✓ Query executed successfully
  Columns: 1
  Rows: 1

Step 4: Committing transaction...
✓ Transaction committed

Step 5: Cleaning up resources...
✓ Resources freed

============================================================
Example completed successfully!
============================================================
```

### 测试导入

```python
import sys
sys.path.insert(0, 'binding')
import gograph

print("GoGraph version:", gograph.__version__ if hasattr(gograph, '__version__') else "N/A")
```

---

## 故障排除

### 问题 1: SWIG 未找到

**错误信息:**
```
./build.sh: line XX: swig: command not found
```

**解决方案:**
```bash
# macOS
brew install swig

# Ubuntu/Debian
sudo apt-get install swig
```

### 问题 2: Python 头文件未找到

**错误信息:**
```
fatal error: Python.h: No such file or directory
```

**解决方案:**
```bash
# macOS
brew install python@3.9

# Ubuntu/Debian
sudo apt-get install python3-dev
```

### 问题 3: 链接错误

**错误信息:**
```
undefined symbol: govector_storage_new
```

**解决方案:**
确保 Go 静态库已正确构建并在链接时被包含：
```bash
ls -la libgograph_c.a  # 确认文件存在
ldd _gograph.so        # 检查动态库依赖
```

### 问题 4: 运行时找不到库

**错误信息:**
```
ImportError: libgograph_c.a: cannot open shared object file
```

**解决方案:**
设置 LD_LIBRARY_PATH（Linux）或 DYLD_LIBRARY_PATH（macOS）：
```bash
# Linux
export LD_LIBRARY_PATH=$PWD:$LD_LIBRARY_PATH

# macOS
export DYLD_LIBRARY_PATH=$PWD:$DYLD_LIBRARY_PATH
```

或者在编译时使用 `-Wl,-rpath,./` 参数。

### 问题 5: Go 编译错误

**错误信息:**
```
undefined: storage.PebbleDB
```

**解决方案:**
确保在 gograph 项目根目录下构建：
```bash
cd /Users/ray/workspaces/ai-ecosystem/gograph
go mod tidy
cd binding
./build.sh python
```

---

## 构建产物

成功构建后，`binding/` 目录将包含以下文件：

### 核心文件
- `gograph_c.h` - C API 头文件
- `gograph_c.go` - Go 导出实现
- `gograph.i` - SWIG 接口文件
- `libgograph_c.a` - Go 静态库

### Python 绑定
- `_gograph.so` - Python 共享库（Linux/macOS）
- `_gograph.pyd` - Python DLL（Windows）
- `gograph.py` - Python 包装模块
- `gograph_wrap.cxx` - SWIG 生成的 C++ 代码

### Java 绑定
- `libgograph_java.so` - Java 本地库
- `gographJAVA.java` - Java 包装类
- `gograph_wrap_java.cxx` - SWIG 生成的 C++ 代码

### C# 绑定
- `gographCSharp.cs` - C# 包装类
- `gograph_wrap_csharp.cxx` - SWIG 生成的 C++ 代码

---

## 架构说明

### 目录结构

```
gograph/
├── pkg/                    # 纯 Go 核心包（无 CGO）
│   ├── api/               # API 层
│   ├── cypher/            # Cypher 解析器
│   ├── graph/             # 图数据结构
│   ├── storage/           # 存储引擎
│   └── tx/                # 事务管理
├── binding/               # CGO 绑定层
│   ├── gograph_c.h       # C API 头文件
│   ├── gograph_c.go      # Go 导出实现
│   ├── gograph.i         # SWIG 接口文件
│   ├── build.sh          # 构建脚本
│   ├── Makefile          # Make 配置
│   └── examples/         # 示例代码
└── cmd/                   # 命令行工具
```

### 依赖关系

```
binding (CGO 层)
    ↓
pkg/* (纯 Go 核心)
```

**关键原则:**
- ✅ `binding` 可以依赖 `pkg`
- ❌ `pkg` 不能依赖 `binding`
- ✅ `pkg` 保持纯净，无 `import "C"`
- ✅ `binding` 专门处理 CGO 和跨语言交互

---

## 下一步

构建成功后，您可以：

1. **查看 API 文档**: 
   ```bash
   godoc github.com/DotNetAge/gograph/binding
   ```

2. **运行更多示例**:
   ```bash
   python3 examples/simple_example.py
   ```

3. **开始开发**: 参考 `examples/` 目录中的示例代码开始您的项目。

4. **性能优化**: 对于生产环境，考虑：
   - 启用 Go 的优化标志：`go build -O2`
   - 使用性能分析工具：`pprof`
   - 调整内存管理策略

---

## 相关资源

- [SWIG 官方文档](http://www.swig.org/Doc.html)
- [Go CGO 教程](https://golang.org/cmd/cgo/)
- [Python C API](https://docs.python.org/3/c-api/)
- [GoGraph 架构文档](../docs/architecture/)

---

**提示**: 如果遇到任何问题，请检查：
1. 所有依赖是否正确安装
2. 环境变量是否设置正确
3. Go 模块依赖是否完整 (`go mod tidy`)
4. 编译器版本是否兼容

祝您构建顺利！🚀
