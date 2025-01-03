package ast

type CreateEnumStmt struct {
	TypeName         *TypeName
	Vals             *List
	AttachedComments []*SQLCommentGroup
}

func (n *CreateEnumStmt) Pos() int {
	return 0
}
