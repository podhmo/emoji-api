seed:
	go run ./seed/tools/gendoc.go > openapi.json
	go run github.com/getkin/kin-openapi/cmd/validate@latest openapi.json
.PHONY: seed

gen:
	mkdir -p api/oapigen
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./seed/tools/oapi-conf.yaml openapi.json > api/oapigen/server.go 
.PHONY: gen

stub:
	go run ./seed/tools/gen-stub --doc openapi.json --src ./api/oapigen --dst ./api/controller