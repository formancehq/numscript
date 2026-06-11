# Numscript CLI

[![GitHub Release](https://img.shields.io/github/v/release/formancehq/numscript)](https://github.com/formancehq/numscript/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/formancehq/numscript.svg)](https://pkg.go.dev/github.com/formancehq/numscript)
[![Go](https://github.com/formancehq/numscript/actions/workflows/main.yml/badge.svg)](https://github.com/formancehq/numscript/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/formancehq/numscript/graph/badge.svg?token=njjqGhFQ2p)](https://codecov.io/gh/formancehq/numscript)

Numscript is the DSL used to express financial transaction within the [Formance](https://www.formance.com/) ledger.
You can try it in its [online playground](https://playground.numscript.org)

The CLI in this repo allows you to play with numscript locally, check if there are parsing or logic errors in your numscript files, and run the numscript language server

The language server features include:

- Diagnostics
- Hover on values
- Detect document symbols
- Go to definition

### Installation

You can install the `numscript` cli with one of the following ways:

**Using curl**

```sh
curl -sSf https://raw.githubusercontent.com/formancehq/numscript/main/install.sh | bash
```

**Using golang toolchain**

```sh
go install github.com/formancehq/numscript/cmd/numscript@latest
```

### Telemetry

Release builds of the `numscript` CLI send anonymous crash reports to [Sentry](https://sentry.io) when the program panics. A crash report contains the panic message, a stack trace, and the CLI version — nothing else. No reports are sent during normal operation, and development builds (built from source with `go install` or `go build`, where the version is `develop`) never send anything.

To disable crash reporting entirely, set the `NUMSCRIPT_NO_TELEMETRY` environment variable to any non-empty value:

```sh
export NUMSCRIPT_NO_TELEMETRY=1
```
