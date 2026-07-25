# GoGraph Binding Makefile
# 
# 使用方法:
#   make          - 构建所有绑定
#   make python   - 只构建 Python 绑定
#   make clean    - 清理构建产物
#   make dist     - 生成 PyPI 发布包
#   make upload   - 上传到 PyPI

.PHONY: all python java csharp clean help dist wheel sdist upload

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
	cd .. && go build -buildmode=c-archive -o binding/libgograph_c.a ./binding/cmd
	@echo "✓ Built: libgograph_c.a"

# 生成 Python SWIG 包装
swig/gograph_wrap.cxx: gograph.i gograph_c.h
	@echo "Generating SWIG wrapper for Python..."
	mkdir -p swig
	swig -python -outdir swig -o swig/gograph_wrap.cxx gograph.i
	@echo "✓ Generated: swig/gograph_wrap.cxx"

# 生成 Java SWIG 包装
swig/gograph_wrap_java.cxx: gograph.i gograph_c.h
	@echo "Generating SWIG wrapper for Java..."
	mkdir -p swig
	swig -java -package org.gograph.binding -o swig/gograph_wrap_java.cxx gograph.i
	@echo "✓ Generated: swig/gograph_wrap_java.cxx"

# 生成 C# SWIG 包装
swig/gograph_wrap_csharp.cxx: gograph.i gograph_c.h
	@echo "Generating SWIG wrapper for C#..."
	mkdir -p swig
	swig -csharp -namespace GoGraph.Binding -o swig/gograph_wrap_csharp.cxx gograph.i
	@echo "✓ Generated: swig/gograph_wrap_csharp.cxx"

# 编译 Python 模块
_gograph.so: swig/gograph_wrap.cxx libgograph_c.a
	@echo "Building Python module..."
	g++ -fPIC -shared \
		swig/gograph_wrap.cxx \
		-o _gograph.so \
		$(shell python3-config --includes) \
		-L. -lgograph_c \
		$(shell python3-config --ldflags --embed) \
		-Wl,-rpath,./
	@echo "✓ Built: _gograph.so"

# 生成发布包
dist: wheel sdist
	@echo "✓ Distribution packages generated in dist/"

# 生成 wheel 包
wheel: python
	@echo "Building wheel package..."
	python3 setup.py bdist_wheel
	@echo "✓ Built: dist/*.whl"

# 生成源码包
sdist:
	@echo "Building source distribution..."
	python3 setup.py sdist
	@echo "✓ Built: dist/*.tar.gz"

# 上传到 PyPI
upload: dist
	@echo "Uploading to PyPI..."
	twine upload dist/*
	@echo "✓ Upload completed"

# 清理
clean:
	@echo "Cleaning..."
	rm -f *.a *.so *.dylib *.dll
	rm -f gograph_wrap*.cxx
	rm -f gograph.py gograph_wrap.c
	rm -f *.class
	rm -rf build dist *.egg-info swig
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
	@echo "  wheel     - Build wheel package for PyPI"
	@echo "  sdist     - Build source distribution"
	@echo "  dist      - Build all distribution packages"
	@echo "  upload    - Upload to PyPI"
	@echo "  clean     - Clean build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make              # Build everything"
	@echo "  make python       # Build Python only"
	@echo "  make dist         # Create packages for PyPI"
	@echo "  make upload       # Upload to PyPI"
	@echo "  make clean        # Clean up"
