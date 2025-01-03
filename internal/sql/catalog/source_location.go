package catalog

import "github.com/sqlc-dev/sqlc/internal/sql/ast"

type SourceLocation struct {
	Filename    string
	StartLine   int
	StartColumn int

	LeadingDetachedComments []string
	LeadingComments         string
	TrailingComments        string
}

func convertSourceLocation(loc *ast.SourceLocation) *SourceLocation {
	if loc == nil {
		return nil
	}
	return &SourceLocation{
		Filename:                loc.Filename,
		StartLine:               loc.StartLine,
		StartColumn:             loc.StartColumn,
		LeadingDetachedComments: loc.LeadingDetachedComments,
		LeadingComments:         loc.LeadingComments,
		TrailingComments:        loc.TrailingComments,
	}
}
