gen:
	go run ./seed/tools/gendoc.go > openapi.json
.PHONY: gen