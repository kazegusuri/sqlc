package catalog

import "github.com/sqlc-dev/sqlc/internal/sql/ast"

type CommentType int

const (
	LeadingComment CommentType = iota
	TrailingComment
)

type CommentGroup struct {
	Comments []string
	Type     CommentType
}

func convertAssociatedComments(gs []*ast.SQLCommentGroup) []*CommentGroup {
	var groups []*CommentGroup
	for _, g := range gs {
		var comments []string
		for _, c := range g.Comments {
			comments = append(comments, c.Text)
		}

		groups = append(groups, &CommentGroup{
			Comments: comments,
			Type:     CommentType(g.Type),
		})
	}
	return groups
}
