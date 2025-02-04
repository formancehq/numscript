antlr4 -Dlanguage=Go Lexer.g4 Numscript.g4 -o internal/parser/antlrParser -package antlrParser
mv internal/parser/antlrParser/_lexer.go internal/parser/antlrParser/lexer.go
