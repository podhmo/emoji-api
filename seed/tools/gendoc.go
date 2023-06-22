package main

import (
	"encoding/json"
	"os"

	"github.com/iancoleman/orderedmap"
	"github.com/podhmo/gos/openapigen"
	"github.com/podhmo/gos/pkg/maplib"
)

func main() {
	b := openapigen.NewBuilder(openapigen.DefaultConfig())

	// types
	Name := openapigen.Define("Name", b.String()).Doc("name of something")
	openapigen.Define("DateTime", b.String().Format("date-time")) // for ReferenceByName

	Task := openapigen.Define("Task", b.Object(
		b.Field("name", b.Reference(Name)),
		b.Field("done", b.Bool()),
		b.Field("createdAt", b.ReferenceByName("DateTime")),
	))

	ListTask := b.Action("ListTask",
		b.Input(b.Param("sort", b.String().Enum([]string{"createdAt", "-createdAt"})).AsQuery()),
		b.Output(b.Array(Task)),
	)

	// routing
	Error := openapigen.Define("Error", b.Object(
		b.Field("message", b.String()),
	)).Doc("default error")
	r := openapigen.NewRouter(Error)
	{
		r := r.Tagged("task")
		r.Get("/tasks", ListTask)
	}

	// openapi data
	doc, err := maplib.Merge(orderedmap.New(), &openapigen.OpenAPI{
		OpenAPI: "3.0.3",
		Info: openapigen.Info{
			Title:   "task API",
			Version: "0.0.0",
			Doc:     "simple list tasks API",
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
