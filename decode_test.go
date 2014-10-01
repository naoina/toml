package toml_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/naoina/toml"
)

const (
	dataDir = "testdata"
)

func loadTestData() ([]byte, error) {
	f := filepath.Join(dataDir, "test.toml")
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func mustTime(tm time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return tm
}

func TestUnmarshal(t *testing.T) {
	type Owner struct {
		Name string
		Org  string `toml:"organization"`
		Bio  string
		Dob  time.Time
	}
	type Database struct {
		Server        string
		Ports         []int
		ConnectionMax uint
		Enabled       bool
	}
	type Server struct {
		IP string
		DC string
	}
	type Servers struct {
		Alpha Server
		Beta  Server
	}
	type Clients struct {
		Data  [][]interface{}
		Hosts []string
	}
	type testStruct struct {
		Title    string
		Owner    Owner
		Database Database
		Servers  Servers
		Clients  Clients
	}

	data, err := loadTestData()
	if err != nil {
		t.Fatal(err)
	}
	var v testStruct
	var actual interface{} = toml.Unmarshal(data, &v)
	var expect interface{} = nil
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, &testStruct{}) => %#v; want %#v`, actual, expect)
	}

	actual = v
	expect = testStruct{
		Title: "TOML Example",
		Owner: Owner{
			Name: "Tom Preston-Werner",
			Org:  "GitHub",
			Bio:  "GitHub Cofounder & CEO\nLikes tater tots and beer.",
			Dob:  mustTime(time.Parse("20060102150405", "19790527073200")),
		},
		Database: Database{
			Server:        "192.168.1.1",
			Ports:         []int{8001, 8001, 8002},
			ConnectionMax: 5000,
			Enabled:       true,
		},
		Servers: Servers{
			Alpha: Server{
				IP: "10.0.0.1",
				DC: "eqdc10",
			},
			Beta: Server{
				IP: "10.0.0.2",
				DC: "eqdc10",
			},
		},
		Clients: Clients{
			Data: [][]interface{}{
				[]interface{}{"gamma", "delta"},
				[]interface{}{int64(1), int64(2)},
			},
			Hosts: []string{"alpha", "omega"},
		},
	}
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
		var actual error = toml.Unmarshal([]byte(v.data), v.actual)
		var expect error = v.err
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`toml.Unmarshal([]byte(%#v), %#v) => %#v; want %#v`, v.data, nil, actual, expect)
		}
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`toml.Unmarshal([]byte(%#v), v); v => %#v; want %#v`, v.data, v.actual, v.expect)
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
			Lines:  `The first newline is trimmed in raw strings.    All other whitespace    is preserved.`,
		}},
	})
}

func TestUnmarshal_WithInteger(t *testing.T) {
	type testStruct struct {
		Intval int64
	}
	testUnmarshal(t, []testcase{
		{`intval = 0`, nil, &testStruct{}, &testStruct{0}},
		{`intval = -0`, nil, &testStruct{}, &testStruct{-0}},
		{`intval = 1`, nil, &testStruct{}, &testStruct{1}},
		{`intval = -1`, nil, &testStruct{}, &testStruct{-1}},
		{`intval = 10`, nil, &testStruct{}, &testStruct{10}},
		{`intval = 777`, nil, &testStruct{}, &testStruct{777}},
		{`intval = 2147483647`, nil, &testStruct{}, &testStruct{2147483647}},
		{`intval = 2147483648`, nil, &testStruct{}, &testStruct{2147483648}},
		{`intval = -2147483648`, nil, &testStruct{}, &testStruct{-2147483648}},
		{`intval = -2147483649`, nil, &testStruct{}, &testStruct{-2147483649}},
		{`intval = 9223372036854775807`, nil, &testStruct{}, &testStruct{9223372036854775807}},
		{`intval = 9223372036854775808`, fmt.Errorf(`toml: unmarshal: line 1: toml_test.testStruct.Intval: strconv.ParseInt: parsing "9223372036854775808": value out of range`), &testStruct{}, &testStruct{}},
		{`intval = -9223372036854775808`, nil, &testStruct{}, &testStruct{-9223372036854775808}},
		{`intval = -9223372036854775809`, fmt.Errorf(`toml: unmarshal: line 1: toml_test.testStruct.Intval: strconv.ParseInt: parsing "-9223372036854775809": value out of range`), &testStruct{}, &testStruct{}},
	})
}

