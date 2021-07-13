package toml

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/naoina/toml/ast"
)

const (
	tagOmitempty = "omitempty"
	tagSkip      = "-"
)

// Marshal returns the TOML encoding of v.
//
// Struct values encode as TOML. Each exported struct field becomes a field of
// the TOML structure unless
//   - the field's tag is "-", or
//   - the field is empty and its tag specifies the "omitempty" option.
//
// The "toml" key in the struct field's tag value is the key name, followed by
// an optional comma and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `toml:"-"`
//
//   // Field appears in TOML as key "myName".
//   Field int `toml:"myName"`
//
//   // Field appears in TOML as key "myName" and the field is omitted from the
//   // result of encoding if its value is empty.
//   Field int `toml:"myName,omitempty"`
//
//   // Field appears in TOML as key "field", but the field is skipped if
//   // empty. Note the leading comma.
//   Field int `toml:",omitempty"`
func (cfg *Config) Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := cfg.NewEncoder(buf).Encode(v)
	return buf.Bytes(), err
}

// A Encoder writes TOML to an output stream.
type Encoder struct {
	w   io.Writer
	cfg *Config
}

// NewEncoder returns a new Encoder that writes to w.
func (cfg *Config) NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w, cfg}
}

// Encode writes the TOML of v to the stream.
// See the documentation for Marshal for details about the conversion of Go values to TOML.
func (e *Encoder) Encode(v interface{}) error {
	var (
		buf = &tableBuf{typ: ast.TableTypeNormal}
		rv  = reflect.ValueOf(v)
		err error
	)

	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return &marshalNilError{rv.Type()}
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		_, err = buf.structFields(e.cfg, rv)
	case reflect.Map:
		_, err = buf.mapFields(e.cfg, rv)
	case reflect.Interface:
		return e.Encode(rv.Interface())
	default:
		err = &marshalTableError{rv.Type()}
	}
	if err != nil {
		return err
	}
	return buf.writeTo(e.w, "")
}

// Marshaler can be implemented to override the encoding of TOML values. The returned text
// must be a simple TOML value (i.e. not a table) and is inserted into marshaler output.
//
// This interface exists for backwards-compatibility reasons. You probably want to
// implement encoding.TextMarshaler or MarshalerRec instead.
type Marshaler interface {
	MarshalTOML() ([]byte, error)
}

// MarshalerRec can be implemented to override the TOML encoding of a type.
// The returned value is marshaled in place of the receiver.
type MarshalerRec interface {
	MarshalTOML() (interface{}, error)
}

type tableBuf struct {
	name string // already escaped / quoted
	typ  ast.TableType

	body     []byte      // text below table header
	children []*tableBuf // sub-tables of this table

	arrayDepth      int // if > 0 in value(x), x is contained in an array.
	mixedArrayDepth int // if > 0 in value(x), x is contained in a mixed array.
}

// writeTo writes b and all of its children to w.
func (b *tableBuf) writeTo(w io.Writer, prefix string) error {
	key := b.name // TODO: escape dots
	if prefix != "" {
		key = prefix + "." + key
	}

	if b.name != "" {
		head := "[" + key + "]"
		if b.typ == ast.TableTypeArray {
			head = "[" + head + "]"
		}
		head += "\n"
		if _, err := io.WriteString(w, head); err != nil {
			return err
		}
	}
	if _, err := w.Write(b.body); err != nil {
		return err
	}

	for i, child := range b.children {
		if len(b.body) > 0 || i > 0 {
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if err := child.writeTo(w, key); err != nil {
			return err
		}
	}
	return nil
}

// newChild creates a new child table of b.
func (b *tableBuf) newChild(name string) *tableBuf {
	child := &tableBuf{name: quoteName(name), typ: ast.TableTypeNormal}
	if b.arrayDepth > 0 {
		child.typ = ast.TableTypeArray
		// Note: arrayDepth does not inherit into child tables!
	}
	if b.mixedArrayDepth > 0 {
		child.typ = ast.TableTypeInline
		child.mixedArrayDepth = b.mixedArrayDepth
		b.body = append(b.body, '{')
	}
	return child
}

// addChild adds a child table to b.
// This is called after all values in child have already been
// written to child.body.
func (b *tableBuf) addChild(child *tableBuf) {
	// Inline tables are not tracked in b.children.
	if child.typ == ast.TableTypeInline {
		b.body = append(b.body, child.body...)
		b.body = append(b.body, '}')
		return
	}

	// Empty table elision: we can avoid writing a table that doesn't have any keys on its
	// own. Array tables can't be elided because they define array elements (which would
	// be missing if elided).
	if len(child.body) == 0 && child.typ == ast.TableTypeNormal {
		for _, gchild := range child.children {
			gchild.name = child.name + "." + gchild.name
			b.addChild(gchild)
		}
		return
	}
	b.children = append(b.children, child)
}

// structFields writes applicable fields of a struct.
func (b *tableBuf) structFields(cfg *Config, rv reflect.Value) (newTables []*tableBuf, err error) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		// Check if the field should be written at all.
		ft := rt.Field(i)
		if ft.PkgPath != "" && !ft.Anonymous { // not exported
			continue
		}
		name, rest := extractTag(ft.Tag.Get(fieldTagName))
		if name == tagSkip {
			continue
		}
		fv := rv.Field(i)
		if rest == tagOmitempty && isEmptyValue(fv) {
			continue
		}
		if name == "" {
			name = cfg.FieldToKey(rt, ft.Name)
		}

		// If the current table is inline, write separators.
		if b.typ == ast.TableTypeInline && i > 0 {
			b.body = append(b.body, ", "...)
		}
		// Write the key/value pair.
		tables, err := b.field(cfg, name, fv)
		if err != nil {
			return newTables, err
		}
		newTables = append(newTables, tables...)
	}
	return newTables, nil
}

