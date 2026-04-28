"""
Database module - Provides a Pythonic interface to GoGraph database
"""

import os
from ._gograph import (
    database_new,
    database_open,
    database_close,
    database_free,
    GOGRAPH_OK,
    GOGRAPH_ERROR_STORAGE_FAILURE,
    new_ErrorInfo,
    delete_ErrorInfo,
    ErrorInfo_code_get,
    ErrorInfo_message_get,
)


class GoGraphError(Exception):
    """Base exception for GoGraph errors"""
    pass


class Database:
    """
    GoGraph Database interface.
    
    Example:
        >>> db = Database("my_graph.db")
        >>> with db.transaction() as tx:
        ...     tx.create_node("Person", {"name": "Alice"})
    """
    
    def __init__(self, path: str, create: bool = True):
        """
        Initialize a database.
        
        Args:
            path: Path to the database file
            create: If True, create the database if it doesn't exist
        """
        self._handle = None
        self._path = path
        self._open(create)
    
    def _open(self, create: bool):
        """Internal method to open the database"""
        error = new_ErrorInfo()
        
        if create:
            self._handle = database_new(self._path.encode('utf-8'), error)
        else:
            self._handle = database_open(self._path.encode('utf-8'), error)
        
        if self._handle is None:
            msg = f"Failed to open database: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to open database"
            delete_ErrorInfo(error)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
    
    def transaction(self, read_only: bool = False):
        """
        Create a new transaction context.
        
        Args:
            read_only: If True, create a read-only transaction
        
        Returns:
            Transaction context manager
        """
        from .transaction import Transaction
        return Transaction(self, read_only)
    
    def close(self):
        """Close the database"""
        if self._handle is not None:
            database_close(self._handle)
            database_free(self._handle)
            self._handle = None
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
    
    def __del__(self):
        self.close()
    
    @property
    def path(self) -> str:
        """Get the database path"""
        return self._path
    
    @property
    def handle(self):
        """Get the underlying database handle"""
        return self._handle
