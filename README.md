# Numscript CLI

[![Go](https://github.com/formancehq/numscript/actions/workflows/checks.yml/badge.svg)](https://github.com/formancehq/numscript/actions/workflows/checks.yml) [![codecov](https://codecov.io/gh/formancehq/numscript/graph/badge.svg?token=njjqGhFQ2p)](https://codecov.io/gh/formancehq/numscript)

Numscript is the DSL used to express financial transaction within the [Formance](https://www.formance.com/) ledger

The CLI in this repo allows you to play with numscript locally, check if there are parsing or logic errors in your numscript files, and run the numscript language server

The language server features include:

- Diagnostics
- Hover on values
- Detect document symbols
- Go to definition

### Develop locally

You can update snaphshots with the
`UPDATE_SNAPS=true` variable while running the tests

If you need to update the grammar, you can generate the parser using the `generate-parser.sh` script (you'll need to install the antlr4 command first)
