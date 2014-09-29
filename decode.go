package toml

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
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
				return fmt.Errorf("line %d: `%T.%s': %v", v.line, t, fieldName, err)
			}
		case *table:
			if fv.Kind() != reflect.Struct {
				return fmt.Errorf("line %d: `%T.%s' must be struct type, but `%T' given", v.line, t, fieldName, v)
			}
			vv := reflect.New(fv.Type())
			if err := d.unmarshal(v, vv.Interface()); err != nil {
				return err
			}
			fv.Set(vv.Elem())
		case []*table:
			if fv.Kind() != reflect.Slice {
				return fmt.Errorf("`line %d: %T.%s' must be slice type, but `%T' given", v[0].line, t, fieldName, v)
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

func (d *decodeState) setValue(fv reflect.Value, v interface{}) error {
	switch lhs, rhs := fv, reflect.ValueOf(v); rhs.Kind() {
	case reflect.Int64:
		if err := d.setInt(lhs, rhs.Interface().(int64)); err != nil {
			return err
		}
	case reflect.Float64:
		if err := d.setFloat(lhs, rhs.Interface().(float64)); err != nil {
			return err
		}
	case reflect.Slice: // array type in toml.
		sliceType := lhs.Type()
		if lhs.Kind() == reflect.Interface {
			sliceType = reflect.SliceOf(sliceType)
		}
		slice := reflect.MakeSlice(sliceType, 0, rhs.Len())
		t := sliceType.Elem()
		for i := 0; i < rhs.Len(); i++ {
			v := reflect.New(t).Elem()
			if err := d.setValue(v, rhs.Index(i).Interface()); err != nil {
				return err
			}
			slice = reflect.Append(slice, v)
		}
		lhs.Set(slice)
	case reflect.Invalid:
		// ignore.
	default:
		if !rhs.Type().AssignableTo(lhs.Type()) {
			return fmt.Errorf("`%v' type is not assignable to `%v' type", rhs.Type(), lhs.Type())
		}
		lhs.Set(rhs)
	}
	return nil
}

func (d *decodeState) setInt(fv reflect.Value, v int64) error {
	switch fv.Kind() {
	case reflect.Int:
		if !inRange(v, int64(minInt), int64(maxInt)) {
			return &errorOutOfRange{fv.Kind(), v}
		}
		fv.SetInt(v)
	case reflect.Int8:
		if !inRange(v, math.MinInt8, math.MaxInt8) {
			return &errorOutOfRange{fv.Kind(), v}
		}
		fv.SetInt(v)
	case reflect.Int16:
		if !inRange(v, math.MinInt16, math.MaxInt16) {
			return &errorOutOfRange{fv.Kind(), v}
		}
		fv.SetInt(v)
	case reflect.Int32:
		if !inRange(v, math.MinInt32, math.MaxInt32) {
			return &errorOutOfRange{fv.Kind(), v}
		}
		fv.SetInt(v)
	case reflect.Int64:
		if !inRange(v, math.MinInt64, math.MaxInt64) {
			return &errorOutOfRange{fv.Kind(), v}
		}
		fv.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fv.SetUint(uint64(v))
	case reflect.Interface:
		fv.Set(reflect.ValueOf(v))
	default:
		return fmt.Errorf("`%v' is not any types of int", fv.Type())
	}
	return nil
}

func (d *decodeState) setFloat(fv reflect.Value, v float64) error {
	switch fv.Kind() {
	case reflect.Float32:
		if v < math.SmallestNonzeroFloat32 || math.MaxFloat32 < v {
			return &errorOutOfRange{fv.Kind(), v}
		}
		fv.SetFloat(v)
	case reflect.Float64:
		fv.SetFloat(v)
	case reflect.Interface:
		fv.Set(reflect.ValueOf(v))
	default:
		return fmt.Errorf("`%v' is not float32 or float64", fv.Type())
	}
	return nil
}

type toml struct {
	table        *table
	line         int
	currentTable *table
	s            string
	key          string
	val          interface{}
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

func (p *toml) SetTime(s string) {
	tm, err := time.Parse(`2006-01-02T15:04:05Z`, s)
	if err != nil {
		p.Error(err)
	}
	p.val = tm
}

func (p *toml) SetFloat64(s string) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		p.Error(err)
	}
	p.val = f
}

func (p *toml) SetInt64(s string) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		p.Error(err)
	}
	p.val = i
}

func (p *toml) SetString() {
	p.val = p.s
	p.s = ""
}

func (p *toml) SetBool(s string) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		p.Error(err)
	}
	p.val = b
}

func (p *toml) StartArray() {
	if p.arr == nil {
		p.arr = &array{line: p.line}
		return
	}
	p.arr.child = &array{parent: p.arr, line: p.line}
	p.arr = p.arr.child
}

func (p *toml) AddArrayVal() {
	rv := reflect.ValueOf(p.val)
	if p.arr.current == nil {
		p.arr.current = reflect.MakeSlice(reflect.SliceOf(rv.Type()), 0, 1).Interface()
	}
	if rv.Kind() == reflect.Slice {
		arrv := reflect.ValueOf(p.arr.current)
		if arrv.Type().Elem() != rv.Type() {
			slice := reflect.MakeSlice(reflect.TypeOf([]interface{}(nil)), 0, arrv.Len())
			for i := 0; i < arrv.Len(); i++ {
				slice = reflect.Append(slice, arrv.Index(i))
			}
			arrv = slice
		}
		p.arr.current = reflect.Append(arrv, rv).Interface()
		return
	}
	if reflect.TypeOf(p.arr.current) != reflect.SliceOf(rv.Type()) {
		p.ErrorArrayMultipleTypes()
	}
	p.arr.current = reflect.Append(reflect.ValueOf(p.arr.current), rv).Interface()
}

func (p *toml) EndArray() {
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
	p.currentTable.fieldMap[p.key] = &keyValue{key: p.key, value: p.val, line: p.line}
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

func (p *toml) ErrorArrayMultipleTypes() {
	p.Error(fmt.Errorf("array cannot contain multiple types"))
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
	value interface{}
	line  int
}

type array struct {
	parent  *array
	child   *array
	current interface{}
	line    int
}
