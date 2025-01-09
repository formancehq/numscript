package encode

import (
	"encoding/binary"
	"io"

	"github.com/formancehq/numscript/internal/parser"
)

// TODO remove iota and hardcode value

const StatementTypeSend byte = iota

// sized_string= len:u16 ...byte

const (
	// i64
	ExprLongIntLit byte = iota

	// sized_string
	ExprAsset

	// sized_string
	ExprAccount

	// NewMonetary(asset: expr, amt: expr)
	ExprNewMonetary
)

const (
	SourceTypeAccount byte = iota
	SourceTypeCapped
	SourceTypeInorder
	SourceTypeAllotment
)

const (
	DestTypeAccount byte = iota
	DestTypeInorder
	DestTypeAllotment
)

type State struct {
	writer io.Writer
}

func (b State) encodeSource(source parser.Source) error {
	switch source := source.(type) {
	case *parser.SourceAccount:
		_, err := b.writer.Write([]byte{SourceTypeAccount})
		if err != nil {
			return err
		}
		return b.encodeValueExpr(source.ValueExpr)

	default:
		panic("TODO src")
	}
}

func (b State) encodeDest(dest parser.Destination) error {
	switch dest := dest.(type) {
	case *parser.DestinationAccount:
		_, err := b.writer.Write([]byte{DestTypeAccount})
		if err != nil {
			return err
		}
		return b.encodeValueExpr(dest.ValueExpr)

	default:
		panic("TODO src")
	}
}

func (b State) encodeUInt32(n uint32) error {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, n)
	_, err := b.writer.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (b State) encodeString(s string) error {
	bs := []byte(s)
	_, err := b.writer.Write(bs)
	if err != nil {
		return err
	}

	_, err = b.writer.Write([]byte{0x00})
	if err != nil {
		return err
	}

	return nil
}

func (b State) encodeValueExpr(expr parser.ValueExpr) error {
	switch expr := expr.(type) {
	case *parser.MonetaryLiteral:
		_, err := b.writer.Write([]byte{ExprNewMonetary})
		if err != nil {
			return err
		}
		err = b.encodeValueExpr(expr.Asset)
		if err != nil {
			return err
		}
		err = b.encodeValueExpr(expr.Amount)
		if err != nil {
			return err
		}
		return nil

	case *parser.AssetLiteral:
		_, err := b.writer.Write([]byte{ExprAsset})
		if err != nil {
			return err
		}
		return b.encodeString(expr.Asset)

	case *parser.AccountLiteral:
		_, err := b.writer.Write([]byte{ExprAccount})
		if err != nil {
			return err
		}
		return b.encodeString(expr.Name)

	case *parser.NumberLiteral:
		_, err := b.writer.Write([]byte{ExprLongIntLit})
		if err != nil {
			return err
		}
		return b.encodeUInt32(uint32(expr.Number))

	default:
		panic("TODO valueExpr")
	}

}

func (b State) encodeSend(statement parser.SendStatement) error {

	switch sentValue := statement.SentValue.(type) {

	case *parser.SentValueLiteral:
		_, err := b.writer.Write([]byte{StatementTypeSend})
		if err != nil {
			return err
		}
		err = b.encodeValueExpr(sentValue.Monetary)
		if err != nil {
			return err
		}

	case *parser.SentValueAll:
		panic("TODO sent*")

	}

	err := b.encodeSource(statement.Source)
	if err != nil {
		return err
	}

	err = b.encodeDest(statement.Destination)
	if err != nil {
		return err
	}

	return nil
}

func (b State) encodeStatement(statement parser.Statement) error {
	switch statement := statement.(type) {
	case *parser.SendStatement:
		return b.encodeSend(*statement)

	case *parser.SaveStatement:
		panic("TODO")
	case *parser.FnCall:
		panic("TODO")
	default:
		panic("TODO")
	}

}

func Encode(src parser.Program, writer io.Writer) error {
	buf := State{writer}
	for _, statement := range src.Statements {
		err := buf.encodeStatement(statement)
		if err != nil {
			return err
		}
	}
	return nil
}
