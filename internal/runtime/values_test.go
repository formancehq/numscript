package runtime

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name  string
		fn    func(string) bool
		valid []string
		bad   []string
	}{
		{
			name:  "account",
			fn:    ValidateAccount,
			valid: []string{"world", "users:001", "a-b_c", "a:b:c"},
			bad:   []string{"", "users:", ":users", "users::001", "a b", "@world"},
		},
		{
			name:  "asset",
			fn:    ValidateAsset,
			valid: []string{"COIN", "USD/2", "EUR", "A", "TOKEN_X", "USD/123456"},
			bad:   []string{"", "usd", "1USD", "USD/", "USD/1234567", "USD 2"},
		},
		{
			name:  "color",
			fn:    ValidateColor,
			valid: []string{"", "RED", "ABC"},
			bad:   []string{"red", "Red", "RED1", "RED_X", " RED"},
		},
		{
			name:  "scope",
			fn:    ValidateScope,
			valid: []string{"", "s", "a_b", "s1"},
			bad:   []string{"S", "a-b", "a:b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range tc.valid {
				require.True(t, tc.fn(v), "expected %q to be valid", v)
			}
			for _, v := range tc.bad {
				require.False(t, tc.fn(v), "expected %q to be invalid", v)
			}
		})
	}
}

func TestParseNumber(t *testing.T) {
	n, ok := ParseNumber("42")
	require.True(t, ok)
	require.Zero(t, n.Cmp(big.NewInt(42)))

	n, ok = ParseNumber("-7")
	require.True(t, ok)
	require.Zero(t, n.Cmp(big.NewInt(-7)))

	_, ok = ParseNumber("4.2")
	require.False(t, ok)
	_, ok = ParseNumber("")
	require.False(t, ok)
	_, ok = ParseNumber("0x10")
	require.False(t, ok)
}

func TestParsePortion(t *testing.T) {
	testCases := []struct {
		input string
		want  *big.Rat
	}{
		{"1/2", big.NewRat(1, 2)},
		{"1 / 2", big.NewRat(1, 2)},
		{"0/2", big.NewRat(0, 1)},
		{"2/2", big.NewRat(1, 1)},
		{"50%", big.NewRat(1, 2)},
		{"12.5%", big.NewRat(1, 8)},
		{"0%", big.NewRat(0, 1)},
		{"100%", big.NewRat(1, 1)},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParsePortion(tc.input)
			require.NoError(t, err)
			require.Zero(t, got.Cmp(tc.want), "got %s", got)
		})
	}

	errCases := []struct {
		input string
		msg   string
	}{
		{"", "invalid format"},
		{"half", "invalid format"},
		{"1/2/3", "invalid format"},
		{"200%", "between 0% and 100%"},
		{"3/2", "between 0% and 100%"},
		{"-1/2", "invalid format"},
		{"1/0", "invalid fractional format"},
	}
	for _, tc := range errCases {
		t.Run("error: "+tc.input, func(t *testing.T) {
			_, err := ParsePortion(tc.input)
			require.ErrorContains(t, err, tc.msg)
		})
	}
}

func TestParseMonetary(t *testing.T) {
	asset, amount, err := ParseMonetary("USD/2 100")
	require.NoError(t, err)
	require.Equal(t, "USD/2", asset)
	require.Zero(t, amount.Cmp(big.NewInt(100)))

	_, _, err = ParseMonetary("USD/2 -1")
	require.NoError(t, err)

	errCases := []struct {
		input string
		msg   string
	}{
		{"", "invalid monetary"},
		{"USD/2", "invalid monetary"},
		{"USD/2 100 extra", "invalid monetary"},
		{"usd 100", "invalid asset"},
		{"USD/2 abc", "invalid monetary amount"},
	}
	for _, tc := range errCases {
		t.Run("error: "+tc.input, func(t *testing.T) {
			_, _, err := ParseMonetary(tc.input)
			require.ErrorContains(t, err, tc.msg)
		})
	}
}
