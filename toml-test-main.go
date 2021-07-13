// +build none

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/naoina/toml"
)

var _ toml.MarshalerRec = (*value)(nil)
var _ toml.UnmarshalerRec = (*value)(nil)

type value struct {
	prim  *prim
	array []*value
	table map[string]*value
}

type prim struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

func toValue(iv interface{}) *value {
	switch gv := iv.(type) {
	case bool:
		return &value{prim: &prim{fmt.Sprint(gv), "bool"}}
	case int64:
		return &value{prim: &prim{fmt.Sprint(gv), "integer"}}
	case float64:
		return &value{prim: &prim{fmt.Sprint(gv), "float"}}
	case string:
		return &value{prim: &prim{fmt.Sprint(gv), "string"}}
	case time.Time:
		return &value{prim: &prim{gv.Format(time.RFC3339Nano), "datetime"}}
	case []interface{}:
		array := make([]*value, len(gv))
		for i := range gv {
			array[i] = toValue(gv[i])
		}
		return &value{array: array}
	case map[string]interface{}:
		table := make(map[string]*value, len(gv))
		for k, v := range gv {
			table[k] = toValue(v)
		}
		return &value{table: table}
	default:
		panic(fmt.Errorf("unhandled %T", iv))
	}
}

// MarshalTOML implements toml.MarshalerRec.
func (v *value) MarshalTOML() (interface{}, error) {
	switch {
	case v.prim != nil:
		gv, _ := v.prim.toInterface()
		return gv, nil
	case v.array != nil:
		return v.array, nil
	case v.table != nil:
		return v.table, nil
	default:
		panic("invalid")
	}
}

func (p *prim) toInterface() (interface{}, error) {
	switch p.Type {
	case "string":
		return p.Value, nil
	case "integer":
		return strconv.Atoi(p.Value)
	case "float":
		return strconv.ParseFloat(p.Value, 64)
	case "datetime":
		return time.Parse("2006-01-02T15:04:05.999999999Z07:00", p.Value)
	case "datetime-local":
		return time.Parse("2006-01-02T15:04:05.999999999", p.Value)
	case "date-local":
		return time.Parse("2006-01-02", p.Value)
	case "time-local":
		return time.Parse("15:04:05.999999999", p.Value)
	case "bool":
		switch p.Value {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
		return nil, errors.New("invalid bool")
	default:
		return nil, fmt.Errorf("invalid type %q", p.Type)
	}
}

// UnmarshalTOML implements toml.UnmarshalerRec.
func (v *value) UnmarshalTOML(decode func(interface{}) error) error {
	var iv interface{}
	if err := decode(&iv); err != nil {
		return err
	}
	*v = *toValue(iv)
	return nil
}

func (v *value) MarshalJSON() ([]byte, error) {
	switch {
	case v.prim != nil:
		return json.Marshal(v.prim)
	case v.array != nil:
		return json.Marshal(v.array)
	case v.table != nil:
		return json.Marshal(v.table)
	default:
		panic("invalid")
	}
}

func (v *value) UnmarshalJSON(input []byte) error {
	// Try array.
	if len(input) > 0 && input[0] == '[' {
		var array []*value
		if err := json.Unmarshal(input, &array); err != nil {
			return err
		}
		*v = value{array: array}
		return nil
	}
	// It might be a primitive value.
	var prim prim
	if err := json.Unmarshal(input, &prim); err == nil {
		if prim.Type != "" {
			*v = value{prim: &prim}
			return nil
		}
	}
	// It's a table object.
	var table map[string]*value
	if err := json.Unmarshal(input, &table); err != nil {
		return err
	}
	*v = value{table: table}
	return nil
}

// This config turns off all table key remapping.
var config = toml.Config{
	NormFieldName: func(typ reflect.Type, keyOrField string) string {
		return keyOrField
	},
	FieldToKey: func(typ reflect.Type, field string) string {
		return field
	},
	WriteEmptyTables: true,
}

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-e" {
			encoder()
			return
		}
	}
	decoder()
}

func encoder() {
	var v map[string]*value // Top-level must be table!
	if err := json.NewDecoder(os.Stdin).Decode(&v); err != nil {
		fmt.Fprintln(os.Stderr, "Error in input JSON:", err)
		os.Exit(1)
	}
	if err := config.NewEncoder(os.Stdout).Encode(&v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func decoder() {
	var v value
	if err := config.NewDecoder(os.Stdin).Decode(&v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(&v); err != nil {
		panic(err)
	}
}
