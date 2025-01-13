package postgresql

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/sqlc-dev/sqlc/internal/sql/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/catalog"
	"github.com/sqlc-dev/sqlc/internal/sql/sqlerr"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func newTestCatalog() *catalog.Catalog {
	return catalog.New("main")
}

func TestUpdate(t *testing.T) {
	p := NewParser()

	for i, tc := range []struct {
		stmt string
		s    *catalog.Schema
	}{
		{
			`
			-- public schema
			CREATE SCHEMA IF NOT EXISTS public;
			`,
			&catalog.Schema{
				Name: "public",
				SourceLocation: &catalog.SourceLocation{
					Filename: "test", StartLine: 2, StartColumn: 3,
					LeadingComments: " public schema",
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text);
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			"CREATE TABLE foo (foo text);\nCREATE TABLE bar (bar text);",
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "foo",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 0, StartColumn: 18,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 0, StartColumn: 0,
						},
					},
					{
						Rel: &ast.TableName{Name: "bar"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 18,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 0,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text); -- foo's comment
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
							TrailingComments: ` foo's comment`,
						},
					},
				},
			},
		},
		{
			`
			-- foo's leading comment
			CREATE TABLE foo (bar text); -- foo's trailing comment
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 2, StartColumn: 21,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 2, StartColumn: 3,
							LeadingComments:  ` foo's leading comment`,
							TrailingComments: ` foo's trailing comment`,
						},
					},
				},
			},
		},
		{
			`
			-- foo's leading detached comment
			-- comment2

			-- foo's leading comment
			-- comment2
			CREATE TABLE foo (bar text); -- foo's trailing comment
			 -- foo's definition ends
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 6, StartColumn: 21,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 6, StartColumn: 3,
							LeadingDetachedComments: []string{" foo's leading detached comment\n comment2"},
							LeadingComments:         " foo's leading comment\n comment2",
							TrailingComments:        " foo's trailing comment\n foo's definition ends",
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text);
			ALTER TABLE foo RENAME TO baz;
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "baz"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text);
			ALTER TABLE foo ADD COLUMN baz bool;
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
							{
								Name:           "baz",
								Type:           ast.TypeName{Name: "bool"},
								SourceLocation: nil,
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text);
			ALTER TABLE foo RENAME COLUMN bar TO baz;
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "baz",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text);
			ALTER TABLE foo RENAME bar TO baz;
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "baz",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text PRIMARY KEY);
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name:      "bar",
								Type:      ast.TypeName{Name: "text"},
								IsNotNull: true,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						Indexes: []*catalog.Index{
							{
								Elems: []*catalog.IndexElem{
									{
										Name:          "bar",
										Ordering:      catalog.SortByDirDefault,
										NullsOrdering: catalog.SortByNullsDefault,
									},
								},
								IsUnique:       false,
								IsPrimary:      true,
								SourceLocation: nil,
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (
				name1 text NOT NULL, -- name1
				name2 text NOT NULL, -- name2
				PRIMARY KEY(name2, name1), -- primary key comment
				UNIQUE(name1) -- unique key comment
			);
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name:      "name1",
								Type:      ast.TypeName{Name: "text"},
								IsNotNull: true,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 2, StartColumn: 4,
									TrailingComments: " name1",
								},
							},
							{
								Name:      "name2",
								Type:      ast.TypeName{Name: "text"},
								IsNotNull: true,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 3, StartColumn: 4,
									TrailingComments: " name2",
								},
							},
						},
						Indexes: []*catalog.Index{
							{
								Elems: []*catalog.IndexElem{
									{
										Name:          "name2",
										Ordering:      catalog.SortByDirDefault,
										NullsOrdering: catalog.SortByNullsDefault,
									},
									{
										Name:          "name1",
										Ordering:      catalog.SortByDirDefault,
										NullsOrdering: catalog.SortByNullsDefault,
									},
								},
								IsUnique:  false,
								IsPrimary: true,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 4, StartColumn: 4,
									TrailingComments: " primary key comment",
								},
							},
							{
								Elems: []*catalog.IndexElem{
									{
										Name:          "name1",
										Ordering:      catalog.SortByDirDefault,
										NullsOrdering: catalog.SortByNullsDefault,
									},
								},
								IsUnique:  true,
								IsPrimary: false,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 5, StartColumn: 4,
									TrailingComments: " unique key comment",
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo ( -- foo's comment
				name1 text NOT NULL, -- name1's comment
				name2 text NOT NULL  -- name2's comment
			); -- ignored comment
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name:      "name1",
								Type:      ast.TypeName{Name: "text"},
								IsNotNull: true,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 2, StartColumn: 4,
									TrailingComments: " name1's comment",
								},
							},
							{
								Name:      "name2",
								Type:      ast.TypeName{Name: "text"},
								IsNotNull: true,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 3, StartColumn: 4,
									TrailingComments: " name2's comment",
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
							TrailingComments: " foo's comment",
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text);
			 -- foo's index leading comment
			CREATE INDEX ON foo (bar); -- foo's index trailing comment
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						Indexes: []*catalog.Index{
							{
								Elems: []*catalog.IndexElem{
									{
										Name:          "bar",
										Ordering:      catalog.SortByDirDefault,
										NullsOrdering: catalog.SortByNullsDefault,
									},
								},
								IsUnique:  false,
								IsPrimary: false,
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 3, StartColumn: 3,
									LeadingComments:  " foo's index leading comment",
									TrailingComments: " foo's index trailing comment",
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TABLE foo (bar text UNIQUE);
			`,
			&catalog.Schema{
				Name: "main",
				Tables: []*catalog.Table{
					{
						Rel: &ast.TableName{Name: "foo"},
						Columns: []*catalog.Column{
							{
								Name: "bar",
								Type: ast.TypeName{Name: "text"},
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 21,
								},
							},
						},
						Indexes: []*catalog.Index{
							{
								Elems: []*catalog.IndexElem{
									{
										Name:          "bar",
										Ordering:      catalog.SortByDirDefault,
										NullsOrdering: catalog.SortByNullsDefault,
									},
								},
								IsUnique:       true,
								IsPrimary:      false,
								SourceLocation: nil,
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TYPE foo AS ENUM ('xx', 'yy');
			`,
			&catalog.Schema{
				Name: "main",
				Types: []catalog.Type{
					&catalog.Enum{
						Name: "foo",
						Vals: []string{"xx", "yy"},
						EnumVals: []*catalog.EnumValue{
							{
								Val: "xx",
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 28,
								},
							},
							{
								Val: "yy",
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 1, StartColumn: 34,
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
						},
					},
				},
			},
		},
		{
			`
			CREATE TYPE foo AS ENUM ( -- foo's comment
				'xx', -- xx's comment
				'yy'  -- yy's comment
			);
			`,
			&catalog.Schema{
				Name: "main",
				Types: []catalog.Type{
					&catalog.Enum{
						Name: "foo",
						Vals: []string{"xx", "yy"},
						EnumVals: []*catalog.EnumValue{
							{
								Val: "xx",
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 2, StartColumn: 4,
									TrailingComments: " xx's comment",
								},
							},
							{
								Val: "yy",
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 3, StartColumn: 4,
									TrailingComments: " yy's comment",
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 1, StartColumn: 3,
							TrailingComments: " foo's comment",
						},
					},
				},
			},
		},
		{
			`
			-- foo's comment
			CREATE TYPE foo AS ENUM (
				-- xx's comment
				'xx',
				-- yy's comment
				'yy'
			);
			`,
			&catalog.Schema{
				Name: "main",
				Types: []catalog.Type{
					&catalog.Enum{
						Name: "foo",
						Vals: []string{"xx", "yy"},
						EnumVals: []*catalog.EnumValue{
							{
								Val: "xx",
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 4, StartColumn: 4,
									LeadingComments: " xx's comment",
								},
							},
							{
								Val: "yy",
								SourceLocation: &catalog.SourceLocation{
									Filename: "test", StartLine: 6, StartColumn: 4,
									LeadingComments: " yy's comment",
								},
							},
						},
						SourceLocation: &catalog.SourceLocation{
							Filename: "test", StartLine: 2, StartColumn: 3,
							LeadingComments: " foo's comment",
						},
					},
				},
			},
		},
	} {
		test := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stmts, err := p.Parse(strings.NewReader(test.stmt), "test")
			if err != nil {
				t.Log(test.stmt)
				t.Fatal(err)
			}

			c := newTestCatalog()
			if err := c.Build(stmts); err != nil {
				t.Log(test.stmt)
				t.Fatal(err)
			}

			e := newTestCatalog()
			if test.s != nil {
				var replaced bool
				for i := range e.Schemas {
					if e.Schemas[i].Name == test.s.Name {
						e.Schemas[i] = test.s
						replaced = true
						break
					}
				}
				if !replaced {
					e.Schemas = append(e.Schemas, test.s)
				}
			}

			if diff := cmp.Diff(e, c, cmpopts.EquateEmpty(), cmpopts.IgnoreUnexported(catalog.Column{})); diff != "" {
				t.Log(test.stmt)
				t.Errorf("catalog mismatch:\n%s", diff)
			}
		})
	}
}

func TestUpdateErrors(t *testing.T) {
	p := NewParser()
	for i, tc := range []struct {
		stmt string
		err  *sqlerr.Error
	}{
		{
			`
			CREATE TABLE foo ();
			CREATE TABLE foo ();
			`,
			sqlerr.RelationExists("foo"),
		},
		{
			`
			CREATE TYPE foo AS ENUM ('bar');
			CREATE TYPE foo AS ENUM ('bar');
			`,
			sqlerr.TypeExists("foo"),
		},
		{
			`
			DROP TABLE foo;
			`,
			sqlerr.RelationNotFound("foo"),
		},
		{
			`
			DROP TYPE foo;
			`,
			sqlerr.TypeNotFound("foo"),
		},
		{
			`
			CREATE TABLE foo ();
			CREATE TABLE bar ();
			ALTER TABLE foo RENAME TO bar;
			`,
			sqlerr.RelationExists("bar"),
		},
		{
			`
			ALTER TABLE foo RENAME TO bar;
			`,
			sqlerr.RelationNotFound("foo"),
		},
		{
			`
			CREATE TABLE foo ();
			ALTER TABLE foo ADD COLUMN bar text;
			ALTER TABLE foo ADD COLUMN bar text;
			`,
			sqlerr.ColumnExists("foo", "bar"),
		},
		{
			`
			CREATE TABLE foo ();
			ALTER TABLE foo DROP COLUMN bar;
			`,
			sqlerr.ColumnNotFound("foo", "bar"),
		},
		{
			`
			CREATE TABLE foo ();
			ALTER TABLE foo ALTER COLUMN bar SET NOT NULL;
			`,
			sqlerr.ColumnNotFound("foo", "bar"),
		},
		{
			`
			CREATE TABLE foo ();
			ALTER TABLE foo ALTER COLUMN bar DROP NOT NULL;
			`,
			sqlerr.ColumnNotFound("foo", "bar"),
		},
		{
			`
			CREATE SCHEMA foo;
			CREATE SCHEMA foo;
			`,
			sqlerr.SchemaExists("foo"),
		},
		{
			`
			ALTER TABLE foo.baz SET SCHEMA bar;
			`,
			sqlerr.SchemaNotFound("foo"),
		},
		{
			`
			CREATE SCHEMA foo;
			ALTER TABLE foo.baz SET SCHEMA bar;
			`,
			sqlerr.RelationNotFound("baz"),
		},
		{
			`
			CREATE SCHEMA foo;
			CREATE TABLE foo.baz ();
			ALTER TABLE foo.baz SET SCHEMA bar;
			`,
			sqlerr.SchemaNotFound("bar"),
		},
		{
			`
			DROP SCHEMA bar;
			`,
			sqlerr.SchemaNotFound("bar"),
		},
		{
			`
			ALTER TABLE foo RENAME bar TO baz;
			`,
			sqlerr.RelationNotFound("foo"),
		},
		{
			`
			CREATE TABLE foo ();
			ALTER TABLE foo RENAME bar TO baz;
			`,
			sqlerr.ColumnNotFound("foo", "bar"),
		},
		{
			`
			CREATE TABLE foo (bar text, baz text);
			ALTER TABLE foo RENAME bar TO baz;
			`,
			sqlerr.ColumnExists("foo", "baz"),
		},
	} {
		test := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stmts, err := p.Parse(strings.NewReader(test.stmt), "test")
			if err != nil {
				t.Log(test.stmt)
				t.Fatal(err)
			}

			c := NewCatalog()
			err = c.Build(stmts)
			if err == nil {
				t.Log(test.stmt)
				t.Fatal("err was nil")
			}

			var actual *sqlerr.Error
			if !errors.As(err, &actual) {
				t.Fatalf("err is not *sqlerr.Error: %#v", err)
			}

			if diff := cmp.Diff(test.err.Error(), actual.Error()); diff != "" {
				t.Log(test.stmt)
				t.Errorf("error mismatch: \n%s", diff)
			}
		})
	}
}
