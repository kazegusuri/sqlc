package postgresql

import (
	"bufio"
	"sort"
	"strings"

	"github.com/sqlc-dev/sqlc/internal/sql/ast"
)

type SourceLocations struct {
	filename string
	contents string
	lines    []int
}

func (l *SourceLocations) GetSourceLocation(offset int32) *ast.SourceLocation {
	line, col := l.GetPosition(offset)
	return &ast.SourceLocation{
		Filename:    l.filename,
		StartLine:   line,
		StartColumn: col,
	}
}

// get line number and column number from offset
func (l *SourceLocations) GetPosition(offset int32) (int, int) {
	index := sort.Search(len(l.lines), func(i int) bool {
		return l.lines[i] > int(offset)
	})

	// index should be greater than 0
	lineno := index - 1
	col := int(offset) - l.lines[lineno]
	return lineno, col
}

func newSourceLocations(filename, contents string) (*SourceLocations, error) {
	lines, err := getLinePositions(contents)
	if err != nil {
		return nil, err
	}

	return &SourceLocations{
		filename: filename,
		contents: contents,
		lines:    lines,
	}, nil
}

func getLinePositions(contents string) ([]int, error) {
	r := strings.NewReader(contents)
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLines)

	lines := []int{0}
	n := int(0)
	for scanner.Scan() {
		b := scanner.Bytes()
		eol := n
		n += len(b)
		eol += len(b)
		lines = append(lines, eol)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
