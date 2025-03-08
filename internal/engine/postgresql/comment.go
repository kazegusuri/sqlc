package postgresql

import (
	"bytes"
	"slices"
	"strings"

	nodes "github.com/pganalyze/pg_query_go/v6"
	"github.com/sqlc-dev/sqlc/internal/source"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

type CommentsDispatcher struct {
	curGroup  int
	groups    []*CommentGroup
	contents  string
	locations *SourceLocations
}

func (d *CommentsDispatcher) Contents(loc, len int) string {
	return d.contents[loc : loc+len]
}

func (d *CommentsDispatcher) FindLocationInStatement(stmt *nodes.RawStmt, str string) int32 {
	stmtContents := d.contents[stmt.StmtLocation : stmt.StmtLocation+stmt.StmtLen]
	index := strings.Index(stmtContents, str)
	if index >= 0 {
		return stmt.StmtLocation + int32(index)
	}
	return -1
}

func (d *CommentsDispatcher) SourceLocationWithComments(offset int32) *ast.SourceLocation {
	loc := d.locations.GetSourceLocation(offset)

	var trailingComment string
	if group := d.lineCommentGroups(offset); group != nil {
		var comments []string
		for _, c := range group.Comments {
			comments = append(comments, c.Text)
		}
		trailingComment = strings.Join(comments, "\n")
	}

	var leadingComments []string
	for _, group := range d.leadingCommentGroups(offset) {
		var comments []string
		for _, c := range group.Comments {
			comments = append(comments, c.Text)
		}
		leadingComments = append(leadingComments, strings.Join(comments, "\n"))
	}

	if len(leadingComments) != 0 {
		loc.LeadingComments = leadingComments[len(leadingComments)-1]
		loc.LeadingDetachedComments = leadingComments[0 : len(leadingComments)-1]
	}
	loc.TrailingComments = trailingComment

	return loc
}

func (d *CommentsDispatcher) leadingCommentGroups(offset int32) []*CommentGroup {
	var count int
	for _, group := range d.groups[d.curGroup:] {
		if offset >= group.End() {
			count++
		}
	}

	groups := d.groups[d.curGroup : d.curGroup+count]
	d.curGroup += count

	return groups
}

func (d *CommentsDispatcher) lineCommentGroups(offset int32) *CommentGroup {
	lineNo, _ := d.locations.GetPosition(offset)
	pos := slices.IndexFunc(d.groups[d.curGroup:], func(g *CommentGroup) bool {
		lineNo2, _ := d.locations.GetPosition(g.Start())
		return lineNo == lineNo2
	})
	if pos < 0 {
		return nil
	}

	group := d.groups[d.curGroup+pos]
	d.groups = slices.Delete(d.groups, d.curGroup+pos, d.curGroup+pos+1)

	return group
}

type Comment struct {
	Start int32
	End   int32
	Text  string
}

type CommentGroup struct {
	Comments []*Comment
}

func (g *CommentGroup) Start() int32 {
	if len(g.Comments) == 0 {
		return 0
	}

	return g.Comments[0].Start
}

func (g *CommentGroup) End() int32 {
	n := len(g.Comments)
	if n == 0 {
		return 0
	}

	return g.Comments[n-1].End
}

func scanLines(data []byte, eof bool) (int, []byte, error) {
	if eof && len(data) == 0 {
		return 0, nil, nil
	}
	if p := bytes.IndexByte(data, '\n'); p >= 0 {
		return p + 1, data[0 : p+1], nil
	}
	if eof {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func newCommentDispatcher(contents string, locations *SourceLocations) (*CommentsDispatcher, error) {
	result, err := ParseScan(contents)
	if err != nil {
		pErr := normalizeErr(err)
		return nil, pErr
	}

	var commentGroups []*CommentGroup
	var comments []*Comment
	for _, token := range result.Tokens {
		switch token.Token {
		case nodes.Token_SQL_COMMENT, nodes.Token_C_COMMENT:
			syntax := source.CommentSyntax{
				Dash:      true,
				Hash:      true,
				SlashStar: true,
			}
			texts, err := source.CleanedComments(string(contents[token.Start:token.End]), syntax)
			if err != nil {
				return nil, err
			}
			c := &Comment{
				Start: token.Start,
				End:   token.End,
				Text:  strings.Join(texts, ""),
			}

			if len(comments) != 0 {
				lastComment := comments[len(comments)-1]
				// break the comment group if there is an empty line between comments.
				if strings.Contains(string(contents[lastComment.End:c.Start]), "\n\n") {
					commentGroups = append(commentGroups, &CommentGroup{
						Comments: comments,
					})
					comments = nil
				}
			}
			comments = append(comments, c)
		default:
			if len(comments) != 0 {
				commentGroups = append(commentGroups, &CommentGroup{
					Comments: comments,
				})
				comments = nil
			}
		}
	}

	if len(comments) != 0 {
		commentGroups = append(commentGroups, &CommentGroup{
			Comments: comments,
		})
	}

	return &CommentsDispatcher{
		groups:    commentGroups,
		contents:  contents,
		locations: locations,
	}, nil
}
