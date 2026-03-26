package api

import "errors"

var ErrNoMoreRows = errors.New("no more rows")

type Rows struct {
	result  []map[string]interface{}
	columns []string
	index   int
}

func (r *Rows) Next() bool {
	r.index++
	return r.index < len(r.result)
}

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
				if s, ok := val.(int); ok {
					*d = s
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
			}
		}
	}
	return nil
}

func (r *Rows) Close() error {
	return nil
}

func (r *Rows) Columns() []string {
	return r.columns
}
