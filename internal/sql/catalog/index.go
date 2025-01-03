package catalog

import (
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

type Index struct {
	Name             string
	Elems            []*IndexElem
	IsUnique         bool
	IsPrimary        bool
	AttachedComments []*CommentGroup
}

func (c *Catalog) createIndex(stmt *ast.IndexStmt) error {
	var name string
	if stmt.Idxname != nil {
		name = *stmt.Idxname
	}

	ns := c.DefaultSchema
	if stmt.Relation.Schemaname != nil && *stmt.Relation.Schemaname != "" {
		ns = *stmt.Relation.Schemaname
	}

	var relation string
	if stmt.Relation.Relname != nil {
		relation = *stmt.Relation.Relname
	}

	tblname := &ast.TableName{
		Catalog: "",
		Schema:  ns,
		Name:    relation,
	}

	_, tbl, err := c.getTable(tblname)
	if err != nil {
		return err
	}

	idx := &Index{
		Name:             name,
		IsUnique:         stmt.Unique,
		IsPrimary:        stmt.Primary,
		AttachedComments: convertAttachedComments(stmt.AttachedComments),
	}

	if stmt.IndexParams != nil {
		var elems []*IndexElem
		for _, param := range stmt.IndexParams.Items {
			elem, ok := param.(*ast.IndexElem)
			if !ok {
				continue
			}

			var name string
			if elem.Name != nil {
				name = *elem.Name
			}
			elems = append(elems, &IndexElem{
				Name:          name,
				Ordering:      SortByDir(elem.Ordering),
				NullsOrdering: SortByNulls(elem.NullsOrdering),
			})
		}
		idx.Elems = elems
	}

	tbl.Indexes = append(tbl.Indexes, idx)

	return nil
}
