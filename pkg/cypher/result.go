package cypher

// Result contains the count of affected graph elements after executing
// a data-modifying Cypher query (CREATE, SET, DELETE, REMOVE).
type Result struct {
	// AffectedNodes is the number of nodes created, modified, or deleted.
	AffectedNodes int
	// AffectedRels is the number of relationships created or deleted.
	AffectedRels int
	// Rows contains matched data rows for MATCH queries.
	Rows []map[string]interface{}
	// Columns contains the names of columns returned in Rows.
	Columns []string
}

// AddAffected increments the count of affected nodes and relationships.
func (r *Result) AddAffected(nodes, rels int) {
	r.AffectedNodes += nodes
	r.AffectedRels += rels
}
