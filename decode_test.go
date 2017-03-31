package toml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

func loadTestData(file string) []byte {
	f := filepath.Join("testdata", file)
	data, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return data
}

func mustTime(tm time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return tm
}

type Name struct {
	First string
	Last  string
}
type Inline struct {
	Name  Name
	Point map[string]int
}
type Subtable struct {
	Key string
}
type Table struct {
	Key      string
	Subtable Subtable
	Inline   Inline
}
type W struct {
}
type Z struct {
	W W
}
type Y struct {
	Z Z
}
type X struct {
	Y Y
}
type Basic struct {
	Basic string
}
type Continued struct {
	Key1 string
	Key2 string
	Key3 string
}
type Multiline struct {
	Key1      string
	Key2      string
	Key3      string
	Continued Continued
}
type LiteralMultiline struct {
	Regex2 string
	Lines  string
}
type Literal struct {
	Winpath   string
	Winpath2  string
	Quoted    string
	Regex     string
	Multiline LiteralMultiline
}
type String struct {
	Basic     Basic
	Multiline Multiline
	Literal   Literal
}
type IntegerUnderscores struct {
	Key1 int
	Key2 int
	Key3 int
}
type Integer struct {
	Key1        int
	Key2        int
	Key3        int
	Key4        int
	Underscores IntegerUnderscores
}
type Fractional struct {
	Key1 float64
	Key2 float64
	Key3 float64
}
type Exponent struct {
	Key1 float64
	Key2 float64
	Key3 float64
}
type Both struct {
	Key float64
}
type FloatUnderscores struct {
	Key1 float64
	Key2 float64
}
type Float struct {
	Fractional  Fractional
	Exponent    Exponent
	Both        Both
	Underscores FloatUnderscores
}
type Boolean struct {
	True  bool
	False bool
}
type Datetime struct {
	Key1 time.Time
	Key2 time.Time
	Key3 time.Time
}
type Array struct {
	Key1 []int
	Key2 []string
	Key3 [][]int
	Key4 [][]interface{}
	Key5 []int
	Key6 []int
}
type Product struct {
	Name  string `toml:",omitempty"`
	Sku   int64  `toml:",omitempty"`
	Color string `toml:",omitempty"`
}
type Physical struct {
	Color string
	Shape string
}
type Variety struct {
	Name string
}
type Fruit struct {
	Name     string
	Physical Physical
	Variety  []Variety
}
type testStruct struct {
	Table    Table
	X        X
	String   String
	Integer  Integer
	Float    Float
	Boolean  Boolean
	Datetime Datetime
	Array    Array
	Products []Product
	Fruit    []Fruit
}

func theTestStruct() testStruct {
	return testStruct{
		Table: Table{
			Key: "value",
			Subtable: Subtable{
				Key: "another value",
			},
			Inline: Inline{
				Name: Name{
					First: "Tom",
					Last:  "Preston-Werner",
				},
				Point: map[string]int{
					"x": 1,
					"y": 2,
				},
			},
		},
		X: X{},
		String: String{
			Basic: Basic{
				Basic: "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF.",
			},
			Multiline: Multiline{
				Key1: "One\nTwo",
				Key2: "One\nTwo",
				Key3: "One\nTwo",
				Continued: Continued{
					Key1: "The quick brown fox jumps over the lazy dog.",
					Key2: "The quick brown fox jumps over the lazy dog.",
					Key3: "The quick brown fox jumps over the lazy dog.",
				},
			},
			Literal: Literal{
				Winpath:  `C:\Users\nodejs\templates`,
				Winpath2: `\\ServerX\admin$\system32\`,
				Quoted:   `Tom "Dubs" Preston-Werner`,
				Regex:    `<\i\c*\s*>`,
				Multiline: LiteralMultiline{
					Regex2: `I [dw]on't need \d{2} apples`,
					Lines:  "The first newline is\ntrimmed in raw strings.\n   All other whitespace\n   is preserved.\n",
				},
			},
		},
		Integer: Integer{
			Key1: 99,
			Key2: 42,
			Key3: 0,
			Key4: -17,
			Underscores: IntegerUnderscores{
				Key1: 1000,
				Key2: 5349221,
				Key3: 12345,
			},
		},
		Float: Float{
			Fractional: Fractional{
				Key1: 1.0,
				Key2: 3.1415,
				Key3: -0.01,
			},
			Exponent: Exponent{
				Key1: 5e22,
				Key2: 1e6,
				Key3: -2e-2,
			},
			Both: Both{
				Key: 6.626e-34,
			},
			Underscores: FloatUnderscores{
				Key1: 9224617.445991228313,
				Key2: 1e100,
			},
		},
		Boolean: Boolean{
			True:  true,
			False: false,
		},
		Datetime: Datetime{
			Key1: mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T07:32:00Z")),
			Key2: mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00-07:00")),
			Key3: mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00.999999-07:00")),
		},
		Array: Array{
			Key1: []int{1, 2, 3},
			Key2: []string{"red", "yellow", "green"},
			Key3: [][]int{{1, 2}, {3, 4, 5}},
			Key4: [][]interface{}{{int64(1), int64(2)}, {"a", "b", "c"}},
			Key5: []int{1, 2, 3},
			Key6: []int{1, 2},
		},
		Products: []Product{
			{Name: "Hammer", Sku: 738594937},
			{},
			{Name: "Nail", Sku: 284758393, Color: "gray"},
		},
		Fruit: []Fruit{
			{
				Name: "apple",
				Physical: Physical{
					Color: "red",
					Shape: "round",
				},
				Variety: []Variety{
					{Name: "red delicious"},
					{Name: "granny smith"},
				},
			},
			{
				Name: "banana",
				Variety: []Variety{
					{Name: "plantain"},
				},
			},
		},
	}
}

