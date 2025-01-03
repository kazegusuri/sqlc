package ast

type String struct {
	Str string

	// Comments attached with this string.
	AttachedComments []*SQLCommentGroup
}

func (n *String) Pos() int {
	return 0
}

func (n *String) Format(buf *TrackedBuffer) {
	if n == nil {
		return
	}
	buf.WriteString(n.Str)
}
