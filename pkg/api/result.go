// Package api provides the public database API for gograph, a graph database engine.
// It offers a high-level interface for executing Cypher queries and managing
// graph data including nodes, relationships, and their properties.
//
// Basic Usage:
//
//	// Open a database
//	db, err := api.Open("/path/to/db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Execute a query that modifies data
//	result, err := db.Exec(ctx, "CREATE (n:Person {name: $name})", map[string]interface{}{
//	    "name": "Alice",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created %d nodes\n", result.NodesCreated)
//
//	// Query data
//	rows, err := db.Query(ctx, "MATCH (n:Person) RETURN n.name")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//	    var name string
//	    if err := rows.Scan(&name); err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(name)
//	}
//
// Thread Safety:
//
// DB is safe for concurrent use by multiple goroutines. However, individual
// transactions should not be shared between goroutines.
package api

import "github.com/DotNetAge/gograph/pkg/cypher"

// Result contains the count of affected graph elements after executing
// a data-modifying Cypher query (CREATE, SET, DELETE, REMOVE).
//
// It provides information about how many nodes and relationships were
// created, modified, or deleted during the query execution.
//
// Example:
//
//	result, err := db.Exec(ctx, "CREATE (n:Person) SET n.name = 'Alice'")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Nodes created: %d\n", result.NodesCreated)
//	fmt.Printf("Nodes affected: %d\n", result.NodesAffected)
type Result = cypher.Result
