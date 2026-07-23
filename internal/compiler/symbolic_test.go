package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func encode(t *testing.T, source string) *Encoding {
	t.Helper()
	program := parser.Parse(source)
	require.Empty(t, program.Errors)
	compiled, cErr := compileProgramToIR(program.Value)
	require.Nil(t, cErr)
	enc, err := simulate(compiled.instructions)
	require.NoError(t, err)
	return enc
}

// Golden snapshot of the emitted SMT-LIB2 for the simplest program. Guards
// against unintended changes to the encoding.
func TestSymbolicSimpleGolden(t *testing.T) {
	enc := encode(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)
	snaps.MatchInlineSnapshot(t, "\n"+enc.SMTLIB, snaps.Inline(`
(declare-const start_src_USD_2 Int)
(assert (>= start_src_USD_2 0))
(declare-const avail_0 Int)
(assert (= avail_0 (ite (< (ite (<= 10 (ite (< (+ start_src_USD_2 (ite (> 0 0) 0 0)) 0) 0 (+ start_src_USD_2 (ite (> 0 0) 0 0)))) 10 (ite (< (+ start_src_USD_2 (ite (> 0 0) 0 0)) 0) 0 (+ start_src_USD_2 (ite (> 0 0) 0 0)))) 0) 0 (ite (<= 10 (ite (< (+ start_src_USD_2 (ite (> 0 0) 0 0)) 0) 0 (+ start_src_USD_2 (ite (> 0 0) 0 0)))) 10 (ite (< (+ start_src_USD_2 (ite (> 0 0) 0 0)) 0) 0 (+ start_src_USD_2 (ite (> 0 0) 0 0)))))))
(declare-const sent_0 Int)
(assert (= sent_0 (+ 0 avail_0)))
(declare-const start_dest_USD_2 Int)
(assert (>= start_dest_USD_2 0))
(declare-const fail Bool)
(assert (= fail (< avail_0 10)))
(declare-const sent_total_src_USD_2 Int)
(assert (= sent_total_src_USD_2 (ite fail 0 avail_0)))
(declare-const recv_total_dest_USD_2 Int)
(assert (= recv_total_dest_USD_2 (ite fail 0 sent_0)))
`))
}

func TestSymbolicSymbolTable(t *testing.T) {
	enc := encode(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)
	src := AccountAsset{"src", "USD/2"}
	dest := AccountAsset{"dest", "USD/2"}
	require.Equal(t, "fail", enc.Symbols.Fail)
	require.Equal(t, "start_src_USD_2", enc.Symbols.Start[src])
	require.Equal(t, "sent_total_src_USD_2", enc.Symbols.Sent[src])
	require.Equal(t, "recv_total_dest_USD_2", enc.Symbols.Received[dest])
}

// Deferred opcodes must fail loudly with a typed error, never silently produce
// a wrong encoding.
func TestSymbolicUnsupported(t *testing.T) {
	program := parser.Parse(`
		set_tx_meta("k", "v")
		send [USD/2 1] (
			source = @world
			destination = @a
		)
	`)
	require.Empty(t, program.Errors)
	compiled, cErr := compileProgramToIR(program.Value)
	require.Nil(t, cErr)
	_, err := simulate(compiled.instructions)
	require.Error(t, err)
	_, ok := err.(*UnsupportedOpError)
	require.True(t, ok, "expected *UnsupportedOpError, got %T: %v", err, err)
}

// Drift guard: every concrete irInstr type must be handled by the switch
// (returning nil or a typed *UnsupportedOpError), never panic or return an
// untyped error. If a new IR instruction type is added to ir_instr.go, add it
// here and give it a case in interp.step.
func TestSymbolicDriftGuard(t *testing.T) {
	allTypes := []irInstr{
		pullAccount{},
		sendToAccount{},
		save{},
		makeAllotment{},
		checkEnoughFunds{},
		assertLeftover{},
		setCurrentAsset{},
		assertSameAsset{},
		assertValidAccount{},
		assertNonNegativeBalance{},
		setTxMeta{},
		setAccountMeta{},
		metaVar{},
		fetchBalance{},
		loadVar{},
		jmpIfZero{},
		loadInt{},
		loadStr{},
		binaryOp{},
		unaryOp{},
		labelMarker{},
	}

	for _, instr := range allTypes {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("step(%T) panicked: %v", instr, r)
				}
			}()
			it := &interp{
				regs:       map[reg]term{},
				curBalance: map[AccountAsset]string{},
				sentRaw:    map[AccountAsset]string{},
				recvRaw:    map[AccountAsset]string{},
				pool:       "0",
				failExpr:   "false",
				seen:       map[AccountAsset]bool{},
				sym: SymbolTable{
					Start:    map[AccountAsset]string{},
					Sent:     map[AccountAsset]string{},
					Received: map[AccountAsset]string{},
					Vars:     map[string]string{},
				},
			}
			err := it.step(instr)
			if err != nil {
				if _, ok := err.(*UnsupportedOpError); !ok {
					t.Errorf("step(%T) returned non-UnsupportedOpError: %v", instr, err)
				}
			}
		}()
	}
}