func TestUnmarshal(t *testing.T) {
	data := loadTestData("test.toml")

	var v testStruct
	var actual interface{} = Unmarshal(data, &v)
	var expect interface{} = nil
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, &testStruct{}) => %#v; want %#v`, actual, expect)
	}

	actual = v
	expect = theTestStruct()
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, v); v => %#v; want %#v`, actual, expect)
	}
}

type testcase struct {
	data   string
	err    error
	actual interface{}
	expect interface{}
}

func testUnmarshal(t *testing.T, testcases []testcase) {
	for _, v := range testcases {
		err := Unmarshal([]byte(v.data), v.actual)
		if !reflect.DeepEqual(err, v.err) {
			t.Errorf("Error mismatch for input:\n%s\ngot: %+v\nwant: %+v", v.data, err, v.err)
		}
		if !reflect.DeepEqual(v.actual, v.expect) {
			t.Errorf("Unmarshal value mismatch for input:\n%s\ndiff:\n%s", v.data, pretty.Compare(v.actual, v.expect))
		}
	}
}

func TestUnmarshal_WithString(t *testing.T) {
	type testStruct struct {
		Str      string
		Key1     string
		Key2     string
		Key3     string
		Winpath  string
		Winpath2 string
		Quoted   string
		Regex    string
		Regex2   string
		Lines    string
	}
	testUnmarshal(t, []testcase{
		{`str = "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF."`, nil, &testStruct{}, &testStruct{
			Str: "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF.",
		}},
		{`key1 = "One\nTwo"
key2 = """One\nTwo"""
key3 = """
One
Two"""
`, nil, &testStruct{}, &testStruct{
			Key1: "One\nTwo",
			Key2: "One\nTwo",
			Key3: "One\nTwo",
		}},
		{`# The following strings are byte-for-byte equivalent:
key1 = "The quick brown fox jumps over the lazy dog."

key2 = """
The quick brown \


  fox jumps over \
    the lazy dog."""

key3 = """\
       The quick brown \
       fox jumps over \
       the lazy dog.\
       """`, nil, &testStruct{}, &testStruct{
			Key1: "The quick brown fox jumps over the lazy dog.",
			Key2: "The quick brown fox jumps over the lazy dog.",
			Key3: "The quick brown fox jumps over the lazy dog.",
		}},
		{`# What you see is what you get.
winpath  = 'C:\Users\nodejs\templates'
winpath2 = '\\ServerX\admin$\system32\'
quoted   = 'Tom "Dubs" Preston-Werner'
regex    = '<\i\c*\s*>'`, nil, &testStruct{}, &testStruct{
			Winpath:  `C:\Users\nodejs\templates`,
			Winpath2: `\\ServerX\admin$\system32\`,
			Quoted:   `Tom "Dubs" Preston-Werner`,
			Regex:    `<\i\c*\s*>`,
		}},
		{`regex2 = '''I [dw]on't need \d{2} apples'''
lines  = '''
The first newline is
trimmed in raw strings.
   All other whitespace
   is preserved.
'''`, nil, &testStruct{}, &testStruct{
			Regex2: `I [dw]on't need \d{2} apples`,
			Lines:  "The first newline is\ntrimmed in raw strings.\n   All other whitespace\n   is preserved.\n",
		}},
	})
}

