package catalog

type IndexElem struct {
	Name          string
	Ordering      SortByDir
	NullsOrdering SortByNulls
}

type SortByDir uint

const (
	SortByDirUndefined SortByDir = 0
	SortByDirDefault   SortByDir = 1
	SortByDirAsc       SortByDir = 2
	SortByDirDesc      SortByDir = 3
	SortByDirUsing     SortByDir = 4
)

type SortByNulls uint

const (
	SortByNullsUndefined SortByNulls = 0
	SortByNullsDefault   SortByNulls = 1
	SortByNullsFirst     SortByNulls = 2
	SortByNullsLast      SortByNulls = 3
)
