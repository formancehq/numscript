package interpreter

import (
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
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

	accountMeta := utils.MapGetOrPutDefault(st.SetAccountsMeta, string(account), func() AccountMetadata {
		return AccountMetadata{}
	})

	accountMeta[string(key)] = meta.String()

	return nil
}
