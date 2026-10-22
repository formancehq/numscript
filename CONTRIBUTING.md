### Develop locally

You can update snaphshots with the
`UPDATE_SNAPS=true` variable while running the tests

The Numscript parser (`internal/parser`) is hand-written, so that it can recover from syntax errors: the LSP needs a meaningful AST for programs that are being typed. The [Numscript.g4](Numscript.g4) and [Lexer.g4](Lexer.g4) grammars are still the reference for the syntax, and they generate the previous, ANTLR-based parser (`internal/parser/internal/legacy`). The tests check that both parsers agree on every program of the repository, so if you change the syntax, update both the hand-written parser and the grammar.

You can generate the ANTLR parsers by running `just generate` (you'll need to install the antlr4 command first)