// mapFields writes the content of a map.
func (b *tableBuf) mapFields(cfg *Config, rv reflect.Value) ([]*tableBuf, error) {
	// Marshal and sort the keys first.
	var keys = rv.MapKeys()
	var keylist = make(mapKeyList, len(keys))
	for i, key := range keys {
		var err error
		keylist[i].key, err = encodeMapKey(key)
		if err != nil {
			return nil, err
		}
		keylist[i].value = rv.MapIndex(key)
	}
	sort.Sort(keylist)

	var newTables []*tableBuf
	var index int
	for _, kv := range keylist {
		// If the current table is inline, add separators.
		if b.typ == ast.TableTypeInline && index > 0 {
			b.body = append(b.body, ", "...)
		}
		// Write the key/value pair.
		tables, err := b.field(cfg, kv.key, kv.value)
		if err != nil {
			return newTables, err
		}
		newTables = append(newTables, tables...)
		index++
	}
	return newTables, nil
}

// field writes a key/value pair.
func (b *tableBuf) field(cfg *Config, name string, rv reflect.Value) ([]*tableBuf, error) {
	off := len(b.body)
	b.body = append(b.body, quoteName(name)...)
	b.body = append(b.body, " = "...)
	tables, err := b.value(cfg, rv, name)
	switch {
	case b.typ == ast.TableTypeInline:
		// Inline tables don't have newlines.
	case len(tables) > 0:
		// Value was written as a new table, rub out "key =".
		b.body = b.body[:off]
	default:
		// Regular key/value pair in table.
		b.body = append(b.body, '\n')
	}
	return tables, err
}

// value writes a plain value.
func (b *tableBuf) value(cfg *Config, rv reflect.Value, name string) ([]*tableBuf, error) {
	isMarshaler, tables, err := b.marshaler(cfg, rv, name)
	if isMarshaler {
		return tables, err
	}

	k := rv.Kind()
	switch {
	case k >= reflect.Int && k <= reflect.Int64:
		b.body = strconv.AppendInt(b.body, rv.Int(), 10)
		return nil, nil

	case k >= reflect.Uint && k <= reflect.Uintptr:
		b.body = strconv.AppendUint(b.body, rv.Uint(), 10)
		return nil, nil

	case k >= reflect.Float32 && k <= reflect.Float64:
		b.body = appendFloat(b.body, rv.Float())
		return nil, nil

	case k == reflect.Bool:
		b.body = strconv.AppendBool(b.body, rv.Bool())
		return nil, nil

	case k == reflect.String:
		b.body = strconv.AppendQuote(b.body, rv.String())
		return nil, nil

	case k == reflect.Ptr || k == reflect.Interface:
		if rv.IsNil() {
			return nil, &marshalNilError{rv.Type()}
		}
		return b.value(cfg, rv.Elem(), name)

	case k == reflect.Slice || k == reflect.Array:
		return b.array(cfg, rv, name)

	case k == reflect.Struct:
		child := b.newChild(name)
		tables, err := child.structFields(cfg, rv)
		b.addChild(child)
		if child.typ == ast.TableTypeInline {
			return nil, err
		}
		tables = append(tables, child)
		return tables, err

	case k == reflect.Map:
		child := b.newChild(name)
		tables, err := child.mapFields(cfg, rv)
		b.addChild(child)
		if child.typ == ast.TableTypeInline {
			return nil, err
		}
		tables = append(tables, child)
		return tables, err

	default:
		return nil, fmt.Errorf("toml: marshal: unsupported type %v", rv.Kind())
	}
}

