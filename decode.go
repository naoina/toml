package toml

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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

// Unmarshal parses the TOML data and sotres the result in the value pointed to by v.
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
func Unmarshal(data []byte, v interface{}) (err error) {
	d := &decodeState{p: &tomlParser{Buffer: string(data)}}
	d.init()
	if err := d.parse(); err != nil {
		return err
	}
	if err := d.unmarshal(d.p.toml.table, v); err != nil {
		return fmt.Errorf("toml: unmarshal: %v", err)
	}
	return nil
}

// Unmarshaler is the interface implemented by objects that can unmarshal a
// TOML description of themselves.
// The input can be assumed to be a valid encoding of a TOML value.
// UnmarshalJSON must copy the TOML data if it wishes to retain the data after
// returning.
type Unmarshaler interface {
	UnmarshalTOML([]byte) error
}

type decodeState struct {
	p *tomlParser
}

func (d *decodeState) init() {
	d.p.Init()
	d.p.toml.init()
}

func (d *decodeState) parse() error {
	if err := d.p.Parse(); err != nil {
		if err, ok := err.(*parseError); ok {
			return fmt.Errorf("toml: line %d: parse error", err.Line())
		}
		return err
	}
	return d.execute()
}

func (d *decodeState) execute() (err error) {
	defer func() {
		e := recover()
		if e != nil {
			cerr, ok := e.(convertError)
			if !ok {
				panic(e)
			}
			err = cerr.err
		}
	}()
	d.p.Execute()
	return nil
}

