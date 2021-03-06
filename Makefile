build:
	go build 

install: 
	go install

go-test:
	go test -cover -v $$(go list ./... | grep -v -E '(vendor|docs)' ) | grep -v "no test files"

fmt:
	find -name '*.go' | grep -v .escape | grep -v vendor | xargs -n 1 go fmt

docs-build:
	escape run release -f --skip-tests --skip-deploy && cd ../escape/ && make docs-build && cd - 
