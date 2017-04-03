package toml

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// toCamelCase returns a copy of the string s with all Unicode letters mapped to their camel case.
// It will convert to upper case previous letter of '_' and first letter, and remove letter of '_'.
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))
	upper := false
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			result = append(result, unicode.ToUpper(r))
			upper = false
			continue
		}
		result = append(result, r)
	}
	result[0] = unicode.ToUpper(result[0])
	return string(result)
}

const (
	fieldTagName = "toml"
)

func matchField(fieldName, name string) bool {
	return fieldName == strings.Title(name) ||
		fieldName == toCamelCase(name) ||
		fieldName == strings.ToUpper(name)
}

func findField(rv reflect.Value, name string) (field reflect.Value, fieldName string, err error) {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		// skip unexported fields
		if ft.PkgPath != "" && !ft.Anonymous {
			continue
		}

		col, _ := extractTag(ft.Tag.Get(fieldTagName))
		if col != "-" && name == col {
			return rv.Field(i), ft.Name, nil
		}
		if matchField(ft.Name, name) {
			if col == "-" {
				return field, "", fmt.Errorf("field corresponding to `%s' in %v cannot be set through TOML", name, rv.Type())
			}
			return rv.Field(i), ft.Name, nil
		}
	}
	return field, "", fmt.Errorf("field corresponding to `%s' is not defined in %v", name, rv.Type())
}

func extractTag(tag string) (col, rest string) {
	tags := strings.SplitN(tag, ",", 2)
	if len(tags) == 2 {
		return strings.TrimSpace(tags[0]), strings.TrimSpace(tags[1])
	}
	return strings.TrimSpace(tags[0]), ""
}
