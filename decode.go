package toml

import (
	"encoding"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/naoina/toml/ast"
)

const (
	tableSeparator = '.'
)

var (
	escapeReplacer = strings.NewReplacer(
		"\b", "\\n",
		"\f", "\\f",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	underscoreReplacer = strings.NewReplacer(
		"_", "",
	)
)

// Unmarshal parses the TOML data and stores the result in the value pointed to by v.
//
// Unmarshal will mapped to v that according to following rules:
//
//	TOML strings to string
//	TOML integers to any int type
//	TOML floats to float32 or float64
//	TOML booleans to bool
//	TOML datetimes to time.Time
//	TOML arrays to any type of slice or []interface{}
//	TOML tables to struct
//	TOML array of tables to slice of struct
func Unmarshal(data []byte, v interface{}) error {
	table, err := Parse(data)
	if err != nil {
		return err
	}
	if err := UnmarshalTable(table, v); err != nil {
		return err
	}
	return nil
}

// A Decoder reads and decodes TOML from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new Decoder that reads from r.
// Note that it reads all from r before parsing it.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

// Decode parses the TOML data from its input and stores it in the value pointed to by v.
// See the documentation for Unmarshal for details about the conversion of TOML into a Go value.
func (d *Decoder) Decode(v interface{}) error {
	b, err := ioutil.ReadAll(d.r)
	if err != nil {
		return err
	}
	return Unmarshal(b, v)
}

// UnmarshalerRec may be implemented by types to customize their behavior when being
// unmarshaled from TOML. You can use it to implement custom validation or to set
// unexported fields.
//
// UnmarshalTOML receives a function that can be called to unmarshal the original TOML
// value into a field or variable. It is safe to call the function more than once if
// necessary.
type UnmarshalerRec interface {
	UnmarshalTOML(fn func(interface{}) error) error
}

// Unmarshaler can be used to capture and process raw TOML source of a table or value.
// UnmarshalTOML must copy the input if it wishes to retain it after returning.
//
// Note: this interface is retained for backwards compatibility. You probably want
// to implement encoding.TextUnmarshaler or UnmarshalerRec instead.
type Unmarshaler interface {
	UnmarshalTOML(input []byte) error
}

// UnmarshalTable applies the contents of an ast.Table to the value pointed at by v.
//
// UnmarshalTable will mapped to v that according to following rules:
//
//	TOML strings to string
//	TOML integers to any int type
//	TOML floats to float32 or float64
//	TOML booleans to bool
//	TOML datetimes to time.Time
//	TOML arrays to any type of slice or []interface{}
//	TOML tables to struct
//	TOML array of tables to slice of struct
func UnmarshalTable(t *ast.Table, v interface{}) error {
	rv := reflect.ValueOf(v)
	toplevelMap := rv.Kind() == reflect.Map
	if (!toplevelMap && rv.Kind() != reflect.Ptr) || rv.IsNil() {
		return &invalidUnmarshalError{reflect.TypeOf(v)}
	}
	return unmarshalTable(rv, t, toplevelMap)
}

// used for UnmarshalerRec.
func unmarshalTableOrValue(rv reflect.Value, av interface{}) error {
	if (rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Map) || rv.IsNil() {
		return &invalidUnmarshalError{rv.Type()}
	}
	rv = indirect(rv)

	switch av.(type) {
	case *ast.KeyValue, *ast.Table, []*ast.Table:
		if err := unmarshalField(rv, av); err != nil {
			return lineError(fieldLineNumber(av), err)
		}
		return nil
	case ast.Value:
		return setValue(rv, av.(ast.Value))
	default:
		panic(fmt.Sprintf("BUG: unhandled AST node type %T", av))
	}
}

// unmarshalTable unmarshals the fields of a table into a struct or map.
//
// toplevelMap is true when rv is an (unadressable) map given to UnmarshalTable. In this
// (special) case, the map is used as-is instead of creating a new map.
func unmarshalTable(rv reflect.Value, t *ast.Table, toplevelMap bool) error {
	rv = indirect(rv)
	if err, ok := setUnmarshaler(rv, t); ok {
		return lineError(t.Line, err)
	}
	switch rv.Kind() {
	case reflect.Struct:
		for key, fieldAst := range t.Fields {
			fv, fieldName, found := findField(rv, key)
			if !found {
				return lineError(fieldLineNumber(fieldAst), fmt.Errorf("field corresponding to '%s' is not defined in %v", key, rv.Type()))
			}
			if err := unmarshalField(fv, fieldAst); err != nil {
				return lineErrorField(fieldLineNumber(fieldAst), rv.Type().String()+"."+fieldName, err)
			}
		}
	case reflect.Map: // TODO: reflect.Interface
		m := rv
		if !toplevelMap {
			m = reflect.MakeMap(rv.Type())
		}
		elemtyp := rv.Type().Elem()
		for key, fieldAst := range t.Fields {
			kv, err := unmarshalMapKey(rv.Type().Key(), key)
			if err != nil {
				return lineError(fieldLineNumber(fieldAst), err)
			}
			fv := reflect.New(elemtyp).Elem()
			if err := unmarshalField(fv, fieldAst); err != nil {
				return lineError(fieldLineNumber(fieldAst), err)
			}
			m.SetMapIndex(kv, fv)
		}
		if !toplevelMap {
			rv.Set(m)
		}
	default:
		return lineError(t.Line, &unmarshalTypeError{"table", "struct or map", rv.Type()})
	}
	return nil
}

func fieldLineNumber(fieldAst interface{}) int {
	switch av := fieldAst.(type) {
	case *ast.KeyValue:
		return av.Line
	case *ast.Table:
		return av.Line
	case []*ast.Table:
		return av[0].Line
	default:
		panic(fmt.Sprintf("BUG: unhandled node type %T", fieldAst))
	}
}

func unmarshalField(rv reflect.Value, fieldAst interface{}) error {
	switch av := fieldAst.(type) {
	case *ast.KeyValue:
		return setValue(rv, av.Value)
	case *ast.Table:
		return unmarshalTable(rv, av, false)
	case []*ast.Table:
		rv = indirect(rv)
		if err, ok := setUnmarshaler(rv, fieldAst); ok {
			return err
		}
		if rv.Kind() != reflect.Slice {
			return &unmarshalTypeError{"array table", "slice", rv.Type()}
		}
		slice := reflect.MakeSlice(rv.Type(), len(av), len(av))
		for i, tbl := range av {
			vv := reflect.New(rv.Type().Elem()).Elem()
			if err := unmarshalTable(vv, tbl, false); err != nil {
				return err
			}
			slice.Index(i).Set(vv)
		}
		rv.Set(slice)
	default:
		panic(fmt.Sprintf("BUG: unhandled AST node type %T", av))
	}
	return nil
}

func unmarshalMapKey(typ reflect.Type, key string) (reflect.Value, error) {
	rv := reflect.New(typ).Elem()
	if u, ok := rv.Addr().Interface().(encoding.TextUnmarshaler); ok {
		return rv, u.UnmarshalText([]byte(key))
	}
	switch typ.Kind() {
	case reflect.String:
		rv.SetString(key)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(key, 10, int(typ.Size()*8))
		if err != nil {
			return rv, convertNumError(typ.Kind(), err)
		}
		rv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(key, 10, int(typ.Size()*8))
		if err != nil {
			return rv, convertNumError(typ.Kind(), err)
		}
		rv.SetUint(i)
	default:
		return rv, fmt.Errorf("invalid map key type %s", typ)
	}
	return rv, nil
}

func setValue(lhs reflect.Value, val ast.Value) error {
	lhs = indirect(lhs)
	if err, ok := setUnmarshaler(lhs, val); ok {
		return err
	}
	if err, ok := setTextUnmarshaler(lhs, val); ok {
		return err
	}
	switch v := val.(type) {
	case *ast.Integer:
		return setInt(lhs, v)
	case *ast.Float:
		return setFloat(lhs, v)
	case *ast.String:
		return setString(lhs, v)
	case *ast.Boolean:
		return setBoolean(lhs, v)
	case *ast.Array:
		return setArray(lhs, v)
	case *ast.Datetime:
		panic("*ast.Datetime should be handled by TextUnmarshaler")
	default:
		panic(fmt.Sprintf("BUG: unhandled node type %T", v))
	}
}

func indirect(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	return rv
}

func setUnmarshaler(lhs reflect.Value, av interface{}) (error, bool) {
	if lhs.CanAddr() {
		if u, ok := lhs.Addr().Interface().(UnmarshalerRec); ok {
			err := u.UnmarshalTOML(func(v interface{}) error {
				return unmarshalTableOrValue(reflect.ValueOf(v), av)
			})
			return err, true
		}
		if u, ok := lhs.Addr().Interface().(Unmarshaler); ok {
			return u.UnmarshalTOML(unmarshalerSource(av)), true
		}
	}
	return nil, false
}

func unmarshalerSource(av interface{}) []byte {
	var source []byte
	switch av := av.(type) {
	case []*ast.Table:
		for i, tab := range av {
			source = append(source, tab.Source()...)
			if i != len(av)-1 {
				source = append(source, '\n')
			}
		}
	case ast.Value:
		source = []byte(av.Source())
	default:
		panic(fmt.Sprintf("BUG: unhandled node type %T", av))
	}
	return source
}

func setTextUnmarshaler(lhs reflect.Value, val ast.Value) (error, bool) {
	if !lhs.CanAddr() {
		return nil, false
	}
	u, ok := lhs.Addr().Interface().(encoding.TextUnmarshaler)
	if !ok {
		return nil, false
	}
	var data string
	switch val := val.(type) {
	case *ast.Array:
		return &unmarshalTypeError{"array", "", lhs.Type()}, true
	case *ast.String:
		data = val.Value
	default:
		data = val.Source()
	}
	return u.UnmarshalText([]byte(data)), true
}

func setInt(fv reflect.Value, v *ast.Integer) error {
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(v.Value, 10, int(fv.Type().Size()*8))
		if err != nil {
			return convertNumError(fv.Kind(), err)
		}
		fv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(v.Value, 10, int(fv.Type().Size()*8))
		if err != nil {
			return convertNumError(fv.Kind(), err)
		}
		fv.SetUint(i)
	case reflect.Interface:
		i, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return convertNumError(reflect.Int64, err)
		}
		fv.Set(reflect.ValueOf(i))
	default:
		return &unmarshalTypeError{"integer", "", fv.Type()}
	}
	return nil
}

