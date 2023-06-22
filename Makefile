gen:
	go run ./seed/tools/gendoc.go > openapi.json
	go run github.com/getkin/kin-openapi/cmd/validate@latest openapi.json
.PHONY: gen