package toml

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/naoina/toml/ast"
)

const (
	tableSeparator = "."
)

var (
	escapeReplacer = strings.NewReplacer(
		"\b", "\\n",
		"\f", "\\f",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	whitespaceReplacer = strings.NewReplacer(
		" ", "",
		"\t", "",
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
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	for name, val := range t.fieldMap {
		fv, fieldName, found := findField(rv, name)
		if !found {
			return fmt.Errorf("field corresponding to `%s' is not defined in `%T'", name, v)
		}
		switch v := val.(type) {
		case *keyValue:
			if err := d.setValue(fv, v.value); err != nil {
				return fmt.Errorf("line %d: %v.%s: %v", v.line, rv.Type(), fieldName, err)
			}
		case *table:
			if fv.Kind() != reflect.Struct {
				return fmt.Errorf("line %d: `%v.%s' must be struct type, but `%v' given", v.line, rv.Type(), fieldName, fv.Type())
			}
			vv := reflect.New(fv.Type())
			if err := d.unmarshal(v, vv.Interface()); err != nil {
				return err
			}
			fv.Set(vv.Elem())
		case []*table:
			if fv.Kind() != reflect.Slice {
				return fmt.Errorf("line %d: `%v.%s' must be slice type, but `%v' given", v[0].line, rv.Type(), fieldName, fv.Type())
			}
			for _, tbl := range v {
				vv := reflect.New(fv.Type().Elem())
				if err := d.unmarshal(tbl, vv.Interface()); err != nil {
					return err
				}
				fv.Set(reflect.Append(fv, vv.Elem()))
			}
		default:
			return fmt.Errorf("BUG: unknown type `%T'", t)
		}
	}
	return nil
}

func (d *decodeState) setValue(lhs reflect.Value, val ast.Value) error {
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
	case reflect.Int:
		if !inRange(i, int64(minInt), int64(maxInt)) {
			return &errorOutOfRange{fv.Kind(), i}
		}
		fv.SetInt(i)
	case reflect.Int8:
		if !inRange(i, math.MinInt8, math.MaxInt8) {
			return &errorOutOfRange{fv.Kind(), i}
		}
		fv.SetInt(i)
	case reflect.Int16:
		if !inRange(i, math.MinInt16, math.MaxInt16) {
			return &errorOutOfRange{fv.Kind(), i}
		}
		fv.SetInt(i)
	case reflect.Int32:
		if !inRange(i, math.MinInt32, math.MaxInt32) {
			return &errorOutOfRange{fv.Kind(), i}
		}
		fv.SetInt(i)
	case reflect.Int64:
		if !inRange(i, math.MinInt64, math.MaxInt64) {
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
	case reflect.Float32:
		if f < math.SmallestNonzeroFloat32 || math.MaxFloat32 < f {
			return &errorOutOfRange{fv.Kind(), f}
		}
		fv.SetFloat(f)
	case reflect.Float64:
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
	tm, err := time.Parse("2006-01-02T15:04:05Z", v.Value)
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

type toml struct {
	table        *table
	line         int
	currentTable *table
	s            string
	key          string
	val          ast.Value
	arr          *array
	tableMap     map[string]*table
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
		Value:    p.Buffer[begin:end],
	}
}

func (p *tomlParser) SetFloat64(begin, end int) {
	p.val = &ast.Float{
		Position: ast.Position{Begin: begin, End: end},
		Value:    p.Buffer[begin:end],
	}
}

func (p *tomlParser) SetInt64(begin, end int) {
	p.val = &ast.Integer{
		Position: ast.Position{Begin: begin, End: end},
		Value:    p.Buffer[begin:end],
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
		Value:    p.Buffer[begin:end],
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

func (p *toml) SetTable(name string) {
	name = whitespaceReplacer.Replace(name)
	if t, exists := p.tableMap[name]; exists {
		p.Error(fmt.Errorf("table `%s' is in conflict with %v table in line %d", name, t.tableType, t.line))
	}
	t, err := p.lookupTable(strings.Split(name, tableSeparator))
	if err != nil {
		p.Error(err)
	}
	p.currentTable = t
	p.tableMap[name] = p.currentTable
}

func (p *toml) SetArrayTable(name string) {
	name = whitespaceReplacer.Replace(name)
	if t, exists := p.tableMap[name]; exists && t.tableType == tableTypeNormal {
		p.Error(fmt.Errorf("table `%s' is in conflict with %v table in line %d", name, t.tableType, t.line))
	}
	names := strings.Split(name, tableSeparator)
	t, err := p.lookupTable(names[:len(names)-1])
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

func (p *toml) AddLineCount(i int) {
	p.line += i
}

func (p *toml) SetKey(key string) {
	p.key = key
}

func (p *toml) AddKeyValue() {
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

func (p *toml) SetBasicString(s string) {
	p.s = p.unquote(s)
}

func (p *toml) SetMultilineString() {
	p.s = p.unquote(`"` + escapeReplacer.Replace(strings.TrimLeft(p.s, "\r\n")) + `"`)
}

func (p *toml) AddMultilineBasicBody(s string) {
	p.s += s
}

func (p *toml) SetLiteralString(s string) {
	p.s = s
}

func (p *toml) SetMultilineLiteralString(s string) {
	p.s = strings.TrimLeft(s, "\r\n")
}

func (p *toml) RuneSlice(buf string, begin, end int) string {
	return string([]rune(buf)[begin:end])
}

func (p *toml) unquote(s string) string {
	s, err := strconv.Unquote(s)
	if err != nil {
		p.Error(err)
	}
	return s
}

func (p *toml) lookupTable(keys []string) (*table, error) {
	t := p.table
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