func setFloat(fv reflect.Value, v *ast.Float) error {
	f, err := v.Float()
	if err != nil {
		return err
	}
	switch fv.Kind() {
	case reflect.Float32, reflect.Float64:
		if fv.OverflowFloat(f) {
			return &overflowError{fv.Kind(), v.Value}
		}
		fv.SetFloat(f)
	case reflect.Interface:
		fv.Set(reflect.ValueOf(f))
	default:
		return &unmarshalTypeError{"float", "", fv.Type()}
	}
	return nil
}

func setString(fv reflect.Value, v *ast.String) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(v.Value)
		return nil
	case reflect.Interface:
		// TODO: need non-empty interface check here
		fv.Set(reflect.ValueOf(v.Value))
		return nil
	default:
		return &unmarshalTypeError{"string", "", fv.Type()}
	}
}

func setBoolean(fv reflect.Value, v *ast.Boolean) error {
	b, _ := v.Boolean()
	switch fv.Kind() {
	case reflect.Bool:
		fv.SetBool(b)
		return nil
	case reflect.Interface:
		// TODO: need non-empty interface check here
		fv.Set(reflect.ValueOf(b))
		return nil
	default:
		return &unmarshalTypeError{"boolean", "", fv.Type()}
	}
}

func setArray(rv reflect.Value, v *ast.Array) error {
	if len(v.Value) == 0 {
		return nil
	}
	tomltyp := reflect.TypeOf(v.Value[0])
	for _, vv := range v.Value[1:] {
		if tomltyp != reflect.TypeOf(vv) {
			return errArrayMultiType
		}
	}

	var slicetyp reflect.Type
	switch rv.Kind() {
	case reflect.Slice:
		slicetyp = rv.Type()
	case reflect.Interface:
		// TODO: need non-empty interface check here
		slicetyp = reflect.SliceOf(rv.Type())
	default:
		return &unmarshalTypeError{"array", "slice", rv.Type()}
	}

	slice := reflect.MakeSlice(slicetyp, len(v.Value), len(v.Value))
	typ := slicetyp.Elem()
	for i, vv := range v.Value {
		tmp := reflect.New(typ).Elem()
		if err := setValue(tmp, vv); err != nil {
			return err
		}
		slice.Index(i).Set(tmp)
	}
	rv.Set(slice)
	return nil
}

