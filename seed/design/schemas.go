package design

import "github.com/podhmo/gos/openapigen"

var Builder = openapigen.NewBuilder(openapigen.DefaultConfig()) // for export
var b = Builder

var (
	Definition = b.Object(
		b.Field("alias", b.String().Example(":dizzy:")),
		b.Field("char", b.String().Example("💫")),
	)
)
