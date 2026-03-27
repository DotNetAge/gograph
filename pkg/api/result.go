// Package api provides the public database API for gograph, a graph database engine.
package api

import "github.com/DotNetAge/gograph/pkg/cypher"

// Result contains the count of affected graph elements after executing
// a data-modifying Cypher query (CREATE, SET, DELETE, REMOVE).
type Result = cypher.Result