type stack struct {
	key   string
	table *ast.Table
}

type toml struct {
	table        *ast.Table
	line         int
	currentTable *ast.Table
	s            string
	key          string
	val          ast.Value
	arr          *array
	tableMap     map[string]*ast.Table
	stack        []*stack
	skip         bool
}

func (p *toml) init(data []rune) {
	p.line = 1
	p.table = &ast.Table{
		Line: p.line,
		Type: ast.TableTypeNormal,
		Data: data[:len(data)-1], // truncate the end_symbol added by PEG parse generator.
	}
	p.tableMap = map[string]*ast.Table{
		"": p.table,
	}
	p.currentTable = p.table
}

func (p *toml) Error(err error) {
	panic(lineError(p.line, err))
}

func (p *tomlParser) SetTime(begin, end int) {
	p.val = &ast.Datetime{
		Position: ast.Position{Begin: begin, End: end},
		Data:     p.buffer[begin:end],
		Value:    string(p.buffer[begin:end]),
	}
}

func (p *tomlParser) SetFloat64(begin, end int) {
	p.val = &ast.Float{
		Position: ast.Position{Begin: begin, End: end},
		Data:     p.buffer[begin:end],
		Value:    underscoreReplacer.Replace(string(p.buffer[begin:end])),
	}
}

