package encode_test

import (
	"bytes"
	"testing"

	"github.com/formancehq/numscript/internal/encode"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestEncodeSendAccount(t *testing.T) {
	src := `
send [COIN 10] (
  source = @src
  destination = @dest
)
 `

	bs := Encode(`
send [COIN 10] (
  source = @src
  destination = @dest
)
 `)

	require.Equal(t, []byte{
		encode.StatementTypeSend,
		encode.ExprNewMonetary,
		encode.ExprAsset,

		// COIN
		0x43, 0x4F, 0x49, 0x4E, 0x00,

		encode.ExprLongIntLit, 0, 0, 0, 0xa,

		// source=
		encode.SourceTypeAccount,
		encode.ExprAccount,
		// 'src'
		0x73, 0x72, 0x63, 0x00,
		// destination=
		encode.DestTypeAccount,
		encode.ExprAccount,
		// 'dest'
		0x64, 0x65, 0x73, 0x74, 0x00,
	}, bs)

	require.Equal(t, len([]byte(src)), len(bs))
}

func Encode(src string) []byte {
	res := parser.Parse(src)
	w := bytes.NewBuffer(nil)
	err := encode.Encode(res.Value, w)
	if err != nil {
		panic(err)
	}
	return w.Bytes()
}
