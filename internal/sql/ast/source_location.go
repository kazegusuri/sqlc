package ast

type SourceLocation struct {
	Filename    string
	StartLine   int
	StartColumn int

	LeadingDetachedComments []string
	LeadingComments         string
	TrailingComments        string
}
