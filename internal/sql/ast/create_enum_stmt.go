package ast

type CreateEnumStmt struct {
	TypeName *TypeName
	Vals     *List

	SourceLocation *SourceLocation
}

func (n *CreateEnumStmt) Pos() int {
	return 0
}
