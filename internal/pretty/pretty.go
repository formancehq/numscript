package pretty

import (
	"strings"
)

type Mode = byte

const (
	ModeUnbroken Mode = iota
	ModeBroken
)

type queueItem struct {
	doc         Document
	indentation int
	mode        Mode
}

type state struct {
	opt PrintBuilder

	width   int
	builder strings.Builder
	queue   Queue[queueItem]
}

type Document interface {
	render(s *state, mode Mode, indentation int)
	fit(s *state, mode Mode, indentation int, width int)
}

type (
	text   struct{ Text string }
	concat struct{ Documents []Document }
	lines  struct{ Lines uint }
	nest   struct{ Document Document }
	break_ struct{ Unbroken string }
	group  struct{ Document Document }
)

// Fits
func (d text) fit(s *state, mode Mode, indentation int, width int) {
	// width -= doc.text.length;
}

func (d concat) fit(s *state, mode Mode, indentation int, width int) {
}

func (d lines) fit(s *state, mode Mode, indentation int, width int) {
}

func (d break_) fit(s *state, mode Mode, indentation int, width int) {
}

func (d nest) fit(s *state, mode Mode, indentation int, width int) {
}

func (d group) fit(s *state, mode Mode, indentation int, width int) {
}

// Render
func (d text) render(s *state, mode Mode, indentation int) {
	s.builder.WriteString(d.Text)
}

func (d concat) render(s *state, mode Mode, indentation int) {
	// TODO reverse
	for _, doc := range d.Documents {
		// d.render(s)
		s.queue.PushFront(queueItem{
			mode:        mode,
			indentation: indentation,
			doc:         doc,
		})
	}
}

func (d lines) render(s *state, mode Mode, indentation int) {
	for i := uint(0); i <= d.Lines; i++ {
		s.builder.WriteByte('\n')
	}
}

func (d break_) render(s *state, mode Mode, indentation int) {
	s.builder.WriteString(d.Unbroken)
}

func (d nest) render(s *state, mode Mode, indentation int) {
	s.queue.PushFront(
		queueItem{
			mode: mode,
			// this assumes nest character has line 1
			indentation: indentation + s.opt.nestSize,
			doc:         d.Document,
		},
	)
}

func (d group) render(s *state, mode Mode, indentation int) {
	// const fit = fits(maxWidth - width, nestSize, {
	// 	indentation,
	// 	mode: "unbroken",
	// 	doc,
	// 	tail: null,
	//   });

	var nestedMode Mode
	if s.fits(s.opt.maxWidth - s.width) {
		nestedMode = ModeUnbroken
	} else {
		nestedMode = ModeBroken
	}

	s.queue.PushFront(queueItem{
		mode:        nestedMode,
		indentation: indentation,
		doc:         d.Document,
	})
}

// Public API
type PrintBuilder struct {
	maxWidth          int
	indentationSymbol rune
	nestSize          int
}

func NewPrintBuilder() PrintBuilder {
	return PrintBuilder{
		maxWidth:          80,
		indentationSymbol: ' ',
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

func (s *state) fits(width int) bool {
	for !s.queue.IsEmpty() {
		if width < 0 {
			return false
		}

		item := s.queue.Pop()
		item.doc.fit(s, item.mode, item.indentation, width)
	}

	return true
}

func (b PrintBuilder) Print(d Document) string {
	state := state{
		queue: NewQueue[queueItem](),
		width: 0,
		opt:   b,
	}
	state.queue.PushFront(queueItem{
		mode:        ModeUnbroken,
		indentation: 0,
		doc:         group{Document: d},
	})

	for {
		if state.queue.IsEmpty() {
			break
		}
		item := state.queue.Pop()
		item.doc.render(&state, item.mode, item.indentation)
	}

	return state.builder.String()
}
