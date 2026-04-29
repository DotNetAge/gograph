"""
CypherDB - A high-performance graph database
=============================================

CypherDB is a graph database built with Go, providing Python bindings
for seamless integration with Python applications.

Quick Start:
------------
>>> import cypherdb
>>> db = cypherdb.Database("test.db")
>>> with db.transaction() as tx:
...     node_id = tx.create_node("Person", {"name": "Alice"})
...     print(f"Created node: {node_id}")
"""

from ._gograph import *
from .database import Database
from .transaction import Transaction

__version__ = "0.2.3"
__author__ = "Ray"
__email__ = "ray@rayainfo.cn"

__all__ = [
    # Core types
    'Database',
    'Transaction',
    
    # Constants
    'GOGRAPH_OK',
    'GOGRAPH_ERROR_GENERAL',
    'GOGRAPH_ERROR_INVALID_PARAM',
    'GOGRAPH_ERROR_NOT_FOUND',
    'GOGRAPH_ERROR_ALREADY_EXISTS',
    'GOGRAPH_ERROR_PARSE_FAILURE',
    'GOGRAPH_ERROR_EXEC_FAILURE',
    'GOGRAPH_ERROR_STORAGE_FAILURE',
    'GOGRAPH_ERROR_MEMORY_ALLOC',
    'GOGRAPH_ERROR_TX_CONFLICT',
    
    # Value types
    'VALUE_NULL',
    'VALUE_BOOL',
    'VALUE_INT',
    'VALUE_FLOAT',
    'VALUE_STRING',
    'VALUE_LIST',
    'VALUE_MAP',
]