func TestUnmarshal_WithInteger(t *testing.T) {
	type testStruct struct {
		Intval int64
	}
	testUnmarshal(t, []testcase{
		{`intval = 0`, nil, &testStruct{}, &testStruct{0}},
		{`intval = +0`, nil, &testStruct{}, &testStruct{0}},
		{`intval = -0`, nil, &testStruct{}, &testStruct{-0}},
		{`intval = 1`, nil, &testStruct{}, &testStruct{1}},
		{`intval = +1`, nil, &testStruct{}, &testStruct{1}},
		{`intval = -1`, nil, &testStruct{}, &testStruct{-1}},
		{`intval = 10`, nil, &testStruct{}, &testStruct{10}},
		{`intval = 777`, nil, &testStruct{}, &testStruct{777}},
		{`intval = 2147483647`, nil, &testStruct{}, &testStruct{2147483647}},
		{`intval = 2147483648`, nil, &testStruct{}, &testStruct{2147483648}},
		{`intval = +2147483648`, nil, &testStruct{}, &testStruct{2147483648}},
		{`intval = -2147483648`, nil, &testStruct{}, &testStruct{-2147483648}},
		{`intval = -2147483649`, nil, &testStruct{}, &testStruct{-2147483649}},
		{`intval = 9223372036854775807`, nil, &testStruct{}, &testStruct{9223372036854775807}},
		{`intval = +9223372036854775807`, nil, &testStruct{}, &testStruct{9223372036854775807}},
		{`intval = -9223372036854775808`, nil, &testStruct{}, &testStruct{-9223372036854775808}},
		{`intval = 1_000`, nil, &testStruct{}, &testStruct{1000}},
		{`intval = 5_349_221`, nil, &testStruct{}, &testStruct{5349221}},
		{`intval = 1_2_3_4_5`, nil, &testStruct{}, &testStruct{12345}},
		// overflow
		{
			data:   `intval = 9223372036854775808`,
			err:    lineErrorField(1, "toml.testStruct.Intval", &overflowError{reflect.Int64, "9223372036854775808"}),
			actual: &testStruct{},
			expect: &testStruct{},
		},
		{
			data:   `intval = +9223372036854775808`,
			err:    lineErrorField(1, "toml.testStruct.Intval", &overflowError{reflect.Int64, "+9223372036854775808"}),
			actual: &testStruct{},
			expect: &testStruct{},
		},
		{
			data:   `intval = -9223372036854775809`,
			err:    lineErrorField(1, "toml.testStruct.Intval", &overflowError{reflect.Int64, "-9223372036854775809"}),
			actual: &testStruct{},
			expect: &testStruct{},
		},
		// invalid _
		{`intval = _1_000`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`intval = 1_000_`, lineError(1, errParse), &testStruct{}, &testStruct{}},
	})
}

func TestUnmarshal_WithFloat(t *testing.T) {
	type testStruct struct {
		Floatval float64
	}
	testUnmarshal(t, []testcase{
		{`floatval = 0.0`, nil, &testStruct{}, &testStruct{0.0}},
		{`floatval = +0.0`, nil, &testStruct{}, &testStruct{0.0}},
		{`floatval = -0.0`, nil, &testStruct{}, &testStruct{-0.0}},
		{`floatval = 0.1`, nil, &testStruct{}, &testStruct{0.1}},
		{`floatval = +0.1`, nil, &testStruct{}, &testStruct{0.1}},
		{`floatval = -0.1`, nil, &testStruct{}, &testStruct{-0.1}},
		{`floatval = 0.2`, nil, &testStruct{}, &testStruct{0.2}},
		{`floatval = +0.2`, nil, &testStruct{}, &testStruct{0.2}},
		{`floatval = -0.2`, nil, &testStruct{}, &testStruct{-0.2}},
		{`floatval = 1.0`, nil, &testStruct{}, &testStruct{1.0}},
		{`floatval = +1.0`, nil, &testStruct{}, &testStruct{1.0}},
		{`floatval = -1.0`, nil, &testStruct{}, &testStruct{-1.0}},
		{`floatval = 1.1`, nil, &testStruct{}, &testStruct{1.1}},
		{`floatval = +1.1`, nil, &testStruct{}, &testStruct{1.1}},
		{`floatval = -1.1`, nil, &testStruct{}, &testStruct{-1.1}},
		{`floatval = 3.1415`, nil, &testStruct{}, &testStruct{3.1415}},
		{`floatval = +3.1415`, nil, &testStruct{}, &testStruct{3.1415}},
		{`floatval = -3.1415`, nil, &testStruct{}, &testStruct{-3.1415}},
		{`floatval = 10.2e5`, nil, &testStruct{}, &testStruct{10.2e5}},
		{`floatval = +10.2e5`, nil, &testStruct{}, &testStruct{10.2e5}},
		{`floatval = -10.2e5`, nil, &testStruct{}, &testStruct{-10.2e5}},
		{`floatval = 10.2E5`, nil, &testStruct{}, &testStruct{10.2e5}},
		{`floatval = +10.2E5`, nil, &testStruct{}, &testStruct{10.2e5}},
		{`floatval = -10.2E5`, nil, &testStruct{}, &testStruct{-10.2e5}},
		{`floatval = 5e+22`, nil, &testStruct{}, &testStruct{5e+22}},
		{`floatval = 1e6`, nil, &testStruct{}, &testStruct{1e6}},
		{`floatval = -2E-2`, nil, &testStruct{}, &testStruct{-2E-2}},
		{`floatval = 6.626e-34`, nil, &testStruct{}, &testStruct{6.626e-34}},
		{`floatval = 9_224_617.445_991_228_313`, nil, &testStruct{}, &testStruct{9224617.445991228313}},
		{`floatval = 1e1_00`, nil, &testStruct{}, &testStruct{1e100}},
		{`floatval = 1e02`, nil, &testStruct{}, &testStruct{1e2}},
		{`floatval = _1e1_00`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`floatval = 1e1_00_`, lineError(1, errParse), &testStruct{}, &testStruct{}},
	})
}

