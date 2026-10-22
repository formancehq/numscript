package parser_test

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/parser/internal/legacy"
	"github.com/kr/pretty"
)

// The hand-written parser replaced the ANTLR one. These tests check that both
// parsers agree on the corpus of numscript programs found in the repository:
//
//   - when the ANTLR parser accepts a program, the new parser must accept it and produce the same AST
//   - when the ANTLR parser rejects a program, the new parser must reject it too
//
// TODO remove these tests (and the legacy package) once the new parser has been released for a while.

// Programs the ANTLR parser rejects, but are now accepted on purpose
func isIntentionalDivergence(src string) bool {
	// a line comment at the end of the file without a trailing newline used to be a parsing error
	lastLine := src[strings.LastIndex(src, "\n")+1:]
	return strings.Contains(lastLine, "//")
}

var handwrittenCorpus = []string{
	// operators precedence and associativity
	`set_tx_meta("k", 1 + 2 + 3)`,
	`set_tx_meta("k", 1 - 2 - 3)`,
	`set_tx_meta("k", 1 / 2 / 3)`,
	`set_tx_meta("k", 1 + 2 / 3 - 4)`,
	`set_tx_meta("k", -1 / 2)`,
	`set_tx_meta("k", - -1)`,
	`set_tx_meta("k", -$x + 1)`,
	`set_tx_meta("k", (1 + 2) / 3)`,
	`set_tx_meta("k", -(1 + 2))`,

	// monetary literals, where the asset and amount expressions are juxtaposed
	`send [COIN -100] ( source = @a destination = @b )`,
	`send [$asset -100] ( source = @a destination = @b )`,
	`send [$a - 1 2] ( source = @a destination = @b )`,
	`send [$a - 1 - 2] ( source = @a destination = @b )`,
	`send [$a - 1 + 2] ( source = @a destination = @b )`,
	`send [$a - 1 - 2 - 3] ( source = @a destination = @b )`,
	`send [($a - 1) -2] ( source = @a destination = @b )`,
	`send [$a $b - $c] ( source = @a destination = @b )`,
	`send [$a -$b - $c] ( source = @a destination = @b )`,
	`send [COIN 10] + [COIN 20] ( source = @a destination = @b )`,
	`send [COIN 10] - $x ( source = @a destination = @b )`,
	`send [[COIN 1] -1] ( source = @a destination = @b )`,
	`send [COIN *] ( source = @a destination = @b )`,
	`send [COIN - 1 *] ( source = @a destination = @b )`,
	`send [COIN] ( source = @a destination = @b )`,
	`send [COIN 1 2] ( source = @a destination = @b )`,

	// account literals
	`set_tx_meta("k", @a:b:c)`,
	`set_tx_meta("k", @a:$b:c)`,
	`set_tx_meta("k", @a-b_C:0)`,
	`set_tx_meta("k", @a :b)`,
	`set_tx_meta("k", @a: b)`,
	`set_tx_meta("k", @ a)`,
	`set_tx_meta("k", @$a)`,
	`set_tx_meta("k", @a:)`,

	// sources and destinations
	`send $x ( source = { } destination = @b )`,
	`send $x ( source = { @a -@b } destination = @b )`,
	`send $x ( source = { @a @b \ "c" } destination = @b )`,
	`send $x ( source = { $p from @a remaining from @b } destination = @b )`,
	`send $x ( source = { remaining from @a } destination = @b )`,
	`send $x ( source = { 1/2 from @a @b } destination = @b )`,
	`send $x ( source = { @a 1/2 from @b } destination = @b )`,
	`send $x ( source = { max $m from @a @b } destination = @b )`,
	`send $x ( source = oneof { @a @b } destination = @b )`,
	`send $x ( source = oneof { } destination = @b )`,
	`send $x ( source = @a \ "c" with scaling through @s destination = @b )`,
	`send $x ( source = @a with scaling through @s \ "c" destination = @b )`,
	`send $x ( source = @a allowing overdraft up to [COIN 10] destination = @b )`,
	`send $x ( source = @a \ $c allowing unbounded overdraft destination = @b )`,
	`send $x ( source = @a destination = { remaining kept } )`,
	`send $x ( source = @a destination = { remaining to @b } )`,
	`send $x ( source = @a destination = { remaining kept remaining to @c } )`,
	`send $x ( source = @a destination = { max $m to @c remaining kept } )`,
	`send $x ( source = @a destination = { max $m kept } )`,
	`send $x ( source = @a destination = { } )`,
	`send $x ( source = @a destination = { 1/2 to @a 1/2 kept } )`,
	`send $x ( source = @a destination = { 1/2 to @a remaining kept } )`,
	`send $x ( source = @a destination = { 1/2 to { max $m to @b remaining kept } remaining kept } )`,
	`send $x ( source = @a destination = oneof { max $m to @c remaining kept } )`,
	`send $x ( source = @a destination = oneof { remaining kept } )`,
	`send $x ( destination = @b source = @a )`,

	// other statements
	`save [COIN 10] from @a`,
	`save [COIN *] from @a`,
	`save $m from @a`,
	`overdraft(@a, COIN)`,
	`f()`,
	`f(,)`,
	`f(1,)`,
	`f(1 2)`,

	// vars and flags
	`vars { }`,
	`vars { monetary $m = balance(@a, COIN) account $a }`,
	`vars { monetary $m = [COIN -1] }`,
	`vars { max $m }`,
	"#![feature(\"a\", \"b\")]\nsend $x ( source = @a destination = @b )",
	"#![feature()]",
	"#![feature(\"a\",)]",
	"vars { }\nvars { }",
	"set_tx_meta(\"k\", 1)\nvars { }",

	// lexing
	`set_tx_meta("k", 1_000)`,
	`set_tx_meta("k", 12.5%)`,
	`set_tx_meta("k", 12%)`,
	`set_tx_meta("k", 12.5)`,
	`set_tx_meta("k", 1.%)`,
	`set_tx_meta("k", "a\"b")`,
	`set_tx_meta("k", "a\"")`,
	`set_tx_meta("k", "a\\"b")`,
	`set_tx_meta("k", "a\\")`,
	`set_tx_meta("k", "a)`,
	`set_tx_meta("k", USD/2)`,
	`set_tx_meta("k", USD/ 2)`,
	`set_tx_meta("k", A1/02)`,
	`set_tx_meta("k", $_x)`,
	`set_tx_meta("k", $x1)`,
	`set_tx_meta("k", $1)`,
	`set_tx_meta("k", $X)`,
	"set_tx_meta(\"k\", /* a /* nested */ comment */ 1)",
	"set_tx_meta(\"k\", /* unterminated 1)",
	"set_tx_meta(\"k\", 1) // comment\n",
	"set_tx_meta(\"k\", 1)\r\n",
	"sendx",
	"send_",
	"#",
	"!",
	"Ω",
	"",
}

