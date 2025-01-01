package ast

type SQLCommentType int

const (
	SQLLeadingComment SQLCommentType = iota
	SQLTrailingComment
)

type SQLCommentGroup struct {
	Comments []*SQLComment
	Type     SQLCommentType
}

type SQLComment struct {
	Text string
}
