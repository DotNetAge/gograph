#!/bin/bash

# GoGraph Binding Build Script
# 一键构建所有绑定
#
# 使用方法:
#   ./build.sh [python|java|csharp|all]
#
# 示例:
#   ./build.sh python    # 只构建 Python 绑定
#   ./build.sh all       # 构建所有绑定

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    print_info "Checking dependencies..."
    
    if ! command -v swig &> /dev/null; then
        print_error "SWIG is not installed."
        print_info "Install SWIG:"
        print_info "  macOS:   brew install swig"
        print_info "  Ubuntu:  sudo apt-get install swig"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed."
        exit 1
    fi
    
    # 检查 Python（如果构建 Python 绑定）
    if [[ "$1" == "python" || "$1" == "all" ]]; then
        if ! command -v python3 &> /dev/null; then
            print_warning "Python 3 not found. Skipping Python bindings."
            SKIP_PYTHON=true
        else
            SKIP_PYTHON=false
        fi
    fi
    
    print_info "All dependencies are available."
}

# 进入项目根目录
cd "$(dirname "$0")/.."

# 构建 Go 静态库
build_go_lib() {
    print_info "Building Go static library..."
    
    cd binding
    go build -buildmode=c-archive -o libgograph_c.a ./...
    
    if [ $? -eq 0 ]; then
        print_info "✓ Go static library built: binding/libgograph_c.a"
    else
        print_error "Failed to build Go library."
        exit 1
    fi
    
    cd ..
}

# 生成 SWIG 包装
generate_swig_wrapper() {
    local lang=$1
    print_info "Generating SWIG wrapper for ${lang}..."
    
    cd binding
    
    case $lang in
        python)
            if [ "$SKIP_PYTHON" = true ]; then
                print_warning "Skipping Python bindings (Python not found)"
                cd ..
                return
            fi
            swig -cgo -python -py3 -o gograph_wrap.cxx gograph.i
            print_info "✓ Python wrapper generated: gograph/gograph_wrap.cxx"
            ;;
        
        java)
            swig -cgo -java -package org.gograph.binding -o gograph_wrap.cxx gograph.i
            print_info "✓ Java wrapper generated: gograph/gograph_wrap.cxx"
            ;;
        
        csharp)
            swig -cgo -csharp -namespace GoGraph.Binding -o gograph_wrap.cxx gograph.i
            print_info "✓ C# wrapper generated: gograph/gograph_wrap.cxx"
            ;;
        
        *)
            print_error "Unsupported language: $lang"
            cd ..
            return 1
            ;;
    esac
    
    cd ..
}

# 编译 Python 模块
build_python_module() {
    if [ "$SKIP_PYTHON" = true ]; then
        return
    fi
    
    print_info "Building Python module..."
    
    cd binding
    
    # 获取 Python 配置
    PYTHON_INCLUDE=$(python3-config --includes 2>/dev/null || echo "-I/usr/include/python3.9")
    PYTHON_LIBS=$(python3-config --ldflags 2>/dev/null || echo "-lpython3.9")
    
    # 编译共享库
    g++ -fPIC -shared \
        gograph_wrap.cxx \
        -o _gograph.so \
        ${PYTHON_INCLUDE} \
        -L. -lgograph_c \
        ${PYTHON_LIBS} \
        -Wl,-rpath,./
    
    if [ $? -eq 0 ]; then
        print_info "✓ Python module built: binding/_gograph.so"
    else
        print_error "Failed to build Python module."
        cd ..
        return 1
    fi
    
    cd ..
}

# 编译 Java 模块
build_java_module() {
    print_info "Building Java module..."
    
    cd binding
    
    # 编译 C++ 代码为共享库
    g++ -fPIC -shared \
        gograph_wrap.cxx \
        -o libgograph_java.so \
        -L. -lgograph_c \
        -ljvm
    
    if [ $? -eq 0 ]; then
        print_info "✓ Java native library built: binding/libgograph_java.so"
    else
        print_error "Failed to build Java module."
        cd ..
        return 1
    fi
    
    cd ..
}

# 清理构建产物
clean() {
    print_info "Cleaning build artifacts..."
    
    cd binding
    rm -f *.a *.so *.dylib *.dll
    rm -f gograph_wrap.cxx
    rm -f gograph.py gograph_wrap.c
    rm -f *.class
    cd ..
    
    print_info "✓ Cleaned."
}

# 显示帮助
show_help() {
    echo "GoGraph Binding Build Script"
    echo ""
    echo "Usage:"
    echo "  ./build.sh [language]"
    echo ""
    echo "Languages:"
    echo "  python    Build Python bindings"
    echo "  java      Build Java bindings"
    echo "  csharp    Build C# bindings"
    echo "  all       Build all bindings"
    echo "  clean     Clean build artifacts"
    echo ""
    echo "Examples:"
    echo "  ./build.sh python    # Build Python only"
    echo "  ./build.sh all       # Build everything"
    echo "  ./build.sh clean     # Clean up"
}

# 主函数
main() {
    local target="${1:-all}"
    
    case $target in
        python)
            check_dependencies python
            build_go_lib
            generate_swig_wrapper python
            build_python_module
            ;;
        
        java)
            check_dependencies
            build_go_lib
            generate_swig_wrapper java
            build_java_module
            ;;
        
        csharp)
            check_dependencies
            build_go_lib
            generate_swig_wrapper csharp
            # C# compilation would require additional steps
            print_info "C# wrapper generated. Use Visual Studio to compile."
            ;;
        
        all)
            check_dependencies python
            build_go_lib
            generate_swig_wrapper python
            build_python_module
            # Add other languages as needed
            ;;
        
        clean)
            clean
            ;;
        
        help|--help|-h)
            show_help
            ;;
        
        *)
            print_error "Unknown target: $target"
            show_help
            exit 1
            ;;
    esac
    
    print_info "Build completed successfully!"
}

# 运行主函数
main "$@"
