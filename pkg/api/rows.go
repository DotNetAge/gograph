// Package api provides the public database API for gograph, a graph database engine.
package api

import "errors"

// ErrNoMoreRows is returned by Scan when there are no more rows to read.
var ErrNoMoreRows = errors.New("no more rows")

// Rows represents a result set iterator from a Cypher query.
// It provides methods to iterate over the rows and scan column values
// into Go variables.
//
// Rows is returned by Query and Tx.Query methods. It must be closed
// when no longer needed to release resources.
//
// Example:
//
//	rows, err := db.Query(ctx, "MATCH (n:Person) RETURN n.name, n.age")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//	    var name string
//	    var age int
//	    if err := rows.Scan(&name, &age); err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Name: %s, Age: %d\n", name, age)
//	}
type Rows struct {
	result  []map[string]interface{}
	columns []string
	index   int
}

// Next advances the iterator to the next row.
// It returns true if there is a next row, or false if there are no more rows.
//
// Example:
//
//	for rows.Next() {
//	    // Process row
//	}
func (r *Rows) Next() bool {
	r.index++
	return r.index < len(r.result)
}

// Scan copies the column values of the current row into the dest variables.
// The number of dest arguments must be less than or equal to the number of
// columns in the result set.
//
// Supported destination types:
//   - *string: For string values
//   - *int: For integer values (int or int64)
//   - *int64: For integer values
//   - *float64: For floating point values
//   - *bool: For boolean values
//   - *interface{}: For any value type
//
// Parameters:
//   - dest: Pointers to variables to receive the column values
//
// Returns ErrNoMoreRows if there are no more rows to read.
//
// Example:
//
//	for rows.Next() {
//	    var name string
//	    var age int
//	    if err := rows.Scan(&name, &age); err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Name: %s, Age: %d\n", name, age)
//	}
func (r *Rows) Scan(dest ...interface{}) error {
	if r.index < 0 || r.index >= len(r.result) {
		return ErrNoMoreRows
	}

	row := r.result[r.index]
	for i, col := range r.columns {
		if i >= len(dest) {
			break
		}
		if val, ok := row[col]; ok {
			switch d := dest[i].(type) {
			case *string:
				if s, ok := val.(string); ok {
					*d = s
				}
			case *int:
				switch v := val.(type) {
				case int:
					*d = v
				case int64:
					*d = int(v)
				}

			case *int64:
				switch v := val.(type) {
				case int64:
					*d = v
				case int:
					*d = int64(v)
				}
			case *float64:
				if f, ok := val.(float64); ok {
					*d = f
				}
			case *bool:
				if b, ok := val.(bool); ok {
					*d = b
				}
			case *interface{}:
				*d = val
			}
		}
	}
	return nil
}

// Close releases the Rows resources. It is idempotent and safe to call
// multiple times.
//
// Example:
//
//	defer rows.Close()
func (r *Rows) Close() error {
	return nil
}

// Columns returns the column names of the result set.
//
// Returns a slice of column names in the order they appear in the query.
//
// Example:
//
//	columns := rows.Columns()
//	for _, col := range columns {
//	    fmt.Println(col)
//	}
func (r *Rows) Columns() []string {
	return r.columns
}
