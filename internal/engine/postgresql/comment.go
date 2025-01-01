package postgresql

import (
	"bufio"
	"bytes"
	"slices"
	"sort"
	"strings"

	nodes "github.com/pganalyze/pg_query_go/v5"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

type CommentsDispatcher struct {
	linePos  []int64
	curGroup int
	groups   []*CommentGroup
	contents string
}

func (d *CommentsDispatcher) FindLocationInStatement(stmt *nodes.RawStmt, str string) int32 {
	stmtContents := d.contents[stmt.StmtLocation : stmt.StmtLocation+stmt.StmtLen]
	index := strings.Index(stmtContents, str)
	if index >= 0 {
		return stmt.StmtLocation + int32(index)
	}
	return -1
}

func (d *CommentsDispatcher) makeASTCommentGroup(g *CommentGroup, s ast.SQLCommentType) *ast.SQLCommentGroup {
	comments := make([]*ast.SQLComment, len(g.Comments))
	for i := range g.Comments {
		comments[i] = &ast.SQLComment{
			Text: g.Comments[i].Text,
		}
	}

	return &ast.SQLCommentGroup{
		Comments: comments,
		Type:     s,
	}
}

func (d *CommentsDispatcher) AssociatedCommentGroups(offset int32) []*ast.SQLCommentGroup {

	var trailingComment *ast.SQLCommentGroup
	if group := d.lineCommentGroups(offset); group != nil {
		trailingComment = d.makeASTCommentGroup(group, ast.SQLTrailingComment)
	}

	var comments []*ast.SQLCommentGroup
	for _, group := range d.leadingCommentGroups(offset) {
		comments = append(comments, d.makeASTCommentGroup(group, ast.SQLLeadingComment))
	}

	if trailingComment != nil {
		comments = append(comments, trailingComment)
	}

	return comments
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

func (d *CommentsDispatcher) getLineNo(offset int32) int32 {
	index := sort.Search(len(d.linePos), func(i int) bool {
		return d.linePos[i] >= int64(offset)
	})
	return int32(index + 1)
}

func (d *CommentsDispatcher) lineCommentGroups(offset int32) *CommentGroup {
	lineNo := d.getLineNo(offset)
	pos := slices.IndexFunc(d.groups[d.curGroup:], func(g *CommentGroup) bool {
		return lineNo == d.getLineNo(g.Start())
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

func getLinePositions(contents string) ([]int64, error) {
	r := strings.NewReader(contents)
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLines)

	var lines []int64
	n := int64(0)
	for scanner.Scan() {
		b := scanner.Bytes()
		eol := n
		n += int64(len(b))
		eol += int64(len(b))
		lines = append(lines, eol)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func parseComments(contents string) (*CommentsDispatcher, error) {
	linePos, err := getLinePositions(contents)
	if err != nil {
		return nil, err
	}

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
			c := &Comment{
				Start: token.Start,
				End:   token.End,
				Text:  string(contents[token.Start:token.End]),
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
		linePos:  linePos,
		groups:   commentGroups,
		contents: contents,
	}, nil
}