func (b *tableBuf) array(cfg *Config, rv reflect.Value, name string) ([]*tableBuf, error) {
	rvlen := rv.Len()
	if rvlen == 0 {
		b.body = append(b.body, '[', ']')
		return nil, nil
	}

	// If any parent value is a mixed array, this array must also be
	// written as a mixed array.
	if b.mixedArrayDepth > 0 {
		err := b.mixedArray(cfg, rv, name)
		return nil, err
	}

	// Bump depth. This ensures that any tables created while
	// encoding the array will become array tables.
	b.arrayDepth++
	defer func() { b.arrayDepth-- }()

	// Take a snapshot of the current state.
	var (
		childrenBeforeArray = b.children
		offsetBeforeArray   = len(b.body)
	)

	var (
		newTables     []*tableBuf
		anyPlainValue = false // true if any non-table was written.
	)
	b.body = append(b.body, '[')
	for i := 0; i < rvlen; i++ {
		if i > 0 {
			b.body = append(b.body, ", "...)
		}

		tables, err := b.value(cfg, rv.Index(i), name)
		if err != nil {
			return newTables, err
		}
		if len(tables) == 0 {
			anyPlainValue = true
		}
		newTables = append(newTables, tables...)

		if anyPlainValue && len(newTables) > 0 {
			// Turns out this is a heterogenous array, mixing table and non-table values.
			// If any tables were already created, we need to remove them again and start
			// over.
			b.children = childrenBeforeArray
			b.body = b.body[:offsetBeforeArray]
			err := b.mixedArray(cfg, rv, name)
			return nil, err
		}
	}

	if anyPlainValue {
		b.body = append(b.body, ']')
	} else {
		// The array contained only tables, rub out the initial '['
		// to reset the buffer.
		b.body = b.body[:offsetBeforeArray]
	}
	return newTables, nil
}

// mixedArray writes rv as an array of mixed table / non-table values.
// When this is called, we already know that rv is non-empty.
func (b *tableBuf) mixedArray(cfg *Config, rv reflect.Value, name string) error {
	// Ensure that any elements written as tables are written inline.
	b.mixedArrayDepth++
	defer func() { b.mixedArrayDepth-- }()

	b.body = append(b.body, '[')
	defer func() { b.body = append(b.body, ']') }()

	for i := 0; i < rv.Len(); i++ {
		if i > 0 {
			b.body = append(b.body, ", "...)
		}
		tables, err := b.value(cfg, rv.Index(i), name)
		if len(tables) > 0 {
			panic("toml: b.value created new tables in inline-table mode")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// marshaler writes a value that implements any of the marshaler interfaces.
func (b *tableBuf) marshaler(cfg *Config, rv reflect.Value, name string) (handled bool, newTables []*tableBuf, err error) {
	switch t := rv.Interface().(type) {
	case encoding.TextMarshaler:
		enc, err := t.MarshalText()
		if err != nil {
			return true, nil, err
		}
		b.body = encodeTextMarshaler(b.body, string(enc))
		return true, nil, nil
	case MarshalerRec:
		newval, err := t.MarshalTOML()
		if err != nil {
			return true, nil, err
		}
		newTables, err = b.value(cfg, reflect.ValueOf(newval), name)
		return true, newTables, err
	case Marshaler:
		enc, err := t.MarshalTOML()
		if err != nil {
			return true, nil, err
		}
		b.body = append(b.body, enc...)
		return true, nil, nil
	}
	return false, nil, nil
}

func encodeTextMarshaler(buf []byte, v string) []byte {
	// Emit the value without quotes if possible.
	if v == "true" || v == "false" {
		return append(buf, v...)
	} else if _, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return append(buf, v...)
	} else if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return append(buf, v...)
	} else if _, err := strconv.ParseUint(v, 10, 64); err == nil {
		return append(buf, v...)
	} else if _, err := strconv.ParseFloat(v, 64); err == nil {
		return append(buf, v...)
	}
	return strconv.AppendQuote(buf, v)
}

func encodeMapKey(rv reflect.Value) (string, error) {
	if rv.Kind() == reflect.String {
		return rv.String(), nil
	}
	if tm, ok := rv.Interface().(encoding.TextMarshaler); ok {
		b, err := tm.MarshalText()
		return string(b), err
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(rv.Uint(), 10), nil
	}
	return "", fmt.Errorf("toml: invalid map key type %v", rv.Type())
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		// encoding/json treats all arrays with non-zero length as non-empty. We check the
		// array content here because zero-length arrays are almost never used.
		len := v.Len()
		for i := 0; i < len; i++ {
			if !isEmptyValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func appendFloat(out []byte, v float64) []byte {
	if math.IsNaN(v) {
		return append(out, "nan"...)
	}
	if math.IsInf(v, -1) {
		return append(out, "-inf"...)
	}
	if math.IsInf(v, 1) {
		return append(out, "inf"...)
	}
	return strconv.AppendFloat(out, v, 'e', -1, 64)
}

func quoteName(s string) string {
	if len(s) == 0 {
		return strconv.Quote(s)
	}
	for _, r := range s {
		if r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '-' || r == '_' {
			continue
		}
		return strconv.Quote(s)
	}
	return s
}

type mapKeyList []struct {
	key   string
	value reflect.Value
}

func (l mapKeyList) Len() int           { return len(l) }
func (l mapKeyList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l mapKeyList) Less(i, j int) bool { return l[i].key < l[j].key }
