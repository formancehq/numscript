# Numscript CLI

[![GitHub Release](https://img.shields.io/github/v/release/formancehq/numscript)](https://github.com/formancehq/numscript/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/formancehq/numscript.svg)](https://pkg.go.dev/github.com/formancehq/numscript)
[![Go](https://github.com/formancehq/numscript/actions/workflows/checks.yml/badge.svg)](https://github.com/formancehq/numscript/actions/workflows/checks.yml)
[![codecov](https://codecov.io/gh/formancehq/numscript/graph/badge.svg?token=njjqGhFQ2p)](https://codecov.io/gh/formancehq/numscript)

Numscript is the DSL used to express financial transaction within the [Formance](https://www.formance.com/) ledger.
You can try it in its [online playground](https://playground.numscript.org)

The CLI in this repo allows you to play with numscript locally, check if there are parsing or logic errors in your numscript files, and run the numscript language server

The language server features include:

- Diagnostics
- Hover on values
- Detect document symbols
- Go to definition

### Build locally
If you have the golang toolchain installed locally, you can run `go install ./internal/numscript/numscript.go` to build the cli from source, which you'll be able to run by executing the `numscript` command
Otherwise, you can download it from this repo's [releases](https://github.com/formancehq/numscript/releases)
