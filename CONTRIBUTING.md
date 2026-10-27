### Develop locally

You can update snaphshots with the
`UPDATE_SNAPS=true` variable while running the tests

If you need to update the grammar, you can generate the parser by running `just generate` (you'll need to install the antlr4 command first)

The Nix development shell uses upstream Go 1.27.3 from the dedicated locked
`nixpkgs-go` stable input, allowing reuse of the upstream compiler binary cache.
The original stable, unstable and NUR tooling pins remain independent.
Golangci-lint retains its 2.15.0 recipe and uses that stable input's matching
`buildGo127Module` builder. The shell sets `GOTOOLCHAIN=local` so Go commands
use the selected compiler without downloading another toolchain.