func (p *tomlParser) SetInt64(begin, end int) {
	p.val = &ast.Integer{
		Position: ast.Position{Begin: begin, End: end},
		Data:     p.buffer[begin:end],
		Value:    underscoreReplacer.Replace(string(p.buffer[begin:end])),
	}
}

func (p *tomlParser) SetString(begin, end int) {
	p.val = &ast.String{
		Position: ast.Position{Begin: begin, End: end},
		Data:     p.buffer[begin:end],
		Value:    p.s,
	}
	p.s = ""
}

func (p *tomlParser) SetBool(begin, end int) {
	p.val = &ast.Boolean{
		Position: ast.Position{Begin: begin, End: end},
		Data:     p.buffer[begin:end],
		Value:    string(p.buffer[begin:end]),
	}
}

func (p *tomlParser) StartArray() {
	if p.arr == nil {
		p.arr = &array{line: p.line, current: &ast.Array{}}
		return
	}
	p.arr.child = &array{parent: p.arr, line: p.line, current: &ast.Array{}}
	p.arr = p.arr.child
}

func (p *tomlParser) AddArrayVal() {
	if p.arr.current == nil {
		p.arr.current = &ast.Array{}
	}
	p.arr.current.Value = append(p.arr.current.Value, p.val)
}

func (p *tomlParser) SetArray(begin, end int) {
	p.arr.current.Position = ast.Position{Begin: begin, End: end}
	p.arr.current.Data = p.buffer[begin:end]
	p.val = p.arr.current
	p.arr = p.arr.parent
}

func (p *toml) SetTable(buf []rune, begin, end int) {
	p.setTable(p.table, buf, begin, end)
}

func (p *toml) setTable(t *ast.Table, buf []rune, begin, end int) {
	name := string(buf[begin:end])
	names := splitTableKey(name)
	if t, exists := p.tableMap[name]; exists {
		if lt := p.tableMap[names[len(names)-1]]; t.Type == ast.TableTypeArray || lt != nil && lt.Type == ast.TableTypeNormal {
			p.Error(fmt.Errorf("table `%s' is in conflict with %v table in line %d", name, t.Type, t.Line))
		}
	}
	t, err := p.lookupTable(t, names)
	if err != nil {
		p.Error(err)
	}
	p.currentTable = t
	p.tableMap[name] = p.currentTable
}

func (p *tomlParser) SetTableString(begin, end int) {
	p.currentTable.Data = p.buffer[begin:end]

	p.currentTable.Position.Begin = begin
	p.currentTable.Position.End = end
}

func (p *toml) SetArrayTable(buf []rune, begin, end int) {
	p.setArrayTable(p.table, buf, begin, end)
}

func (p *toml) setArrayTable(t *ast.Table, buf []rune, begin, end int) {
	name := string(buf[begin:end])
	if t, exists := p.tableMap[name]; exists && t.Type == ast.TableTypeNormal {
		p.Error(fmt.Errorf("table `%s' is in conflict with %v table in line %d", name, t.Type, t.Line))
	}
	names := splitTableKey(name)
	t, err := p.lookupTable(t, names[:len(names)-1])
	if err != nil {
		p.Error(err)
	}
	last := names[len(names)-1]
	tbl := &ast.Table{
		Position: ast.Position{begin, end},
		Line:     p.line,
		Name:     last,
		Type:     ast.TableTypeArray,
	}
	switch v := t.Fields[last].(type) {
	case nil:
		if t.Fields == nil {
			t.Fields = make(map[string]interface{})
		}
		t.Fields[last] = []*ast.Table{tbl}
	case []*ast.Table:
		t.Fields[last] = append(v, tbl)
	case *ast.KeyValue:
		p.Error(fmt.Errorf("key `%s' is in conflict with line %d", last, v.Line))
	default:
		p.Error(fmt.Errorf("BUG: key `%s' is in conflict but it's unknown type `%T'", last, v))
	}
	p.currentTable = tbl
	p.tableMap[name] = p.currentTable
}