func (d *decodeState) unmarshal(t *table, v interface{}) (err error) {
	if v == nil {
		return fmt.Errorf("v must not be nil")
	}
	rv := reflect.ValueOf(v)
	if kind := rv.Kind(); kind != reflect.Ptr && kind != reflect.Map {
		return fmt.Errorf("v must be a pointer or map")
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	for key, val := range t.fieldMap {
		switch av := val.(type) {
		case *keyValue:
			fv, fieldName, found := findField(rv, key)
			if !found {
				return fmt.Errorf("line %d: field corresponding to `%s' is not defined in `%T'", av.line, key, v)
			}
			switch fv.Kind() {
			case reflect.Map:
				mv := reflect.New(fv.Type().Elem()).Elem()
				if err := d.unmarshal(t, mv.Addr().Interface()); err != nil {
					return err
				}
				fv.SetMapIndex(reflect.ValueOf(fieldName), mv)
			default:
				if err := d.setValue(fv, av.value); err != nil {
					return fmt.Errorf("line %d: %v.%s: %v", av.line, rv.Type(), fieldName, err)
				}
			}
		case *table:
			fv, fieldName, found := findField(rv, key)
			if !found {
				return fmt.Errorf("line %d: field corresponding to `%s' is not defined in `%T'", av.line, key, v)
			}
			if err, ok := d.setUnmarshaler(fv, string(d.p.buffer[av.begin:av.end])); ok {
				if err != nil {
					return err
				}
				continue
			}
			for fv.Kind() == reflect.Ptr {
				fv.Set(reflect.New(fv.Type().Elem()))
				fv = fv.Elem()
			}
			switch fv.Kind() {
			case reflect.Struct:
				vv := reflect.New(fv.Type()).Elem()
				if err := d.unmarshal(av, vv.Addr().Interface()); err != nil {
					return err
				}
				fv.Set(vv)
				if rv.Kind() == reflect.Map {
					rv.SetMapIndex(reflect.ValueOf(fieldName), fv)
				}
			case reflect.Map:
				mv := reflect.MakeMap(fv.Type())
				if err := d.unmarshal(av, mv.Interface()); err != nil {
					return err
				}
				fv.Set(mv)
			default:
				return fmt.Errorf("line %d: `%v.%s' must be struct or map, but %v given", av.line, rv.Type(), fieldName, fv.Kind())
			}
		case []*table:
			fv, fieldName, found := findField(rv, key)
			if !found {
				return fmt.Errorf("line %d: field corresponding to `%s' is not defined in `%T'", av[0].line, key, v)
			}
			data := make([]string, 0, len(av))
			for _, tbl := range av {
				data = append(data, string(d.p.buffer[tbl.begin:tbl.end]))
			}
			if err, ok := d.setUnmarshaler(fv, strings.Join(data, "\n")); ok {
				if err != nil {
					return err
				}
				continue
			}
			t := fv.Type().Elem()
			pc := 0
			for ; t.Kind() == reflect.Ptr; pc++ {
				t = t.Elem()
			}
			if fv.Kind() != reflect.Slice {
				return fmt.Errorf("line %d: `%v.%s' must be slice type, but %v given", av[0].line, rv.Type(), fieldName, fv.Kind())
			}
			for _, tbl := range av {
				var vv reflect.Value
				switch t.Kind() {
				case reflect.Map:
					vv = reflect.MakeMap(t)
					if err := d.unmarshal(tbl, vv.Interface()); err != nil {
						return err
					}
				default:
					vv = reflect.New(t).Elem()
					if err := d.unmarshal(tbl, vv.Addr().Interface()); err != nil {
						return err
					}
				}
				for i := 0; i < pc; i++ {
					vv = vv.Addr()
					pv := reflect.New(vv.Type()).Elem()
					pv.Set(vv)
					vv = pv
				}
				fv.Set(reflect.Append(fv, vv))
			}
			if rv.Kind() == reflect.Map {
				rv.SetMapIndex(reflect.ValueOf(fieldName), fv)
			}
		default:
			return fmt.Errorf("BUG: unknown type `%T'", t)
		}
	}
	return nil
}

func (d *decodeState) setUnmarshaler(lhs reflect.Value, data string) (error, bool) {
	for lhs.Kind() == reflect.Ptr {
		lhs.Set(reflect.New(lhs.Type().Elem()))
		lhs = lhs.Elem()
	}
	if lhs.CanAddr() {
		if u, ok := lhs.Addr().Interface().(Unmarshaler); ok {
			return u.UnmarshalTOML([]byte(data)), true
		}
	}
	return nil, false
}

func (d *decodeState) setValue(lhs reflect.Value, val ast.Value) error {
	for lhs.Kind() == reflect.Ptr {
		lhs.Set(reflect.New(lhs.Type().Elem()))
		lhs = lhs.Elem()
	}
	if err, ok := d.setUnmarshaler(lhs, string(d.p.buffer[val.Pos():val.End()])); ok {
		return err
	}
	switch v := val.(type) {
	case *ast.Integer:
		if err := d.setInt(lhs, v); err != nil {
			return err
		}
	case *ast.Float:
		if err := d.setFloat(lhs, v); err != nil {
			return err
		}
	case *ast.String:
		if err := d.setString(lhs, v); err != nil {
			return err
		}
	case *ast.Boolean:
		if err := d.setBoolean(lhs, v); err != nil {
			return err
		}
	case *ast.Datetime:
		if err := d.setDatetime(lhs, v); err != nil {
			return err
		}
	case *ast.Array:
		if err := d.setArray(lhs, v); err != nil {
			return err
		}
	}
	return nil
}

func (d *decodeState) setInt(fv reflect.Value, v *ast.Integer) error {
	i, err := strconv.ParseInt(v.Value, 10, 64)
	if err != nil {
		return err
	}
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fv.OverflowInt(i) {
			return &errorOutOfRange{fv.Kind(), i}
		}
		fv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fv.SetUint(uint64(i))
	case reflect.Interface:
		fv.Set(reflect.ValueOf(i))
	default:
		return fmt.Errorf("`%v' is not any types of int", fv.Type())
	}
	return nil
}

