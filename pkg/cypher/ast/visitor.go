package ast

type Visitor interface {
	Visit(node Node) (w Visitor, err error)
}

type Walker struct {
	visitor Visitor
}

func NewWalker(v Visitor) *Walker {
	return &Walker{visitor: v}
}

func Walk(v Visitor, node Node) error {
	if node == nil {
		return nil
	}
	walker := &Walker{visitor: v}
	return walker.Walk(node)
}

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
