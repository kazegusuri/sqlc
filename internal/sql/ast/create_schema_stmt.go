package ast

type CreateSchemaStmt struct {
	Name           *string
	SchemaElts     *List
	Authrole       *RoleSpec
	IfNotExists    bool
	SourceLocation *SourceLocation
}

func (n *CreateSchemaStmt) Pos() int {
	return 0
}