func (d *decodeState) setFloat(fv reflect.Value, v *ast.Float) error {
	f, err := strconv.ParseFloat(v.Value, 64)
	if err != nil {
		return err
	}
	switch fv.Kind() {
	case reflect.Float32, reflect.Float64:
		if fv.OverflowFloat(f) {
			return &errorOutOfRange{fv.Kind(), f}
		}
		fv.SetFloat(f)
	case reflect.Interface:
		fv.Set(reflect.ValueOf(f))
	default:
		return fmt.Errorf("`%v' is not float32 or float64", fv.Type())
	}
	return nil
}

func (d *decodeState) setString(fv reflect.Value, v *ast.String) error {
	return d.set(fv, v.Value)
}

func (d *decodeState) setBoolean(fv reflect.Value, v *ast.Boolean) error {
	b, err := strconv.ParseBool(v.Value)
	if err != nil {
		return err
	}
	return d.set(fv, b)
}

func (d *decodeState) setDatetime(fv reflect.Value, v *ast.Datetime) error {
	tm, err := time.Parse(time.RFC3339Nano, v.Value)
	if err != nil {
		return err
	}
	return d.set(fv, tm)
}

func (d *decodeState) setArray(fv reflect.Value, v *ast.Array) error {
	if len(v.Value) == 0 {
		return nil
	}
	typ := reflect.TypeOf(v.Value[0])
	for _, vv := range v.Value[1:] {
		if typ != reflect.TypeOf(vv) {
			return fmt.Errorf("array cannot contain multiple types")
		}
	}
	sliceType := fv.Type()
	if fv.Kind() == reflect.Interface {
		sliceType = reflect.SliceOf(sliceType)
	}
	slice := reflect.MakeSlice(sliceType, 0, len(v.Value))
	t := sliceType.Elem()
	for _, vv := range v.Value {
		tmp := reflect.New(t).Elem()
		if err := d.setValue(tmp, vv); err != nil {
			return err
		}
		slice = reflect.Append(slice, tmp)
	}
	fv.Set(slice)
	return nil
}

func (d *decodeState) set(fv reflect.Value, v interface{}) error {
	rhs := reflect.ValueOf(v)
	if !rhs.Type().AssignableTo(fv.Type()) {
		return fmt.Errorf("`%v' type is not assignable to `%v' type", rhs.Type(), fv.Type())
	}
	fv.Set(rhs)
	return nil
}

type stack struct {
	key   string
	table *table
}

type toml struct {
	table        *table
	line         int
	currentTable *table
	s            string
	key          string
	val          ast.Value
	arr          *array
	tableMap     map[string]*table
	stack        []*stack
	skip         bool
}

func (p *toml) init() {
	p.line = 1
	p.table = &table{line: p.line, tableType: tableTypeNormal}
	p.tableMap = map[string]*table{
		"": p.table,
	}
	p.currentTable = p.table
}

func (p *toml) Error(err error) {
	panic(convertError{fmt.Errorf("toml: line %d: %v", p.line, err)})
}

func (p *tomlParser) SetTime(begin, end int) {
	p.val = &ast.Datetime{
		Position: ast.Position{Begin: begin, End: end},
		Value:    string(p.buffer[begin:end]),
	}
}

func (p *tomlParser) SetFloat64(begin, end int) {
	p.val = &ast.Float{
		Position: ast.Position{Begin: begin, End: end},
		Value:    underscoreReplacer.Replace(string(p.buffer[begin:end])),
	}
}

func (p *tomlParser) SetInt64(begin, end int) {
	p.val = &ast.Integer{
		Position: ast.Position{Begin: begin, End: end},
		Value:    underscoreReplacer.Replace(string(p.buffer[begin:end])),
	}
}

func (p *tomlParser) SetString(begin, end int) {
	p.val = &ast.String{
		Position: ast.Position{Begin: begin, End: end},
		Value:    p.s,
	}
	p.s = ""
}

