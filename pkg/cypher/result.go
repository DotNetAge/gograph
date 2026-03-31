// Package cypher provides Cypher query parsing and execution for gograph.
// It implements a Cypher-compatible query language for graph database operations.
//
// The package supports:
//   - CREATE: Create nodes and relationships
//   - MATCH: Query and pattern matching
//   - SET: Update properties
//   - DELETE: Remove nodes and relationships
//   - REMOVE: Remove labels and properties
//   - RETURN: Return query results
//
// Basic Usage:
//
//	executor := cypher.NewExecutor(store)
//	result, err := executor.Execute(ctx, "CREATE (n:Person {name: 'Alice'})", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created %d nodes\n", result.AffectedNodes)
package cypher

// Result contains the count of affected graph elements after executing
// a data-modifying Cypher query (CREATE, SET, DELETE, REMOVE).
//
// It also contains the result data for MATCH queries, including the matched
// rows and column names.
//
// Example:
//
//	result, err := executor.Execute(ctx, "CREATE (n:Person)", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Nodes created: %d\n", result.AffectedNodes)
//	fmt.Printf("Relationships created: %d\n", result.AffectedRels)
//
//	// For MATCH queries
//	result, err := executor.Execute(ctx, "MATCH (n:Person) RETURN n.name", nil)
//	for _, row := range result.Rows {
//	    fmt.Println(row["n.name"])
//	}
type Result struct {
	// AffectedNodes is the number of nodes created, modified, or deleted.
	AffectedNodes int

	// AffectedRels is the number of relationships created or deleted.
	AffectedRels int

	// NodesCreated is the number of nodes created during the query.
	NodesCreated int

	// RelsCreated is the number of relationships created during the query.
	RelsCreated int

	// Rows contains matched data rows for MATCH queries.
	// Each row is a map of column names to values.
	Rows []map[string]interface{}

	// Columns contains the names of columns returned in Rows.
	Columns []string
}

// AddAffected increments the count of affected nodes and relationships.
//
// Parameters:
//   - nodes: The number of nodes to add to the affected count
//   - rels: The number of relationships to add to the affected count
//
// Example:
//
//	result := &cypher.Result{}
//	result.AddAffected(2, 1) // 2 nodes and 1 relationship affected
//	fmt.Printf("Total affected: %d nodes, %d relationships\n",
//	    result.AffectedNodes, result.AffectedRels)
func (r *Result) AddAffected(nodes, rels int) {
	r.AffectedNodes += nodes
	r.AffectedRels += rels
}
