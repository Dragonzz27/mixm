package model

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Record represents extensible resource fields and computed projections.
type Record = map[string]any

func String(r Record, k string) string { s, _ := r[k].(string); return s }
func Bool(r Record, k string) bool     { v, _ := r[k].(bool); return v }
func Number(r Record, k string) float64 {
	switch v := r[k].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		n, _ := v.Float64()
		return n
	}
	return 0
}
func Strings(v any) []string {
	out := []string{}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		for _, a := range x {
			if s, ok := a.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}
func JSON(v any) string { b, _ := json.Marshal(v); return string(b) }
func Object(v any) Record {
	if r, ok := v.(map[string]any); ok {
		return r
	}
	if s, ok := v.(string); ok {
		var r Record
		if json.Unmarshal([]byte(s), &r) == nil && r != nil {
			return r
		}
	}
	return Record{}
}
func Objects(v any) []Record {
	out := []Record{}
	if s, ok := v.(string); ok {
		_ = json.Unmarshal([]byte(s), &out)
		return out
	}
	switch a := v.(type) {
	case []Record:
		return a
	case []any:
		for _, r := range a {
			out = append(out, Object(r))
		}
	}
	return out
}

func EntityRecord(entity any) Record {
	value := reflect.Indirect(reflect.ValueOf(entity))
	typ := value.Type()
	record := Record{}
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		key := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		if key == "" || key == "-" {
			continue
		}
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				record[key] = nil
				continue
			}
			field = field.Elem()
		}
		v := field.Interface()
		if raw, ok := v.(json.RawMessage); ok {
			if len(raw) == 0 {
				v = nil
			} else {
				_ = json.Unmarshal(raw, &v)
			}
		}
		record[key] = v
	}
	return record
}
