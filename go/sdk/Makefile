.PHONY: gen tidy

gen:
	oapi-codegen -config oapi-codegen.yaml ../api/openapi.json

tidy:
	go mod tidy
