//go:build !windows && cgo

package postgresql

import (
	nodes "github.com/pganalyze/pg_query_go/v6"
	"github.com/wasilibs/go-pgquery/parser"
)

var ParseScan = nodes.Scan
var Parse = nodes.Parse
var DeparseFromProtobuf = parser.DeparseFromProtobuf
var Fingerprint = nodes.Fingerprint