func TestUnmarshal_WithFloat(t *testing.T) {
	type testStruct struct {
		Floatval float64
	}
	testUnmarshal(t, []testcase{
		{`floatval = 0.0`, nil, &testStruct{}, &testStruct{0.0}},
		{`floatval = -0.0`, nil, &testStruct{}, &testStruct{-0.0}},
		{`floatval = 0.1`, nil, &testStruct{}, &testStruct{0.1}},
		{`floatval = -0.1`, nil, &testStruct{}, &testStruct{-0.1}},
		{`floatval = 0.2`, nil, &testStruct{}, &testStruct{0.2}},
		{`floatval = -0.2`, nil, &testStruct{}, &testStruct{-0.2}},
		{`floatval = 1.0`, nil, &testStruct{}, &testStruct{1.0}},
		{`floatval = -1.0`, nil, &testStruct{}, &testStruct{-1.0}},
		{`floatval = 1.1`, nil, &testStruct{}, &testStruct{1.1}},
		{`floatval = -1.1`, nil, &testStruct{}, &testStruct{-1.1}},
		{`floatval = 3.1415`, nil, &testStruct{}, &testStruct{3.1415}},
		{`floatval = -3.1415`, nil, &testStruct{}, &testStruct{-3.1415}},
		{`floatval = 10.2e5`, nil, &testStruct{}, &testStruct{10.2e5}},
		{`floatval = -10.2e5`, nil, &testStruct{}, &testStruct{-10.2e5}},
		{`floatval = 10.2E5`, nil, &testStruct{}, &testStruct{10.2e5}},
		{`floatval = -10.2E5`, nil, &testStruct{}, &testStruct{10.2e5}},
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
			mustTime(time.Parse("20060102150405", "19790527073200")),
		}},
		{`datetimeval = 2014-09-13T12:37:39Z`, nil, &testStruct{}, &testStruct{
			mustTime(time.Parse("20060102150405", "20140913123739")),
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
					[]interface{}{1, 2},
					[]interface{}{"a", "b", "c"},
				},
			}},
		{`arrayval = [ [ 1, 2 ], [ [3, 4], [5, 6] ] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{1, 2},
					[]interface{}{
						[]interface{}{3, 4},
						[]interface{}{5, 6},
					},
				},
			}},
		{`arrayval = [ [ 1, 2 ], [ [3, 4], [5, 6], [7, 8] ] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{1, 2},
					[]interface{}{
						[]interface{}{3, 4},
						[]interface{}{5, 6},
						[]interface{}{7, 8},
					},
				},
			}},
		{`arrayval = [ [[ 1, 2 ]], [3, 4], [5, 6] ] # this is ok`, nil, &struct{ Arrayval [][]interface{} }{},
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					[]interface{}{
						[]interface{}{1, 2},
					},
					[]interface{}{3, 4},
					[]interface{}{5, 6},
				},
			}},
		{`arrayval = [ 1, 2.0 ] # note: this is NOT ok`, fmt.Errorf("toml: unmarshal: line 1: struct { Arrayval []interface {} }.Arrayval: array cannot contain multiple types"), &struct{ Arrayval []interface{} }{}, &struct{ Arrayval []interface{} }{}},
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
	type testStruct struct {
		Table struct {
			Key string
		}
		Dog struct {
			Tater struct{}
		}
		X X
		A struct {
			D int
			B struct {
				C int
			}
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
		{`[x.y.z.w] # for this to work`, nil, &testStruct{},
			&testStruct{
				X: X{},
			}},
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
c = 2`, fmt.Errorf("toml: line 6: table `a' is in conflict with normal table in line 3"), &testStruct{}, &testStruct{}},
		{`# DO NOT DO THIS EITHER

[a]
b = 1

[a.b]
c = 2`, fmt.Errorf("toml: line 6: key `b' is in conflict with line 4"), &testStruct{}, &testStruct{}},
		{`# DO NOT DO THIS EITHER

[a.b]
c = 2

[a]
b = 1`, fmt.Errorf("toml: line 7: key `b' is in conflict with normal table in line 3"), &testStruct{}, &testStruct{}},
		{`[]`, fmt.Errorf("toml: line 1: parse error"), &testStruct{}, &testStruct{}},
		{`[a.]`, fmt.Errorf("toml: line 1: parse error"), &testStruct{}, &testStruct{}},
		{`[a..b]`, fmt.Errorf("toml: line 1: parse error"), &testStruct{}, &testStruct{}},
		{`[.b]`, fmt.Errorf("toml: line 1: parse error"), &testStruct{}, &testStruct{}},
		{`[.]`, fmt.Errorf("toml: line 1: parse error"), &testStruct{}, &testStruct{}},
		{` = "no key name" # not allowed`, fmt.Errorf("toml: line 1: parse error"), &testStruct{}, &testStruct{}},
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
						Variety: []Variety{
							{Name: "plantain"},
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
  name = "granny smith"`, fmt.Errorf("toml: line 9: table `fruit.variety' is in conflict with array table in line 5"), &testStruct{}, &testStruct{}},
	})
}