func (p *toml) StartInlineTable() {
	p.skip = false
	p.stack = append(p.stack, &stack{p.key, p.currentTable})
	buf := []rune(p.key)
	if p.arr == nil {
		p.setTable(p.currentTable, buf, 0, len(buf))
	} else {
		p.setArrayTable(p.currentTable, buf, 0, len(buf))
	}
}

func (p *toml) EndInlineTable() {
	st := p.stack[len(p.stack)-1]
	p.key, p.currentTable = st.key, st.table
	p.stack[len(p.stack)-1] = nil
	p.stack = p.stack[:len(p.stack)-1]
	p.skip = true
}

func (p *toml) AddLineCount(i int) {
	p.line += i
}

func (p *toml) SetKey(buf []rune, begin, end int) {
	p.key = string(buf[begin:end])
}

func (p *toml) AddKeyValue() {
	if p.skip {
		p.skip = false
		return
	}
	if val, exists := p.currentTable.Fields[p.key]; exists {
		switch v := val.(type) {
		case *ast.Table:
			p.Error(fmt.Errorf("key `%s' is in conflict with %v table in line %d", p.key, v.Type, v.Line))
		case *ast.KeyValue:
			p.Error(fmt.Errorf("key `%s' is in conflict with line %d", p.key, v.Line))
		default:
			p.Error(fmt.Errorf("BUG: key `%s' is in conflict but it's unknown type `%T'", p.key, v))
		}
	}
	if p.currentTable.Fields == nil {
		p.currentTable.Fields = make(map[string]interface{})
	}
	p.currentTable.Fields[p.key] = &ast.KeyValue{
		Key:   p.key,
		Value: p.val,
		Line:  p.line,
	}
}

func (p *toml) SetBasicString(buf []rune, begin, end int) {
	p.s = p.unquote(string(buf[begin:end]))
}

func (p *toml) SetMultilineString() {
	p.s = p.unquote(`"` + escapeReplacer.Replace(strings.TrimLeft(p.s, "\r\n")) + `"`)
}

func (p *toml) AddMultilineBasicBody(buf []rune, begin, end int) {
	p.s += string(buf[begin:end])
}

func (p *toml) SetLiteralString(buf []rune, begin, end int) {
	p.s = string(buf[begin:end])
}

func (p *toml) SetMultilineLiteralString(buf []rune, begin, end int) {
	p.s = strings.TrimLeft(string(buf[begin:end]), "\r\n")
}

func (p *toml) unquote(s string) string {
	s, err := strconv.Unquote(s)
	if err != nil {
		p.Error(err)
	}
	return s
}

func (p *toml) lookupTable(t *ast.Table, keys []string) (*ast.Table, error) {
	for _, s := range keys {
		val, exists := t.Fields[s]
		if !exists {
			tbl := &ast.Table{
				Line: p.line,
				Name: s,
				Type: ast.TableTypeNormal,
			}
			if t.Fields == nil {
				t.Fields = make(map[string]interface{})
			}
			t.Fields[s] = tbl
			t = tbl
			continue
		}
		switch v := val.(type) {
		case *ast.Table:
			t = v
		case []*ast.Table:
			t = v[len(v)-1]
		case *ast.KeyValue:
			return nil, fmt.Errorf("key `%s' is in conflict with line %d", s, v.Line)
		default:
			return nil, fmt.Errorf("BUG: key `%s' is in conflict but it's unknown type `%T'", s, v)
		}
	}
	return t, nil
}

func splitTableKey(tk string) []string {
	key := make([]byte, 0, 1)
	keys := make([]string, 0, 1)
	inQuote := false
	for i := 0; i < len(tk); i++ {
		k := tk[i]
		switch {
		case k == tableSeparator && !inQuote:
			keys = append(keys, string(key))
			key = key[:0] // reuse buffer.
		case k == '"':
			inQuote = !inQuote
		case (k == ' ' || k == '\t') && !inQuote:
			// skip.
		default:
			key = append(key, k)
		}
	}
	keys = append(keys, string(key))
	return keys
}

type array struct {
	parent  *array
	child   *array
	current *ast.Array
	line    int
}
