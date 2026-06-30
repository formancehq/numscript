package interpreter

import (
	"github.com/formancehq/numscript/internal/parser"
)

func setTxMeta(st *programState, r parser.Range, args []Value) InterpreterError {
	p := NewArgsParser(args)
	key := parseArg(p, r, expectString)
	meta := parseArg(p, r, expectAnything)
	err := p.parse()
	if err != nil {
		return err
	}

	st.TxMeta[string(key)] = meta
	return nil
}

func setAccountMeta(st *programState, r parser.Range, args []Value) InterpreterError {
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	key := parseArg(p, r, expectString)
	meta := parseArg(p, r, expectAnything)
	err := p.parse()
	if err != nil {
		return err
	}

	st.SetAccountsMeta.Set(account.Name, account.Scope, string(key), meta)

	return nil
}