func TestDifferentialAgainstAntlrParser(t *testing.T) {
	corpus := collectCorpus(t)
	if len(corpus) < 300 {
		t.Fatalf("the corpus is suspiciously small: %d programs", len(corpus))
	}

	mutate := !testing.Short()
	checked := 0
	for name, src := range corpus {
		checkAgreement(t, name, src)
		checked++

		if mutate {
			// remove one word at a time, to get many slightly broken (and some still valid) programs.
			// Words are joined by newlines, so that line comments keep spanning a single word
			words := strings.Fields(src)
			if len(words) > 120 {
				continue
			}
			for i := range words {
				removed := slices.Delete(slices.Clone(words), i, i+1)
				checkAgreement(t, name+"#remove"+strconv.Itoa(i), strings.Join(removed, "\n"))

				duplicated := slices.Insert(slices.Clone(words), i, words[i])
				checkAgreement(t, name+"#duplicate"+strconv.Itoa(i), strings.Join(duplicated, "\n"))

				if i+1 < len(words) {
					swapped := slices.Clone(words)
					swapped[i], swapped[i+1] = swapped[i+1], swapped[i]
					checkAgreement(t, name+"#swap"+strconv.Itoa(i), strings.Join(swapped, "\n"))
				}
				checked += 3
			}
		}
	}
	t.Logf("checked %d programs", checked)
}

func checkAgreement(t *testing.T, name string, src string) {
	t.Helper()

	expected := legacy.Parse(src)
	got := parser.Parse(src)

	legacyOk := len(expected.Errors) == 0
	newOk := len(got.Errors) == 0

	switch {
	case legacyOk && !newOk:
		t.Errorf("%s: the new parser rejects a valid program:\n%s\n\n%s", name, src, parser.ParseErrorsToString(got.Errors, src))

	case !legacyOk && newOk:
		if isIntentionalDivergence(src) {
			return
		}
		t.Errorf("%s: the new parser accepts an invalid program:\n%s\n\nANTLR errors: %v", name, src, expected.Errors)

	case legacyOk && newOk:
		expectedProgram, gotProgram := expected.Value, got.Value
		expectedProgram.Comments, gotProgram.Comments = nil, nil
		if diff := pretty.Diff(expectedProgram, gotProgram); len(diff) != 0 {
			t.Errorf("%s: the ASTs differ for:\n%s\n\n%s", name, src, strings.Join(diff, "\n"))
		}
	}
}

func collectCorpus(t *testing.T) map[string]string {
	t.Helper()

	root := "../.."
	corpus := map[string]string{}
	for i, src := range handwrittenCorpus {
		corpus["handwritten#"+strconv.Itoa(i)] = src
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// the oracle package contains the grammar of a different language (the machine numscript)
			if d.Name() == "oracle" || d.Name() == "legacy" {
				return filepath.SkipDir
			}
			return nil
		}

		switch {
		case strings.HasSuffix(path, ".num"):
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			corpus[path] = string(content)

		case strings.HasSuffix(path, "_test.go"):
			// Numscript programs embedded in go tests
			for i, lit := range stringLiteralsOf(t, path) {
				if looksLikeNumscript(lit) {
					corpus[path+"#"+strconv.Itoa(i)] = lit
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return corpus
}

func stringLiteralsOf(t *testing.T, path string) []string {
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, path, nil, goparser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}

	var literals []string
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(lit.Value)
		if err == nil {
			literals = append(literals, value)
		}
		return true
	})
	return literals
}

func looksLikeNumscript(s string) bool {
	for _, keyword := range []string{"send ", "save ", "vars ", "set_tx_meta(", "set_account_meta("} {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

const benchmarkProgram = `vars {
  monetary $amount
  account $user
  portion $fee = 2.5%
}

send $amount (
  source = {
    $user allowing overdraft up to [USD/2 100]
    @users:$user:savings
    @world
  }
  destination = {
    $fee to @platform:fees
    remaining to {
      max [USD/2 1000] to @merchant
      remaining kept
    }
  }
)

set_tx_meta("processed", true_value(1 + 2 / 3))
`

func BenchmarkParse(b *testing.B) {
	b.Run("handwritten", func(b *testing.B) {
		for b.Loop() {
			parser.Parse(benchmarkProgram)
		}
	})
	b.Run("antlr", func(b *testing.B) {
		for b.Loop() {
			legacy.Parse(benchmarkProgram)
		}
	})
}
