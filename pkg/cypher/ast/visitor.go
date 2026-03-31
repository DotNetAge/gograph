package ast

// Visitor is the interface for visiting AST nodes.
// Implementations can perform operations on nodes during the walk.
type Visitor interface {
	// Visit is called for each node in the AST.
	// It returns the visitor to use for child nodes, or nil to skip children.
	// An error can be returned to stop the walk.
	Visit(node Node) (w Visitor, err error)
}

// Walker walks an AST and calls the visitor for each node.
type Walker struct {
	visitor Visitor
}

// NewWalker creates a new Walker with the given visitor.
//
// Parameters:
//   - v: The visitor to use during the walk
//
// Returns a new Walker instance.
func NewWalker(v Visitor) *Walker {
	return &Walker{visitor: v}
}

// Walk walks the AST starting from the given node, calling the visitor for each node.
//
// Parameters:
//   - v: The visitor to use
//   - node: The root node to start walking from
//
// Returns an error if the walk is stopped early.
//
// Example:
//
//	visitor := &myVisitor{}
//	err := ast.Walk(visitor, query)
func Walk(v Visitor, node Node) error {
	if node == nil {
		return nil
	}
	walker := &Walker{visitor: v}
	return walker.Walk(node)
}

// Walk walks the AST starting from the given node.
// It implements the visitor pattern for the AST.
//
// Parameters:
//   - node: The node to walk
//
// Returns an error if the walk encounters an error.
func (w *Walker) Walk(node Node) error {
	if node == nil {
		return nil
	}

	v, err := w.visitor.Visit(node)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}

	switch n := node.(type) {
	case *Query:
		for _, stmt := range n.Statements {
			if err := w.Walk(stmt); err != nil {
				return err
			}
		}
	case *MatchStmt:
		for _, clause := range n.Clauses {
			if err := w.Walk(clause); err != nil {
				return err
			}
		}
	case *CreateStmt:
		if n.Pattern != nil {
			if err := w.Walk(n.Pattern); err != nil {
				return err
			}
		}
		for _, clause := range n.Clauses {
			if err := w.Walk(clause); err != nil {
				return err
			}
		}
	case *MergeStmt:
		if n.Pattern != nil {
			if err := w.Walk(n.Pattern); err != nil {
				return err
			}
		}
		if n.Clause != nil {
			if err := w.Walk(n.Clause); err != nil {
				return err
			}
		}
	case *SetStmt:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *DeleteStmt:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *RemoveStmt:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *WithStmt:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
		if n.Where != nil {
			if err := w.Walk(n.Where); err != nil {
				return err
			}
		}
	case *ReturnStmt:
		if n.Return != nil {
			if err := w.Walk(n.Return); err != nil {
				return err
			}
		}
	case *MatchClause:
		if n.Pattern != nil {
			if err := w.Walk(n.Pattern); err != nil {
				return err
			}
		}
		if n.Where != nil {
			if err := w.Walk(n.Where); err != nil {
				return err
			}
		}
		if n.Return != nil {
			if err := w.Walk(n.Return); err != nil {
				return err
			}
		}
		if n.Delete != nil {
			if err := w.Walk(n.Delete); err != nil {
				return err
			}
		}
	case *CreateClause:
		if n.Pattern != nil {
			if err := w.Walk(n.Pattern); err != nil {
				return err
			}
		}
	case *MergeClause:
		if n.Pattern != nil {
			if err := w.Walk(n.Pattern); err != nil {
				return err
			}
		}
	case *SetClause:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *DeleteClause:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *RemoveClause:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *WhereClause:
		if n.Expr != nil {
			if err := w.Walk(n.Expr); err != nil {
				return err
			}
		}
	case *WhereExpr:
		if n.Expr != nil {
			if err := w.Walk(n.Expr); err != nil {
				return err
			}
		}
	case *ReturnClause:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
		if n.OrderBy != nil {
			if err := w.Walk(n.OrderBy); err != nil {
				return err
			}
		}
		if n.Skip != nil {
			if err := w.Walk(n.Skip); err != nil {
				return err
			}
		}
		if n.Limit != nil {
			if err := w.Walk(n.Limit); err != nil {
				return err
			}
		}
	case *ReturnExpr:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
		if n.OrderBy != nil {
			if err := w.Walk(n.OrderBy); err != nil {
				return err
			}
		}
		if n.Skip != nil {
			if err := w.Walk(n.Skip); err != nil {
				return err
			}
		}
		if n.Limit != nil {
			if err := w.Walk(n.Limit); err != nil {
				return err
			}
		}
	case *OrderByExpr:
		for _, item := range n.Items {
			if err := w.Walk(item); err != nil {
				return err
			}
		}
	case *PatternExpr:
		for _, part := range n.Parts {
			if err := w.Walk(part); err != nil {
				return err
			}
		}
	case *PathExpr:
		for _, node := range n.Nodes {
			if err := w.Walk(node); err != nil {
				return err
			}
		}
		for _, rel := range n.Relationships {
			if err := w.Walk(rel); err != nil {
				return err
			}
		}
	case *BinaryExpr:
		if n.Left != nil {
			if err := w.Walk(n.Left); err != nil {
				return err
			}
		}
		if n.Right != nil {
			if err := w.Walk(n.Right); err != nil {
				return err
			}
		}
	case *FuncCall:
		for _, arg := range n.Args {
			if err := w.Walk(arg); err != nil {
				return err
			}
		}
	case *ListLit:
		for _, elem := range n.Elements {
			if err := w.Walk(elem); err != nil {
				return err
			}
		}
	}

	return nil
}
