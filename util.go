package toml

import (
	"fmt"
	"reflect"
	"strings"
)

const fieldTagName = "toml"

// fieldCache maps normalized field names to their position in a struct.
type fieldCache map[string]fieldInfo

type fieldInfo struct {
	index   []int
	name    string
	ignored bool
}

func makeFieldCache(rt reflect.Type) fieldCache {
	fc := make(fieldCache)
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		// skip unexported fields
		if ft.PkgPath != "" && !ft.Anonymous {
			continue
		}
		col, _ := extractTag(ft.Tag.Get(fieldTagName))
		key := col
		if col == "" || col == "-" {
			key = normFieldName(ft.Name)
		}
		fc[key] = fieldInfo{index: ft.Index, name: ft.Name, ignored: col == "-"}
	}
	return fc
}

func (fc fieldCache) findField(rv reflect.Value, name string) (reflect.Value, string, error) {
	info, found := fc[normFieldName(name)]
	if !found {
		return reflect.Value{}, "", fmt.Errorf("field corresponding to `%s' is not defined in %v", name, rv.Type())
	} else if info.ignored {
		return reflect.Value{}, "", fmt.Errorf("field corresponding to `%s' in %v cannot be set through TOML", name, rv.Type())
	}
	return rv.FieldByIndex(info.index), info.name, nil
}

func normFieldName(s string) string {
	return strings.Replace(strings.ToLower(s), "_", "", -1)
}

func extractTag(tag string) (col, rest string) {
	tags := strings.SplitN(tag, ",", 2)
	if len(tags) == 2 {
		return strings.TrimSpace(tags[0]), strings.TrimSpace(tags[1])
	}
	return strings.TrimSpace(tags[0]), ""
}