func TestUnmarshal_WithBoolean(t *testing.T) {
	type testStruct struct {
		Boolval bool
	}
	testUnmarshal(t, []testcase{
		{`boolval = true`, nil, &testStruct{}, &testStruct{true}},
		{`boolval = false`, nil, &testStruct{}, &testStruct{false}},
	})
}

func TestUnmarshal_WithDatetime(t *testing.T) {
	type testStruct struct {
		Datetimeval time.Time
	}
	testUnmarshal(t, []testcase{
		{`datetimeval = 1979-05-27T07:32:00Z`, nil, &testStruct{}, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T07:32:00Z")),
		}},
		{`datetimeval = 2014-09-13T12:37:39Z`, nil, &testStruct{}, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "2014-09-13T12:37:39Z")),
		}},
		{`datetimeval = 1979-05-27T00:32:00-07:00`, nil, &testStruct{}, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00-07:00")),
		}},
		{`datetimeval = 1979-05-27T00:32:00.999999-07:00`, nil, &testStruct{}, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00.999999-07:00")),
		}},
	})
}

func TestUnmarshal_WithArray(t *testing.T) {
	testUnmarshal(t, []testcase{
		{`arrayval = []`, nil, &struct{ Arrayval []interface{} }{}, &struct{ Arrayval []interface{} }{}},
		{`arrayval = [ 1 ]`, nil, &struct{ Arrayval []int }{},
			&struct {
				Arrayval []int
			}{
				[]int{1},
			}},
		{`arrayval = [ 1, 2, 3 ]`, nil, &struct{ Arrayval []int }{},
			&struct {
				Arrayval []int
			}{
				[]int{1, 2, 3},
			}},
		{`arrayval = [ 1, 2, 3, ]`, nil, &struct{ Arrayval []int }{},
			&struct {
				Arrayval []int
			}{
				[]int{1, 2, 3},
			}},
		{`arrayval = ["red", "yellow", "green"]`, nil, &struct{ Arrayval []string }{},
			&struct{ Arrayval []string }{
				[]string{"red", "yellow", "green"},
			}},
		{`arrayval = [ "all", 'strings', """are the same""", '''type''']`, nil, &struct{ Arrayval []string }{},
			&struct{ Arrayval []string }{
				[]string{"all", "strings", "are the same", "type"},
			}},
		{`arrayval = [[1,2],[3,4,5]]`, nil, &struct{ Arrayval [][]int }{},
			&struct{ Arrayval [][]int }{
				[][]int{
					[]int{1, 2},
					[]int{3, 4, 5},
				},
			}},
		{`arrayval = [ [ 1, 2 ], ["a", "b", "c"] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{int64(1), int64(2)},
					[]interface{}{"a", "b", "c"},
				},
			}},
		{`arrayval = [ [ 1, 2 ], [ [3, 4], [5, 6] ] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{int64(1), int64(2)},
					[]interface{}{
						[]interface{}{int64(3), int64(4)},
						[]interface{}{int64(5), int64(6)},
					},
				},
			}},
		{`arrayval = [ [ 1, 2 ], [ [3, 4], [5, 6], [7, 8] ] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{int64(1), int64(2)},
					[]interface{}{
						[]interface{}{int64(3), int64(4)},
						[]interface{}{int64(5), int64(6)},
						[]interface{}{int64(7), int64(8)},
					},
				},
			}},
		{`arrayval = [ [[ 1, 2 ]], [3, 4], [5, 6] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{
						[]interface{}{int64(1), int64(2)},
					},
					[]interface{}{int64(3), int64(4)},
					[]interface{}{int64(5), int64(6)},
				},
			}},
		{
			data:   `arrayval = [ 1, 2.0 ] # note: this is NOT ok`,
			err:    lineErrorField(1, "struct { Arrayval []interface {} }.Arrayval", errArrayMultiType),
			actual: &struct{ Arrayval []interface{} }{},
			expect: &struct{ Arrayval []interface{} }{},
		},
		{`key = [
  1, 2, 3
]`, nil, &struct{ Key []int }{},
			&struct{ Key []int }{
				[]int{1, 2, 3},
			}},
		{`key = [
  1,
  2, # this is ok
]`, nil, &struct{ Key []int }{},
			&struct{ Key []int }{
				[]int{1, 2},
			}},
	})
}

