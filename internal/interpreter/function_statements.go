package interpreter

import (
	"github.com/formancehq/numscript/internal/flags"
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

	if err := rejectScopedAccountMeta(meta, r); err != nil {
		return err
	}

	st.TxMeta[string(key)] = meta
	return nil
}

func rejectScopedAccountMeta(value Value, r parser.Range) InterpreterError {
	if account, ok := value.(AccountAddress); ok && account.Scope != "" {
		return CannotStoreScopedAccountInMeta{Range: r, Account: account.Name, Scope: account.Scope}
	}
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

	if err := rejectScopedAccountMeta(meta, r); err != nil {
		return err
	}

	st.SetAccountsMeta.Set(account.Name, account.Scope, string(key), meta)

	return nil
}

func deleteAccountMetadata(st *programState, r parser.Range, args []Value) InterpreterError {
	if err := st.checkFeatureFlag(flags.ExperimentalDeleteAccountMetadata); err != nil {
		return err
	}

	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	key := parseArg(p, r, expectString)
	err := p.parse()
	if err != nil {
		return err
	}

	st.SetAccountsMeta.Delete(account.Name, account.Scope, string(key))
	if err := st.Store.DeleteAccountMetadata(st.ctx, account.Name, account.Scope, string(key)); err != nil {
		return DeleteMetadataError{Range: r, WrappedError: err}
	}
	return nil
}
