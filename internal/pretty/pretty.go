package pretty

import (
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/utils"
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
	fit(s *state, mode Mode, indentation int)
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
func (d text) fit(s *state, mode Mode, indentation int) {
	s.width -= len(d.Text)
}

func (d concat) fit(s *state, mode Mode, indentation int) {
	// Push docs in reverse order
	docsCopy := slices.Clone(d.Documents)
	slices.Reverse(docsCopy)

	for _, doc := range docsCopy {
		s.queue.PushFront(queueItem{
			mode:        mode,
			indentation: indentation,
			doc:         doc,
		})
	}
}

func (d lines) fit(s *state, mode Mode, indentation int) {
	// Lines always fits
}

func (d break_) fit(s *state, mode Mode, indentation int) {
	s.queue = NewQueue[queueItem]()
	// switch mode {
	// case ModeBroken:
	// 	// Exit immediately by emptying the queue

	// case ModeUnbroken:
	// 	s.width -= len(d.Unbroken)

	// default:
	// 	utils.NonExhaustiveMatchPanic[any](mode)
	// }
}

func (d nest) fit(s *state, mode Mode, indentation int) {
	s.queue.PushFront(queueItem{
		mode:        mode,
		doc:         d.Document,
		indentation: indentation + s.opt.nestSize,
	})
}

func (d group) fit(s *state, mode Mode, indentation int) {
	panic("TODO fits group")
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
	switch mode {
	case ModeBroken:
		s.builder.WriteByte('\n')
		// Restore indentation
		padding := strings.Repeat(string(s.opt.indentationSymbol), indentation)
		s.builder.WriteString(padding)

	case ModeUnbroken:
		s.builder.WriteString(d.Unbroken)

	default:
		utils.NonExhaustiveMatchPanic[any](mode)
	}
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
	stateCopy := state{
		opt:   s.opt,
		width: s.opt.maxWidth - s.width,
		queue: NewQueueOf(queueItem{
			doc:         d.Document,
			indentation: indentation,
			mode:        mode,
		}),

		// Do not pass the builder
		// this will result in a runtime error if accessed
		// however the fit() function is not supposed to write on the buffer
	}

	var nestedMode Mode
	if stateCopy.fits() {
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

func (s state) fits() bool {
	for !s.queue.IsEmpty() {
		if s.width < 0 {
			return false
		}

		item := s.queue.Pop()
		item.doc.fit(&s, item.mode, item.indentation)
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
