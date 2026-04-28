"""
Transaction module - Provides transaction management for GoGraph
"""

import json
from typing import Dict, Optional, Any
from ._gograph import (
    transaction_begin,
    transaction_commit,
    transaction_rollback,
    transaction_free,
    node_create,
    node_get,
    node_delete,
    relationship_create,
    relationship_get,
    cypher_execute,
    GOGRAPH_OK,
    GOGRAPH_ERROR_NOT_FOUND,
    GOGRAPH_ERROR_EXEC_FAILURE,
    new_ErrorInfo,
    delete_ErrorInfo,
    ErrorInfo_code_get,
    ErrorInfo_message_get,
    new_Node,
    delete_Node,
    new_Relationship,
    delete_Relationship,
)
from .database import GoGraphError


class Transaction:
    """
    Transaction interface for GoGraph.
    
    Example:
        >>> with db.transaction() as tx:
        ...     node_id = tx.create_node("Person", {"name": "Alice"})
        ...     tx.commit()
    """
    
    def __init__(self, database, read_only: bool = False):
        """
        Initialize a transaction.
        
        Args:
            database: The database instance
            read_only: If True, create a read-only transaction
        """
        self._database = database
        self._handle = None
        self._read_only = read_only
        self._committed = False
        self._rollback = False
        self._begin()
    
    def _begin(self):
        """Internal method to begin the transaction"""
        error = new_ErrorInfo()
        self._handle = transaction_begin(
            self._database.handle,
            self._read_only,
            error
        )
        
        if self._handle is None:
            msg = f"Failed to begin transaction: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to begin transaction"
            delete_ErrorInfo(error)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
    
    def create_node(self, label: str, properties: Optional[Dict[str, Any]] = None) -> int:
        """
        Create a new node.
        
        Args:
            label: The node label
            properties: Dictionary of properties
        
        Returns:
            The node ID
        """
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        props_json = json.dumps(properties or {}).encode('utf-8')
        node_id = [0]
        error = new_ErrorInfo()
        
        result = node_create(
            self._handle,
            label.encode('utf-8'),
            props_json,
            node_id,
            error
        )
        
        if result != GOGRAPH_OK:
            msg = f"Failed to create node: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to create node"
            delete_ErrorInfo(error)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
        return node_id[0]
    
    def get_node(self, node_id: int):
        """
        Get a node by ID.
        
        Args:
            node_id: The node ID
        
        Returns:
            Node object
        """
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        node = new_Node()
        error = new_ErrorInfo()
        
        result = node_get(self._handle, node_id, node, error)
        
        if result != GOGRAPH_OK:
            msg = f"Failed to get node: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to get node"
            delete_ErrorInfo(error)
            delete_Node(node)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
        return node
    
    def delete_node(self, node_id: int):
        """
        Delete a node by ID.
        
        Args:
            node_id: The node ID
        """
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        error = new_ErrorInfo()
        
        result = node_delete(self._handle, node_id, error)
        
        if result != GOGRAPH_OK:
            msg = f"Failed to delete node: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to delete node"
            delete_ErrorInfo(error)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
    
    def create_relationship(self, rel_type: str, start_node_id: int, end_node_id: int, 
                          properties: Optional[Dict[str, Any]] = None) -> int:
        """
        Create a new relationship.
        
        Args:
            rel_type: The relationship type
            start_node_id: The start node ID
            end_node_id: The end node ID
            properties: Dictionary of properties
        
        Returns:
            The relationship ID
        """
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        props_json = json.dumps(properties or {}).encode('utf-8')
        rel_id = [0]
        error = new_ErrorInfo()
        
        result = relationship_create(
            self._handle,
            rel_type.encode('utf-8'),
            start_node_id,
            end_node_id,
            props_json,
            rel_id,
            error
        )
        
        if result != GOGRAPH_OK:
            msg = f"Failed to create relationship: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to create relationship"
            delete_ErrorInfo(error)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
        return rel_id[0]
    
    def get_relationship(self, rel_id: int):
        """
        Get a relationship by ID.
        
        Args:
            rel_id: The relationship ID
        
        Returns:
            Relationship object
        """
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        rel = new_Relationship()
        error = new_ErrorInfo()
        
        result = relationship_get(self._handle, rel_id, rel, error)
        
        if result != GOGRAPH_OK:
            msg = f"Failed to get relationship: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to get relationship"
            delete_ErrorInfo(error)
            delete_Relationship(rel)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
        return rel
    
    def execute(self, query: str, params: Optional[Dict[str, Any]] = None):
        """
        Execute a Cypher query.
        
        Args:
            query: The Cypher query string
            params: Dictionary of query parameters
        
        Returns:
            QueryResult
        """
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        pass
    
    def commit(self):
        """Commit the transaction"""
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        error = new_ErrorInfo()
        result = transaction_commit(self._handle, error)
        
        if result != GOGRAPH_OK:
            msg = f"Failed to commit transaction: {ErrorInfo_message_get(error).decode('utf-8')}" if ErrorInfo_message_get(error) else "Failed to commit transaction"
            delete_ErrorInfo(error)
            raise GoGraphError(msg)
        
        delete_ErrorInfo(error)
        self._committed = True
    
    def rollback(self):
        """Rollback the transaction"""
        if self._committed or self._rollback:
            raise GoGraphError("Transaction has already been committed or rolled back")
        
        transaction_rollback(self._handle)
        self._rollback = True
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        if not self._committed and not self._rollback:
            if exc_type is not None:
                self.rollback()
            else:
                self.commit()
    
    def __del__(self):
        if self._handle is not None and not self._committed and not self._rollback:
            transaction_rollback(self._handle)
            transaction_free(self._handle)
    
    @property
    def read_only(self) -> bool:
        """Check if this is a read-only transaction"""
        return self._read_only
    
    @property
    def committed(self) -> bool:
        """Check if the transaction has been committed"""
        return self._committed
    
    @property
    def rolled_back(self) -> bool:
        """Check if the transaction has been rolled back"""
        return self._rollback
