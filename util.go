package toml

import (
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

func findField(rv reflect.Value, name string) (field reflect.Value, fieldName string, found bool) {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		if ft.PkgPath != "" && !ft.Anonymous {
			continue
		}
		if col, _ := extractTag(ft.Tag.Get(fieldTagName)); col == name {
			return rv.Field(i), ft.Name, true
		}
	}
	for _, name := range []string{
		strings.Title(name),
		toCamelCase(name),
		strings.ToUpper(name),
	} {
		if field := rv.FieldByName(name); field.IsValid() {
			return field, name, true
		}
	}
	return field, "", false
}

func extractTag(tag string) (col, rest string) {
	tags := strings.SplitN(tag, ",", 2)
	if len(tags) == 2 {
		return strings.TrimSpace(tags[0]), strings.TrimSpace(tags[1])
	}
	return strings.TrimSpace(tags[0]), ""
}
