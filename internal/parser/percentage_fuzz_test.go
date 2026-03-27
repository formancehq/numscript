package parser

import (
	"testing"
)

func FuzzParsePercentageRatio(f *testing.F) {
	// Seed corpus with valid and edge case percentages
	f.Add("0%")
	f.Add("100%")
	f.Add("50%")
	f.Add("50.5%")
	f.Add("0.1%")
	f.Add("99.99%")
	f.Add("1%")
	f.Add("10.0%")
	f.Add("33.333%")
	f.Add("100.00%")
	f.Add("0.0%")
	// Invalid cases to test error handling
	f.Add("101%")
	f.Add("-5%")
	f.Add("abc%")
	f.Add("%")
	f.Add("50")
	f.Add("50.%")
	f.Add(".50%")

	f.Fuzz(func(t *testing.T, input string) {
		// Call ParsePercentageRatio and ensure it doesn't panic
		num, floatingDigits, err := ParsePercentageRatio(input)

		if err == nil {
			// If parsing succeeded, verify the result is reasonable
			if num == nil {
				t.Errorf("ParsePercentageRatio succeeded but returned nil num for input: %q", input)
			}
			// Note: negative numbers are allowed - the parser doesn't enforce constraints
			// Floating digits should be limited to 18 (now enforced by the parser)
			if floatingDigits > 18 {
				t.Errorf("ParsePercentageRatio returned floatingDigits=%d > 18 for input: %q", floatingDigits, input)
			}
		}
	})
}
