#!/usr/bin/env python3
"""
GoGraph Python 绑定简单示例

这个示例演示了如何使用 GoGraph 的 Python 绑定进行基本的图操作。

使用方法:
    python3 examples/simple_example.py
"""

import sys
import os

# 添加 binding 目录到路径
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

try:
    import gograph
except ImportError as e:
    print(f"Error importing gograph: {e}")
    print("Make sure you have built the bindings first:")
    print("  cd binding")
    print("  ./build.sh python")
    sys.exit(1)


def main():
    """主函数 - 演示基本用法"""
    
    print("=" * 60)
    print("GoGraph Python Binding Example")
    print("=" * 60)
    print()
    
    # 初始化错误结构
    error = gograph.ErrorInfo()
    
    try:
        # 1. 创建数据库
        print("Step 1: Creating database...")
        db_path = b"gograph_example.db".decode('utf-8')
        db_handle = gograph.gograph_database_new(db_path, error)
        
        if db_handle is None:
            print(f"Failed to create database: {error.message}")
            return
        
        print(f"✓ Database created at: {db_path}")
        print()
        
        # 2. 开始事务
        print("Step 2: Beginning transaction...")
        tx_handle = gograph.gograph_transaction_begin(db_handle, False, error)
        
        if tx_handle is None:
            print(f"Failed to begin transaction: {error.message}")
            gograph.gograph_database_free(db_handle)
            return
        
        print("✓ Transaction started (read-write)")
        print()
        
        # 3. 执行 Cypher 查询
        print("Step 3: Executing Cypher query...")
        query = b"MATCH (n) RETURN count(n) as count".decode('utf-8')
        result = gograph.QueryResult()
        
        ret = gograph.gograph_cypher_execute(
            tx_handle,
            query,
            None,  # params
            result,
            error
        )
        
        if ret == gograph.GOGRAPH_OK:
            print("✓ Query executed successfully")
            print(f"  Columns: {result.column_count}")
            print(f"  Rows: {result.row_count}")
        else:
            print(f"Query failed with code: {ret}")
        
        print()
        
        # 4. 提交事务
        print("Step 4: Committing transaction...")
        ret = gograph.gograph_transaction_commit(tx_handle, error)
        
        if ret == gograph.GOGRAPH_OK:
            print("✓ Transaction committed")
        else:
            print(f"Failed to commit: {ret}")
        
        print()
        
        # 5. 清理资源
        print("Step 5: Cleaning up resources...")
        gograph.gograph_database_free(db_handle)
        print("✓ Resources freed")
        
        print()
        print("=" * 60)
        print("Example completed successfully!")
        print("=" * 60)
        
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


def advanced_example():
    """高级示例 - 展示更多功能"""
    
    print("\n" + "=" * 60)
    print("Advanced Example: Working with Nodes and Relationships")
    print("=" * 60)
    print()
    
    error = gograph.ErrorInfo()
    
    try:
        # 打开数据库
        db_handle = gograph.gograph_database_open(b"gograph_example.db".decode('utf-8'), error)
        if db_handle is None:
            print("Failed to open database")
            return
        
        # 开始事务
        tx_handle = gograph.gograph_transaction_begin(db_handle, False, error)
        if tx_handle is None:
            print("Failed to begin transaction")
            gograph.gograph_database_free(db_handle)
            return
        
        # 创建节点
        print("Creating nodes...")
        node_id = gograph.uint64_t()
        
        # 创建 Person 节点
        properties = '{"name": "Alice", "age": 30}'
        ret = gograph.gograph_node_create(
            tx_handle,
            b"Person".decode('utf-8'),
            properties.encode('utf-8'),
            node_id,
            error
        )
        
        if ret == gograph.GOGRAPH_OK:
            print(f"✓ Created Person node with ID: {node_id}")
        else:
            print(f"Failed to create node: {ret}")
        
        # 提交事务
        gograph.gograph_transaction_commit(tx_handle, error)
        
        # 开始新的事务进行查询
        tx_handle2 = gograph.gograph_transaction_begin(db_handle, True, error)
        
        # 查询节点
        print("Querying nodes...")
        query = "MATCH (p:Person) RETURN p.name, p.age"
        result = gograph.QueryResult()
        
        ret = gograph.gograph_cypher_execute(
            tx_handle2,
            query,
            None,
            result,
            error
        )
        
        if ret == gograph.GOGRAPH_OK:
            print(f"✓ Query returned {result.row_count} rows")
        
        # 清理
        gograph.gograph_transaction_free(tx_handle)
        gograph.gograph_transaction_free(tx_handle2)
        gograph.gograph_database_free(db_handle)
        
        print("✓ Advanced example completed")
        
    except Exception as e:
        print(f"\n❌ Advanced example error: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    # 运行简单示例
    main()
    
    # 可选：运行高级示例
    # advanced_example()
