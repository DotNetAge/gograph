# GoGraph 技术文档

欢迎使用 **GoGraph**，这是一款纯 Go 编写的嵌入式图数据库（Embedded Graph Database），采用类似于 SQLite 的单文件零部署设计，支持核心 Cypher 查询子集。

本文档包含完整的系统架构、接口使用说明和开发者最佳实践，基于实际代码实现编写。

## 目录

1. [快速开始 (Quick Start)](quick-start.md) - 环境配置与基础使用示例，助你在30分钟内跑通核心功能。
2. [核心接口说明 (Core Interfaces)](core-interfaces.md) - `api.DB`, `api.Tx`, `api.Rows` 等公共接口详情。
3. [系统架构 (Architecture)](architecture.md) - 模块划分、数据流向及底层存储设计图解。
4. [开发者指南 (Developer Guide)](developer-guide.md) - 最佳实践、可观测性注入、性能优化建议及常见问题。
5. [API参考 (API Reference)](api-reference.md) - 涵盖 `api`, `cypher`, `graph`, `storage` 核心公共方法列表。
6. [命令行工具指南 (CLI & TUI Guide)](cli-guide.md) - 提供如何编译、运行交互式 Cypher 终端及基础命令的详尽参考。

---

**核心定位**
- **轻量嵌入式**：无额外进程依赖，直接通过 `go get` 引入。
- **高可靠性**：基于 [Pebble DB](https://github.com/cockroachdb/pebble) 的单文件键值对持久化，提供 WAL 崩溃恢复支持与 MVCC 读写事务。
- **Cypher 支持**：原生支持核心的 `MATCH`、`CREATE`、`SET`、`DELETE`、`REMOVE` 语句，底层查询优化器自动集成 Index Scan 与 $O(1)$ 邻接表图遍历。
