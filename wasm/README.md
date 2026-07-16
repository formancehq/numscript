# numscript WASM bindings

Compiles the numscript parser, static analysis, and interpreter to WebAssembly
so consumers (e.g. the numscript playground) can parse, analyze, and run scripts
client-side, without a server.

## Build

```bash
./wasm/build.sh
```

Outputs to `wasm/dist/`:

- `numscript.wasm` — the module
- `wasm_exec.js` — Go's JS glue (must match the Go toolchain used to build)

## JS API

After loading the glue and instantiating the module, `globalThis.__numscript`
exposes three functions. All are pure — nothing is retained across calls except
an internal content-keyed parse cache.

- `analyze(source: string): string` — JSON array of diagnostics (parsing errors
  + static analysis) in Monaco's shape: `{ startLineNumber, startColumn,
  endLineNumber, endColumn, message, severity }` (1-based; `severity` is
  `"error"` or `"warning"`).
- `run(source: string, argsJson: string): string` — `argsJson` is
  `{ variables, balances, metadata, featureFlags }`. Returns JSON
  `{ ok: true, value: { postings, txMeta, accountsMeta } }` or
  `{ ok: false, error }`. Refuses to execute when there are error-severity
  diagnostics; warnings still run.
- `configureCache(size: number): void` — max parses memoized by the LRU
  (default 1). A hit is byte-identical source, so `analyze` then `run` on the
  same source parse once; there is no lifecycle to manage.