func TestUnmarshal_WithTable(t *testing.T) {
	type W struct{}
	type Z struct {
		W W
	}
	type Y struct {
		Z Z
	}
	type X struct {
		Y Y
	}
	type A struct {
		D int
		B struct {
			C int
		}
	}
	type testStruct struct {
		Table struct {
			Key string
		}
		Dog struct {
			Tater struct{}
		}
		X X
		A A
	}
	type testQuotedKeyStruct struct {
		Dog struct {
			TaterMan struct {
				Type string
			} `toml:"tater.man"`
		}
	}
	type testQuotedKeyWithWhitespaceStruct struct {
		Dog struct {
			TaterMan struct {
				Type string
			} `toml:"tater . man"`
		}
	}
	type testStructWithMap struct {
		Servers map[string]struct {
			IP string
			DC string
		}
	}
	testUnmarshal(t, []testcase{
		{`[table]`, nil, &testStruct{}, &testStruct{}},
		{`[table]
key = "value"`, nil, &testStruct{},
			&testStruct{
				Table: struct {
					Key string
				}{
					Key: "value",
				},
			}},
		{`[dog.tater]`, nil, &testStruct{},
			&testStruct{
				Dog: struct {
					Tater struct{}
				}{
					Tater: struct{}{},
				},
			}},
		{`[dog."tater.man"]
type = "pug"`, nil, &testQuotedKeyStruct{},
			&testQuotedKeyStruct{
				Dog: struct {
					TaterMan struct {
						Type string
					} `toml:"tater.man"`
				}{
					TaterMan: struct {
						Type string
					}{
						Type: "pug",
					},
				},
			}},
		{`[dog."tater . man"]
type = "pug"`, nil, &testQuotedKeyWithWhitespaceStruct{},
			&testQuotedKeyWithWhitespaceStruct{
				Dog: struct {
					TaterMan struct {
						Type string
					} `toml:"tater . man"`
				}{
					TaterMan: struct {
						Type string
					}{
						Type: "pug",
					},
				},
			}},
		{`[x.y.z.w] # for this to work`, nil, &testStruct{},
			&testStruct{
				X: X{},
			}},
		{`[ x .  y  . z . w ]`, nil, &testStruct{},
			&testStruct{
				X: X{},
			}},
		{`[ x . "y" . z . "w" ]`, nil, &testStruct{},
			&testStruct{
				X: X{},
			}},
		{`table = {}`, nil, &testStruct{}, &testStruct{}},
		{`table = { key = "value" }`, nil, &testStruct{}, &testStruct{
			Table: struct {
				Key string
			}{
				Key: "value",
			},
		}},
		{`x = { y = { "z" = { w = {} } } }`, nil, &testStruct{}, &testStruct{X: X{}}},
		{`[a.b]
c = 1

[a]
d = 2`, nil, &testStruct{},
			&testStruct{
				A: struct {
					D int
					B struct {
						C int
					}
				}{
					D: 2,
					B: struct {
						C int
					}{
						C: 1,
					},
				},
			}},
		{`# DO NOT DO THIS

[a]
b = 1

[a]
c = 2`, lineError(6, fmt.Errorf("table `a' is in conflict with normal table in line 3")), &testStruct{}, &testStruct{}},
		{`# DO NOT DO THIS EITHER

[a]
b = 1

[a.b]
c = 2`, lineError(6, fmt.Errorf("key `b' is in conflict with line 4")), &testStruct{}, &testStruct{}},
		{`# DO NOT DO THIS EITHER

[a.b]
c = 2

[a]
b = 1`, lineError(7, fmt.Errorf("key `b' is in conflict with normal table in line 3")), &testStruct{}, &testStruct{}},
		{`[]`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`[a.]`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`[a..b]`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`[.b]`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`[.]`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{` = "no key name" # not allowed`, lineError(1, errParse), &testStruct{}, &testStruct{}},
		{`[servers]
[servers.alpha]
ip = "10.0.0.1"
dc = "eqdc10"
[servers.beta]
ip = "10.0.0.2"
dc = "eqdc10"
`, nil, &testStructWithMap{},
			&testStructWithMap{
				Servers: map[string]struct {
					IP string
					DC string
				}{
					"alpha": {
						IP: "10.0.0.1",
						DC: "eqdc10",
					},
					"beta": {
						IP: "10.0.0.2",
						DC: "eqdc10",
					},
				},
			}},
		{`
[a]
d = 2
y = 3
        `, lineErrorField(4, "toml.testStruct.A", fmt.Errorf("field corresponding to 'y' is not defined in toml.A")), &testStruct{}, &testStruct{A: A{D: 2}}},
	})
}

func TestUnmarshal_WithArrayTable(t *testing.T) {
	type Product struct {
		Name  string
		SKU   int64
		Color string
	}
	type Physical struct {
		Color string
		Shape string
	}
	type Variety struct {
		Name string
	}
	type Fruit struct {
		Name     string
		Physical Physical
		Variety  []Variety
	}
	type testStruct struct {
		Products []Product
		Fruit    []Fruit
	}
	type testStructWithMap struct {
		Fruit []map[string][]struct {
			Name string
		}
	}
	testUnmarshal(t, []testcase{
		{`[[products]]
		name = "Hammer"
		sku = 738594937

		[[products]]

		[[products]]
		name = "Nail"
		sku = 284758393
		color = "gray"`, nil, &testStruct{},
			&testStruct{
				Products: []Product{
					{Name: "Hammer", SKU: 738594937},
					{},
					{Name: "Nail", SKU: 284758393, Color: "gray"},
				},
			}},
		{`products = [{name = "Hammer", sku = 738594937}, {},
{name = "Nail", sku = 284758393, color = "gray"}]`, nil, &testStruct{}, &testStruct{
			Products: []Product{
				{Name: "Hammer", SKU: 738594937},
				{},
				{Name: "Nail", SKU: 284758393, Color: "gray"},
			},
		}},
		{`[[fruit]]
		name = "apple"

		[fruit.physical]
		color = "red"
		shape = "round"

		[[fruit.variety]]
		name = "red delicious"

		[[fruit.variety]]
		name = "granny smith"

		[[fruit]]
		name = "banana"

		[fruit.physical]
		color = "yellow"
		shape = "lune"

		[[fruit.variety]]
		name = "plantain"`, nil, &testStruct{},
			&testStruct{
				Fruit: []Fruit{
					{
						Name: "apple",
						Physical: Physical{
							Color: "red",
							Shape: "round",
						},
						Variety: []Variety{
							{Name: "red delicious"},
							{Name: "granny smith"},
						},
					},
					{
						Name: "banana",
						Physical: Physical{
							Color: "yellow",
							Shape: "lune",
						},
						Variety: []Variety{
							{Name: "plantain"},
						},
					},
				},
			}},
		{`[[fruit]]

		[[fruit.variety]]
		name = "red delicious"

		[[fruit.variety]]
		name = "granny smith"

		[[fruit]]

		[[fruit.variety]]
		name = "plantain"

		[[fruit.area]]
		name = "phillippines"`, nil, &testStructWithMap{},
			&testStructWithMap{
				Fruit: []map[string][]struct {
					Name string
				}{
					{
						"variety": {
							{Name: "red delicious"},
							{Name: "granny smith"},
						},
					},
					{
						"variety": {
							{Name: "plantain"},
						},
						"area": {
							{Name: "phillippines"},
						},
					},
				},
			}},
		{`# INVALID TOML DOC
		[[fruit]]
		name = "apple"

		[[fruit.variety]]
		name = "red delicious"

		# This table conflicts with the previous table
		[fruit.variety]
		name = "granny smith"`, lineError(9, fmt.Errorf("table `fruit.variety' is in conflict with array table in line 5")), &testStruct{}, &testStruct{}},
		{`# INVALID TOML DOC
		[[fruit]]
		name = "apple"

		[fruit.variety]
		name = "granny smith"

		# This table conflicts with the previous table
		[[fruit.variety]]
		name = "red delicious"`, lineError(9, fmt.Errorf("table `fruit.variety' is in conflict with normal table in line 5")), &testStruct{}, &testStruct{}},
	})
}

type testUnmarshalerString string

func (u *testUnmarshalerString) UnmarshalTOML(data []byte) error {
	*u = testUnmarshalerString("Unmarshaled: " + string(data))
	return nil
}

type testUnmarshalerStruct struct {
	Title  string
	Author testUnmarshalerString
}

func (u *testUnmarshalerStruct) UnmarshalTOML(data []byte) error {
	u.Title = "Unmarshaled: " + string(data)
	return nil
}

func TestUnmarshal_WithUnmarshaler(t *testing.T) {
	type testStruct struct {
		Title         testUnmarshalerString
		MaxConn       testUnmarshalerString
		Ports         testUnmarshalerString
		Servers       testUnmarshalerString
		Table         testUnmarshalerString
		Arraytable    testUnmarshalerString
		ArrayOfStruct []testUnmarshalerStruct
	}
	data := loadTestData("unmarshal-unmarshaler.toml")
	var v testStruct
	if err := Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	actual := v
	expect := testStruct{
		Title:      `Unmarshaled: "testtitle"`,
		MaxConn:    `Unmarshaled: 777`,
		Ports:      `Unmarshaled: [8080, 8081, 8082]`,
		Servers:    `Unmarshaled: [1, 2, 3]`,
		Table:      "Unmarshaled: [table]\nname = \"alice\"",
		Arraytable: "Unmarshaled: [[arraytable]]\nname = \"alice\"\n[[arraytable]]\nname = \"bob\"",
		ArrayOfStruct: []testUnmarshalerStruct{
			{
				Title:  "Unmarshaled: [[array_of_struct]]\ntitle = \"Alice's Adventures in Wonderland\"\nauthor = \"Lewis Carroll\"",
				Author: "",
			},
		},
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, &v); v => %#v; want %#v`, actual, expect)
	}
}

