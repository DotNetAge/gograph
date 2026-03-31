/*
 * GoGraph C API - C 语言接口定义
 * 
 * 这是 GoGraph 图数据库的 C 语言绑定接口，
 * 允许从 Python、C++、Java 等语言调用 GoGraph 核心功能。
 * 
 * @file gograph_c.h
 * @package github.com/DotNetAge/gograph/binding
 */

#ifndef GOGRAPH_C_H
#define GOGRAPH_C_H

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ============================================================================
 * 类型定义
 * ============================================================================ */

/**
 * @brief 值类型枚举
 */
typedef enum {
    VALUE_NULL = 0,
    VALUE_BOOL = 1,
    VALUE_INT = 2,
    VALUE_FLOAT = 3,
    VALUE_STRING = 4,
    VALUE_LIST = 5,
    VALUE_MAP = 6
} ValueType;

/**
 * @brief 通用值结构
 */
typedef struct {
    ValueType type;
    union {
        bool bool_val;
        int64_t int_val;
        double float_val;
        const char* string_val;
        void* list_val;
        void* map_val;
    } data;
} Value;

/**
 * @brief 节点结构
 */
typedef struct {
    uint64_t id;
    const char* label;
    Value* properties;
    int property_count;
} Node;

/**
 * @brief 关系结构
 */
typedef struct {
    uint64_t id;
    const char* type;
    uint64_t start_node_id;
    uint64_t end_node_id;
    Value* properties;
    int property_count;
} Relationship;

/**
 * @brief 路径结构
 */
typedef struct {
    Node** nodes;
    int node_count;
    Relationship** relationships;
    int relationship_count;
} Path;

/**
 * @brief 查询结果结构
 */
typedef struct {
    const char** columns;
    int column_count;
    Value** rows;
    int row_count;
} QueryResult;

/**
 * @brief 错误信息结构
 */
typedef struct {
    int code;
    const char* message;
} ErrorInfo;

/**
 * @brief 数据库句柄（不透明指针）
 */
typedef void* DatabaseHandle;

/**
 * @brief 事务句柄（不透明指针）
 */
typedef void* TransactionHandle;

/* ============================================================================
 * 错误码定义
 * ============================================================================ */

#define GOGRAPH_OK                      0   ///< 成功
#define GOGRAPH_ERROR_GENERAL           1   ///< 一般错误
#define GOGRAPH_ERROR_INVALID_PARAM     2   ///< 无效参数
#define GOGRAPH_ERROR_NOT_FOUND         3   ///< 资源未找到
#define GOGRAPH_ERROR_ALREADY_EXISTS    4   ///< 资源已存在
#define GOGRAPH_ERROR_PARSE_FAILURE     5   ///< 解析失败
#define GOGRAPH_ERROR_EXEC_FAILURE      6   ///< 执行失败
#define GOGRAPH_ERROR_STORAGE_FAILURE   7   ///< 存储故障
#define GOGRAPH_ERROR_MEMORY_ALLOC      8   ///< 内存分配失败
#define GOGRAPH_ERROR_TX_CONFLICT       9   ///< 事务冲突

/* ============================================================================
 * 数据库管理 API
 * ============================================================================ */

/**
 * @brief 创建新的数据库实例
 * 
 * @param db_path 数据库文件路径（例如："gograph.db"）
 * @param error 输出错误信息（可为 NULL）
 * @return DatabaseHandle 数据库句柄，失败返回 NULL
 */
DatabaseHandle gograph_database_new(const char* db_path, ErrorInfo* error);

/**
 * @brief 打开已存在的数据库
 * 
 * @param db_path 数据库文件路径
 * @param error 输出错误信息
 * @return DatabaseHandle 数据库句柄，失败返回 NULL
 */
DatabaseHandle gograph_database_open(const char* db_path, ErrorInfo* error);

/**
 * @brief 关闭数据库
 * 
 * @param handle 数据库句柄
 * @return int 错误码
 */
int gograph_database_close(DatabaseHandle handle);

/**
 * @brief 释放数据库句柄
 * 
 * @param handle 数据库句柄
 */
void gograph_database_free(DatabaseHandle handle);

/* ============================================================================
 * 事务管理 API
 * ============================================================================ */

/**
 * @brief 开始一个新事务
 * 
 * @param db_handle 数据库句柄
 * @param read_only 是否只读事务
 * @param error 输出错误信息
 * @return TransactionHandle 事务句柄，失败返回 NULL
 */
