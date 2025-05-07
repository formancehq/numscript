package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type irEmitterState struct {
	currentAsset string

	featureFlags map[string]struct{}

	// the evaluated value of each var
	vars map[string]Value
}

func (s irEmitterState) checkFeatureFlag(flag string) InterpreterError {
	_, ok := s.featureFlags[flag]
	if ok {
		return nil
	} else {
		return ExperimentalFeature{FlagName: flag}
	}
}

// func (s *irEmitterState) populateVars(varDeclaration parser.VarDeclaration, rawVars map[string]string) InterpreterError {
// 	if varDeclaration.Origin == nil {
// 		raw, ok := rawVars[varDeclaration.Name.Name]
// 		if !ok {
// 			return MissingVariableErr{Name: varDeclaration.Name.Name}
// 		}

// 		parsed, err := parseVar(varDeclaration.Type.Name, raw, varDeclaration.Type.Range)
// 		if err != nil {
// 			return err
// 		}
// 		s.vars[varDeclaration.Name.Name] = parsed
// 	} else {
// 		value, err := s.handleFnOrigin(varDeclaration.Type.Name, *varDeclaration.Origin)
// 		if err != nil {
// 			return err
// 		}
// 		s.vars[varDeclaration.Name.Name] = value
// 	}

// 	return nil
// }

func (s *irEmitterState) emitSource(astSource parser.Source) (Source, InterpreterError) {
	switch astSource := astSource.(type) {
	case *parser.SourceAccount:
		account, err := evaluateExprAs__IR(s, astSource.ValueExpr, expectAccount)
		if err != nil {
			return nil, err
		}
		color, err := evaluateExprAs__IR(s, astSource.Color, expectString)
		if err != nil {
			return nil, err
		}
		return SourceAccount{
			Range:     astSource.GetRange(),
			Account:   *account,
			Color:     *color,
			Overdraft: big.NewInt(0),
		}, nil
	case *parser.SourceOverdraft:
		account, err := evaluateExprAs__IR(s, astSource.Address, expectAccount)
		if err != nil {
			return nil, err
		}
		color, err := evaluateExprAs__IR(s, astSource.Color, expectString)
		if err != nil {
			return nil, err
		}
		overdraft, err := evaluateExprAs__IR(s, *astSource.Bounded, expectMonetaryOfAsset(s.currentAsset))
		if err != nil {
			return nil, err
		}
		return SourceAccount{
			Range:     astSource.Range,
			Account:   *account,
			Color:     *color,
			Overdraft: overdraft,
		}, nil
	case *parser.SourceCapped:
		cap, err := evaluateExprAs__IR(s, astSource.Cap, expectMonetaryOfAsset(s.currentAsset))
		if err != nil {
			return nil, err
		}
		from, err := s.emitSource(astSource.From)
		if err != nil {
			return nil, err
		}
		return SourceCapped{
			Range: astSource.Range,
			From:  from,
			Cap:   cap,
		}, nil
	case *parser.SourceInorder:
		sources, err := s.emitSources(astSource.Sources)
		if err != nil {
			return nil, err
		}
		return SourceInorder{
			Range:   astSource.GetRange(),
			Sources: sources,
		}, nil

	case *parser.SourceOneof:
		sources, err := s.emitSources(astSource.Sources)
		if err != nil {
			return nil, err
		}
		return SourceInorder{
			Range:   astSource.GetRange(),
			Sources: sources,
		}, nil

	case *parser.SourceAllotment:
		panic("TODO")

	default:
		utils.NonExhaustiveMatchPanic[Source](astSource)
		return nil, nil
	}

}

func (s *irEmitterState) emitSources(astSources []parser.Source) ([]Source, InterpreterError) {
	sources := make([]Source, len(astSources))
	for index, astSource := range astSources {
		source, err := s.emitSource(astSource)
		if err != nil {
			return nil, err
		}
		sources[index] = source
	}
	return sources, nil
}