func TestUnmarshal_WithUnmarshalerForTopLevelStruct(t *testing.T) {
	data := `title = "Alice's Adventures in Wonderland"
author = "Lewis Carroll"
`
	var v testUnmarshalerStruct
	if err := Unmarshal([]byte(data), &v); err != nil {
		t.Fatal(err)
	}
	actual := v
	expect := testUnmarshalerStruct{
		Title: `Unmarshaled: title = "Alice's Adventures in Wonderland"
author = "Lewis Carroll"
`,
		Author: "",
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, &v); v => %#v; want %#v`, actual, expect)
	}
}

type testTextUnmarshaler string

var errTextUnmarshaler = errors.New("UnmarshalText called with data = error")

func (x *testTextUnmarshaler) UnmarshalText(data []byte) error {
	*x = testTextUnmarshaler("Unmarshaled: " + string(data))
	if string(data) == "error" {
		return errTextUnmarshaler
	}
	return nil
}

func TestUnmarshal_WithTextUnmarshaler(t *testing.T) {
	type testStruct struct {
		Str        testTextUnmarshaler
		Int        testTextUnmarshaler
		Float      testTextUnmarshaler
		Arraytable []testStruct
	}

	tests := []testcase{
		{
			data:   string(loadTestData("unmarshal-textunmarshaler.toml")),
			actual: &testStruct{},
			expect: &testStruct{
				Str:        "Unmarshaled: str",
				Int:        "Unmarshaled: 11",
				Float:      "Unmarshaled: 12.0",
				Arraytable: []testStruct{{Str: "Unmarshaled: str2", Int: "Unmarshaled: 22", Float: "Unmarshaled: 23.0"}},
			},
		},
		{
			data:   `str = "error"`,
			actual: &testStruct{},
			expect: &testStruct{Str: "Unmarshaled: error"},
			err:    lineErrorField(1, "toml.testStruct.Str", errTextUnmarshaler),
		},
	}
	testUnmarshal(t, tests)
}