TransactionHandle gograph_transaction_begin(
    DatabaseHandle db_handle,
    bool read_only,
    ErrorInfo* error
);

/**
 * @brief 提交事务
 * 
 * @param tx_handle 事务句柄
 * @param error 输出错误信息
 * @return int 错误码
 */
int gograph_transaction_commit(TransactionHandle tx_handle, ErrorInfo* error);

/**
 * @brief 回滚事务
 * 
 * @param tx_handle 事务句柄
 * @return int 错误码
 */
int gograph_transaction_rollback(TransactionHandle tx_handle);

/**
 * @brief 释放事务句柄
 * 
 * @param tx_handle 事务句柄
 */
void gograph_transaction_free(TransactionHandle tx_handle);

/* ============================================================================
 * Cypher 查询 API
 * ============================================================================ */

/**
 * @brief 执行 Cypher 查询
 * 
 * @param tx_handle 事务句柄
 * @param query Cypher 查询语句
 * @param params 查询参数（JSON 格式）
 * @param result 输出查询结果
 * @param error 输出错误信息
 * @return int 错误码
 * 
 * @example
 * ```c
 * QueryResult result;
 * int ret = gograph_cypher_execute(
 *     tx,
 *     "MATCH (n) RETURN n LIMIT 10",
 *     NULL,
 *     &result,
 *     &error
 * );
 * ```
 */
int gograph_cypher_execute(
    TransactionHandle tx_handle,
    const char* query,
    const char* params,
    QueryResult* result,
    ErrorInfo* error
);

/**
 * @brief 创建节点
 * 
 * @param tx_handle 事务句柄
 * @param label 节点标签
 * @param properties 属性（JSON 格式）
 * @param node_id 输出节点 ID
 * @param error 输出错误信息
 * @return int 错误码
 */
int gograph_node_create(
    TransactionHandle tx_handle,
    const char* label,
    const char* properties,
    uint64_t* node_id,
    ErrorInfo* error
);

/**
 * @brief 查询节点
 * 
 * @param tx_handle 事务句柄
 * @param node_id 节点 ID
 * @param node 输出节点结构
 * @param error 输出错误信息
 * @return int 错误码
 */
int gograph_node_get(
    TransactionHandle tx_handle,
    uint64_t node_id,
    Node* node,
    ErrorInfo* error
);

/**
 * @brief 删除节点
 * 
 * @param tx_handle 事务句柄
 * @param node_id 节点 ID
 * @param error 输出错误信息
 * @return int 错误码
 */
int gograph_node_delete(
    TransactionHandle tx_handle,
    uint64_t node_id,
    ErrorInfo* error
);

/**
 * @brief 创建关系
 * 
 * @param tx_handle 事务句柄
 * @param rel_type 关系类型
 * @param start_node_id 起始节点 ID
 * @param end_node_id 结束节点 ID
 * @param properties 属性（JSON 格式）
 * @param rel_id 输出关系 ID
 * @param error 输出错误信息
 * @return int 错误码
 */
int gograph_relationship_create(
    TransactionHandle tx_handle,
    const char* rel_type,
    uint64_t start_node_id,
    uint64_t end_node_id,
    const char* properties,
    uint64_t* rel_id,
    ErrorInfo* error
);

/**
 * @brief 查询关系
 * 
 * @param tx_handle 事务句柄
 * @param rel_id 关系 ID
 * @param rel 输出关系结构
 * @param error 输出错误信息
 * @return int 错误码
 */
int gograph_relationship_get(
    TransactionHandle tx_handle,
    uint64_t rel_id,
    Relationship* rel,
    ErrorInfo* error
);

/* ============================================================================
 * 内存管理 API
 * ============================================================================ */

/**
 * @brief 释放查询结果
 * 
 * @param result 查询结果
 */
void gograph_query_result_free(QueryResult* result);

/**
 * @brief 释放节点
 * 
 * @param node 节点结构
 */
void gograph_node_free(Node* node);

/**
 * @brief 释放关系
 * 
 * @param rel 关系结构
 */
void gograph_relationship_free(Relationship* rel);

/**
 * @brief 释放路径
 * 
 * @param path 路径结构
 */
void gograph_path_free(Path* path);

/**
 * @brief 释放值
 * 
 * @param value 值结构
 */
void gograph_value_free(Value* value);

/**
 * @brief 释放错误信息
 * 
 * @param error 错误信息结构
 */
void gograph_error_free(ErrorInfo* error);

#ifdef __cplusplus
}
#endif

#endif /* GOGRAPH_C_H */