func (p *tomlParser) SetBool(begin, end int) {
	p.val = &ast.Boolean{
		Position: ast.Position{Begin: begin, End: end},
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
	p.val = p.arr.current
	p.arr = p.arr.parent
}

func (p *toml) SetTable(buf []rune, begin, end int) {
	p.setTable(p.table, buf, begin, end)
}

func (p *toml) setTable(t *table, buf []rune, begin, end int) {
	name := string(buf[begin:end])
	if t, exists := p.tableMap[name]; exists {
		p.Error(fmt.Errorf("table `%s' is in conflict with %v table in line %d", name, t.tableType, t.line))
	}
	t, err := p.lookupTable(t, splitTableKey(name))
	if err != nil {
		p.Error(err)
	}
	p.currentTable = t
	p.tableMap[name] = p.currentTable
}

func (p *toml) SetTableString(begin, end int) {
	p.currentTable.begin = begin
	p.currentTable.end = end
}

func (p *toml) SetArrayTable(buf []rune, begin, end int) {
	p.setArrayTable(p.table, buf, begin, end)
}

func (p *toml) setArrayTable(t *table, buf []rune, begin, end int) {
	name := string(buf[begin:end])
	if t, exists := p.tableMap[name]; exists && t.tableType == tableTypeNormal {
		p.Error(fmt.Errorf("table `%s' is in conflict with %v table in line %d", name, t.tableType, t.line))
	}
	names := splitTableKey(name)
	t, err := p.lookupTable(t, names[:len(names)-1])
	if err != nil {
		p.Error(err)
	}
	last := names[len(names)-1]
	tbl := &table{
		name:      last,
		line:      p.line,
		tableType: tableTypeArray,
	}
	switch v := t.fieldMap[last].(type) {
	case nil:
		if t.fieldMap == nil {
			t.fieldMap = make(map[string]interface{})
		}
		t.fieldMap[last] = []*table{tbl}
	case []*table:
		t.fieldMap[last] = append(v, tbl)
	case *keyValue:
		p.Error(fmt.Errorf("key `%s' is in conflict with line %d", last, v.line))
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
	if val, exists := p.currentTable.fieldMap[p.key]; exists {
		switch v := val.(type) {
		case *table:
			p.Error(fmt.Errorf("key `%s' is in conflict with %v table in line %d", p.key, v.tableType, v.line))
		case *keyValue:
			p.Error(fmt.Errorf("key `%s' is in conflict with line %d", p.key, v.line))
		default:
			p.Error(fmt.Errorf("BUG: key `%s' is in conflict but it's unknown type `%T'", p.key, v))
		}
	}
	if p.currentTable.fieldMap == nil {
		p.currentTable.fieldMap = make(map[string]interface{})
	}
	p.currentTable.fieldMap[p.key] = &keyValue{
		key:   p.key,
		value: p.val,
		line:  p.line,
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

func (p *toml) lookupTable(t *table, keys []string) (*table, error) {
	for _, s := range keys {
		val, exists := t.fieldMap[s]
		if !exists {
			tbl := &table{
				name:      s,
				line:      p.line,
				tableType: tableTypeNormal,
			}
			if t.fieldMap == nil {
				t.fieldMap = make(map[string]interface{})
			}
			t.fieldMap[s] = tbl
			t = tbl
			continue
		}
		switch v := val.(type) {
		case *table:
			t = v
		case []*table:
			t = v[len(v)-1]
		case *keyValue:
			return nil, fmt.Errorf("key `%s' is in conflict with line %d", s, v.line)
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

type convertError struct {
	err error
}

func (e convertError) Error() string {
	return e.err.Error()
}

type tableType uint8

const (
	tableTypeNormal tableType = iota
	tableTypeArray
)

var tableTypes = [...]string{
	"normal",
	"array",
}

func (t tableType) String() string {
	return tableTypes[t]
}

type table struct {
	name      string
	fieldMap  map[string]interface{}
	line      int
	tableType tableType
	begin     int
	end       int
}

type keyValue struct {
	key   string
	value ast.Value
	line  int
}

type array struct {
	parent  *array
	child   *array
	current *ast.Array
	line    int
}