type testUnmarshalerRecString string

func (u *testUnmarshalerRecString) UnmarshalTOML(fn func(interface{}) error) error {
	var s string
	if err := fn(&s); err != nil {
		return err
	}
	*u = testUnmarshalerRecString("Unmarshaled: " + s)
	return nil
}

type testUnmarshalerRecStruct struct {
	a, b int
}

func (u *testUnmarshalerRecStruct) UnmarshalTOML(fn func(interface{}) error) error {
	var uu struct{ A, B int }
	if err := fn(&uu); err != nil {
		return err
	}
	u.a, u.b = uu.A, uu.B
	return nil
}

func TestUnmarshal_WithUnmarshalerRec(t *testing.T) {
	type testStruct struct {
		String     testUnmarshalerRecString
		Struct     testUnmarshalerRecStruct
		Arraytable []testStruct
	}
	var v testStruct
	err := Unmarshal(loadTestData("unmarshal-unmarshalerrec.toml"), &v)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	expect := testStruct{
		String: "Unmarshaled: str1",
		Struct: testUnmarshalerRecStruct{a: 1, b: 2},
		Arraytable: []testStruct{
			{
				String: "Unmarshaled: str2",
				Struct: testUnmarshalerRecStruct{a: 3, b: 4},
			},
		},
	}
	if !reflect.DeepEqual(v, expect) {
		t.Errorf(`toml.Unmarshal(data, &v); v => %#v; want %#v`, v, expect)
	}
}

