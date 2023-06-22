package main

import (
	"encoding/json"
	"os"

	"github.com/iancoleman/orderedmap"
	"github.com/podhmo/emoji-api/seed/design"
	"github.com/podhmo/gos/openapigen"
	"github.com/podhmo/gos/pkg/maplib"
)

func main() {
	b := openapigen.NewBuilder(openapigen.DefaultConfig())

	// routing
	Error := openapigen.Define("Error", b.Object(
		b.Field("message", b.String()),
	)).Doc("default error")
	r := openapigen.NewRouter(Error)
	{
		r := r.Tagged("emoji")
		r.Post("/emoji/translate", design.Translate)
		r.Post("/emoji/suggest", design.Suggest)
	}

	// openapi data
	doc, err := maplib.Merge(orderedmap.New(), &openapigen.OpenAPI{
		OpenAPI: "3.0.3",
		Info: openapigen.Info{
			Title:   "emoji API",
			Version: "0.0.0",
			Doc:     "emoji API",
		},
		Servers: []openapigen.Server{
			{
				URL: "http://localhost:8080",
				Doc: "local development",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	r.ToSchemaWith(b, doc)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		panic(err)
	}
}
