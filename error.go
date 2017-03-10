package toml

import (
	"fmt"
	"reflect"
)

func (e *parseError) Line() int {
	tokens := []token32{e.max}
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	for _, t := range translatePositions(e.p.buffer, positions) {
		if e.p.line < t.line {
			e.p.line = t.line
		}
	}
	return e.p.line
}

type errorOutOfRange struct {
	kind reflect.Kind
	v    interface{}
}

func (err *errorOutOfRange) Error() string {
	return fmt.Sprintf("value %d is out of range for `%v` type", err.v, err.kind)
}

type marshalNilError struct {
	typ reflect.Type
}

func (err *marshalNilError) Error() string {
	return fmt.Sprintf("toml: cannot marshal nil %s", err.typ)
}

type marshalTableError struct {
	typ reflect.Type
}

func (err *marshalTableError) Error() string {
	return fmt.Sprintf("toml: cannot marshal %s as table, want struct or map type", err.typ)
}
