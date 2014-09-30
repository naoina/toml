package ast

type Position struct {
	Begin int
	End   int
}

type Value interface {
	Pos() int
	End() int
}

type String struct {
	Position Position
	Value    string
}

func (s *String) Pos() int {
	return s.Position.Begin
}

func (s *String) End() int {
	return s.Position.End
}

type Integer struct {
	Position Position
	Value    string
}

func (i *Integer) Pos() int {
	return i.Position.Begin
}

func (i *Integer) End() int {
	return i.Position.End
}

type Float struct {
	Position Position
	Value    string
}

func (f *Float) Pos() int {
	return f.Position.Begin
}

func (f *Float) End() int {
	return f.Position.End
}

type Boolean struct {
	Position Position
	Value    string
}

func (b *Boolean) Pos() int {
	return b.Position.Begin
}

func (b *Boolean) End() int {
	return b.Position.End
}

type Datetime struct {
	Position Position
	Value    string
}

func (d *Datetime) Pos() int {
	return d.Position.Begin
}

func (d *Datetime) End() int {
	return d.Position.End
}

type Array struct {
	Position Position
	Value    []Value
}

func (a *Array) Pos() int {
	return a.Position.Begin
}

func (a *Array) End() int {
	return a.Position.End
}
