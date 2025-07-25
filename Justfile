set dotenv-load

default:
    @just --list

pre-commit: generate tidy lint
pc: pre-commit

lint:
    @golangci-lint run --fix --build-tags it --timeout 5m

tidy:
    @go mod tidy

generate:
    @antlr4 -Dlanguage=Go Lexer.g4 Numscript.g4 -o internal/parser/antlrParser -package antlrParser
    @mv internal/parser/antlrParser/_lexer.go internal/parser/antlrParser/lexer.go

tests:
    @go test -race -covermode=atomic \
        -coverprofile coverage.txt \
        ./...
test:
    @go test -race -covermode=atomic -coverprofile coverage.txt ./...

release-local:
    @goreleaser release --nightly --skip=publish --clean

release-ci:
    @goreleaser release --nightly --clean

release:
    @goreleaser release --clean
