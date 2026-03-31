# GoGraph Binding Makefile
# 
# 使用方法:
#   make          - 构建所有绑定
#   make python   - 只构建 Python 绑定
#   make clean    - 清理构建产物

.PHONY: all python java csharp clean help

# 默认目标
all: python

# Python 绑定
python: libgograph_c.a gograph_wrap.cxx _gograph.so

# Java 绑定
java: libgograph_c.a gograph_wrap_java.cxx

# C# 绑定
csharp: libgograph_c.a gograph_wrap_csharp.cxx

# 构建 Go 静态库
libgograph_c.a: *.go
	@echo "Building Go static library..."
	cd .. && go build -buildmode=c-archive -o binding/libgograph_c.a ./binding
	@echo "✓ Built: libgograph_c.a"

# 生成 Python SWIG 包装
gograph_wrap.cxx: gograph.i gograph_c.h
	@echo "Generating SWIG wrapper for Python..."
	swig -cgo -python -py3 -o gograph_wrap.cxx gograph.i
	@echo "✓ Generated: gograph_wrap.cxx"

# 生成 Java SWIG 包装
gograph_wrap_java.cxx: gograph.i gograph_c.h
	@echo "Generating SWIG wrapper for Java..."
	swig -cgo -java -package org.gograph.binding -o gograph_wrap_java.cxx gograph.i
	@echo "✓ Generated: gograph_wrap_java.cxx"

# 生成 C# SWIG 包装
gograph_wrap_csharp.cxx: gograph.i gograph_c.h
	@echo "Generating SWIG wrapper for C#..."
	swig -cgo -csharp -namespace GoGraph.Binding -o gograph_wrap_csharp.cxx gograph.i
	@echo "✓ Generated: gograph_wrap_csharp.cxx"

# 编译 Python 模块
_gograph.so: gograph_wrap.cxx libgograph_c.a
	@echo "Building Python module..."
	g++ -fPIC -shared \
		gograph_wrap.cxx \
		-o _gograph.so \
		$(shell python3-config --includes) \
		-L. -lgograph_c \
		$(shell python3-config --ldflags) \
		-Wl,-rpath,./
	@echo "✓ Built: _gograph.so"

# 清理
clean:
	@echo "Cleaning..."
	rm -f *.a *.so *.dylib *.dll
	rm -f gograph_wrap*.cxx
	rm -f gograph.py gograph_wrap.c
	rm -f *.class
	@echo "✓ Cleaned."

# 帮助
help:
	@echo "GoGraph Binding Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all       - Build all bindings (default)"
	@echo "  python    - Build Python bindings"
	@echo "  java      - Build Java bindings"
	@echo "  csharp    - Build C# bindings"
	@echo "  clean     - Clean build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make              # Build everything"
	@echo "  make python       # Build Python only"
	@echo "  make clean        # Clean up"
