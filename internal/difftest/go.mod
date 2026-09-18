module github.com/formancehq/numscript/internal/difftest

go 1.26.0

require (
	github.com/formancehq/go-libs/v5 v5.6.1
	github.com/formancehq/numscript v0.0.0-00010101000000-000000000000
	github.com/formancehq/numscript/internal/oracle v0.0.0-00010101000000-000000000000
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v1.4.10 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/mailru/easyjson v0.9.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/formancehq/numscript => ../..

replace github.com/formancehq/numscript/internal/oracle => ../oracle
