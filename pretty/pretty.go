package pretty

import "strings"

type state struct {
	// width   int
	builder strings.Builder
}

type Document interface {
	render(*state)
}

type (
	text   struct{ Text string }
	concat struct{ Documents []Document }
	lines  struct{ Lines uint }
	nest   struct{ Document Document }
)

func (d text) render(s *state) {
	s.builder.WriteString(d.Text)
}

func (d concat) render(s *state) {
	for _, d := range d.Documents {
		d.render(s)
	}
}

func (d lines) render(s *state) {
	for i := uint(0); i <= d.Lines; i++ {
		s.builder.WriteByte('\n')
	}
}

func (d nest) render(s *state) {
	d.Document.render(s)
}

// Public API
type PrintBuilder struct {
	maxWidth          int
	indentationSymbol string
	nestSize          int
}

func NewPrintBuilder() PrintBuilder {
	return PrintBuilder{
		maxWidth:          80,
		indentationSymbol: " ",
		nestSize:          2,
	}
}

func (b PrintBuilder) WithMaxWidth(maxWidth int) PrintBuilder {
	b.maxWidth = maxWidth
	return b
}

func (b PrintBuilder) WithNestSize(nestSize int) PrintBuilder {
	b.nestSize = nestSize
	return b
}

func PrintDefault(d Document) string {
	return NewPrintBuilder().Print(d)
}

func (b PrintBuilder) Print(d Document) string {
	state := state{}
	d.render(&state)
	return state.builder.String()
}