func TestUnmarshal_WithMultibyteString(t *testing.T) {
	type testStruct struct {
		Name    string
		Numbers []string
	}
	v := testStruct{}
	data := `name = "七一〇七"
numbers = ["壱", "弐", "参"]
`
	if err := Unmarshal([]byte(data), &v); err != nil {
		t.Fatal(err)
	}
	actual := v
	expect := testStruct{
		Name:    "七一〇七",
		Numbers: []string{"壱", "弐", "参"},
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal([]byte(data), &v); v => %#v; want %#v`, actual, expect)
	}
}

func TestUnmarshal_WithPointers(t *testing.T) {
	type Inline struct {
		Key1 string
		Key2 *string
		Key3 **string
	}
	type Table struct {
		Key1 *string
		Key2 **string
		Key3 ***string
	}
	type testStruct struct {
		Inline *Inline
		Tables []*Table
	}
	type testStruct2 struct {
		Inline **Inline
		Tables []**Table
	}
	type testStruct3 struct {
		Inline ***Inline
		Tables []***Table
	}
	data := `
inline = { key1 = "test", key2 = "a", key3 = "b" }
[[tables]]
key1 = "a"
key2 = "a"
key3 = "a"
[[tables]]
key1 = "b"
key2 = "b"
key3 = "b"
`
	s1 := "a"
	s2 := &s1
	s3 := &s2
	s4 := &s3
	s5 := "b"
	s6 := &s5
	s7 := &s6
	s8 := &s7
	i1 := &Inline{"test", s2, s7}
	i2 := &i1
	i3 := &i2
	t1 := &Table{s2, s3, s4}
	t2 := &Table{s6, s7, s8}
	t3 := &t1
	t4 := &t2
	sc := &testStruct{
		Inline: i1, Tables: []*Table{t1, t2},
	}
	ac := &testStruct{}
	testUnmarshal(t, []testcase{
		{data, nil, ac, sc},
		{data, nil, &testStruct2{}, &testStruct2{
			Inline: i2,
			Tables: []**Table{&t1, &t2},
		}},
		{data, nil, &testStruct3{}, &testStruct3{
			Inline: i3,
			Tables: []***Table{&t3, &t4},
		}},
	})
}

func TestUnmarshalMap(t *testing.T) {
	testUnmarshal(t, []testcase{
		{
			data: `
name = "evan"
foo = 1
`,
			actual: map[string]interface{}{},
			expect: map[string]interface{}{"name": "evan", "foo": int64(1)},
		},
		{
			data: `[p]
first = "evan"
`,
			actual: map[string]*Name{},
			expect: map[string]*Name{"p": {First: "evan"}},
		},
		{
			data: `foo = 1
bar = 2
`,
			actual: map[testTextUnmarshaler]int{},
			expect: map[testTextUnmarshaler]int{"Unmarshaled: foo": 1, "Unmarshaled: bar": 2},
		},
		{
			data: `1 = 1
-2 = 2
`,
			actual: map[int]int{},
			expect: map[int]int{1: 1, -2: 2},
		},
		{
			data: `1 = 1
-129 = 2
`,
			actual: map[int8]int{},
			expect: map[int8]int{1: 1},
			err:    lineError(2, &overflowError{reflect.Int8, "-129"}),
		},
	})
}

func TestUnmarshal_WithQuotedKeyValue(t *testing.T) {
	type nestedStruct struct {
		Truthy bool
	}
	type testStruct struct {
		Table map[string]nestedStruct
	}

	testUnmarshal(t, []testcase{
		{`
[table]
"some.key" = {truthy = true}
`, nil, &testStruct{},
			&testStruct{Table: map[string]nestedStruct{
				"some.key": {Truthy: true},
			}},
		},
	})
}

func TestUnmarshal_WithCustomPrimitiveType(t *testing.T) {
	type (
		String string
		Int    int
		Bool   bool
	)
	type X struct {
		S String
		I Int
		B Bool
	}

	input := `
s = "string"
i = 1
b = true
`
	testUnmarshal(t, []testcase{
		{input, nil, &X{}, &X{"string", 1, true}},
	})
}
