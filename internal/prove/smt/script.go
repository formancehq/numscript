package smt

import (
	"fmt"
	"strings"
)

// Script accumulates declarations, definitions and assertions in order. It
// also remembers every definition, so Eval can evaluate a term that refers to
// defined names.
type Script struct {
	lines   []string
	defined map[string]Term
	sorts   map[string]Sort
	order   []string
}

// Header is what every complete script starts with.
const Header = "(set-logic QF_LIA)\n(set-option :produce-models true)\n"

func NewScript() *Script {
	s := NewBareScript()
	s.lines = append(s.lines, strings.TrimSuffix(Header, "\n"))
	return s
}

// NewBareScript starts a script without the header, for a body whose
// declarations are emitted separately in front of it.
func NewBareScript() *Script {
	return &Script{
		defined: map[string]Term{},
		sorts:   map[string]Sort{},
	}
}

// Lines returns the commands emitted so far.
func (s *Script) Lines() []string {
	return s.lines
}

func (s *Script) Comment(format string, args ...any) {
	for _, line := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		s.lines = append(s.lines, "; "+line)
	}
}

func (s *Script) Blank() {
	s.lines = append(s.lines, "")
}

// DeclareConst declares a free symbol. Declaring a name twice is a bug.
func (s *Script) DeclareConst(name string, sort Sort) Term {
	s.bind(name, sort)
	s.lines = append(s.lines, fmt.Sprintf("(declare-const %s %s)", Quote(name), sort))
	return symbol{name, sort}
}

// Define names t. A literal or a symbol is returned as is, so an alias costs
// no definition and the printed script stays a DAG of real computations.
func (s *Script) Define(name string, t Term) Term {
	if IsLeaf(t) {
		return t
	}
	s.bind(name, t.Sort())
	s.defined[name] = t
	s.lines = append(s.lines, fmt.Sprintf("(define-fun %s () %s %s)", Quote(name), t.Sort(), Print(t)))
	return symbol{name, t.Sort()}
}

// DefineAlways is Define for names that must exist in the script (the
// outputs a caller will get-value), even when their term is a leaf.
func (s *Script) DefineAlways(name string, t Term) Term {
	if !IsLeaf(t) {
		return s.Define(name, t)
	}
	s.bind(name, t.Sort())
	s.defined[name] = t
	s.lines = append(s.lines, fmt.Sprintf("(define-fun %s () %s %s)", Quote(name), t.Sort(), Print(t)))
	return symbol{name, t.Sort()}
}

func (s *Script) bind(name string, sort Sort) {
	if _, dup := s.sorts[name]; dup {
		panic("smt.Script: name bound twice: " + name)
	}
	s.sorts[name] = sort
	s.order = append(s.order, name)
}

func (s *Script) Assert(t Term) {
	s.lines = append(s.lines, fmt.Sprintf("(assert %s)", Print(t)))
}

// Text is the script so far, one command per line.
func (s *Script) Text() string {
	return strings.Join(s.lines, "\n") + "\n"
}

// Definition returns the term bound to a defined name.
func (s *Script) Definition(name string) (Term, bool) {
	t, ok := s.defined[name]
	return t, ok
}

// SortOf returns the sort of a declared or defined name.
func (s *Script) SortOf(name string) (Sort, bool) {
	sort, ok := s.sorts[name]
	return sort, ok
}
