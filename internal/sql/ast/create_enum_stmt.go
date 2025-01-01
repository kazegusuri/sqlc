package ast

type CreateEnumStmt struct {
	TypeName           *TypeName
	Vals               *List
	AssociatedComments []*SQLCommentGroup
}

func (n *CreateEnumStmt) Pos() int {
	return 0
}
