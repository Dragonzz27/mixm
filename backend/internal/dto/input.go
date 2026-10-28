package dto

import (
	"bytes"
	"encoding/json"

	"io"
	"reflect"
	"strings"

	"mixm/backend/internal/model"
)

// Optional preserves absent, explicit null, and zero-valued PATCH fields.
type Optional[T any] struct {
	Value   T
	Present bool
	Null    bool
}

func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.Present = true
	o.Null = bytes.Equal(bytes.TrimSpace(b), []byte("null"))
	if o.Null {
		return nil
	}
	return json.Unmarshal(b, &o.Value)
}
func (o Optional[T]) Field() (bool, any) {
	if !o.Present {
		return false, nil
	}
	if o.Null {
		return true, nil
	}
	return true, o.Value
}

type Upload struct {
	Reader      io.Reader
	Name        string
	Size        int64
	ContentType string
}
type File struct {
	Path        string
	Reader      io.ReadCloser
	Size        int64
	ContentType string
}

// requestFields projects present fields after typed JSON binding, retaining null and zero values.
func requestFields(input any) model.Record {
	out := model.Record{}
	value := reflect.ValueOf(input)
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		optional, ok := value.Field(i).Interface().(interface{ Field() (bool, any) })
		if !ok {
			continue
		}
		present, v := optional.Field()
		if !present {
			continue
		}
		key := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		raw, err := json.Marshal(v)
		if err != nil {
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var normalized any
		if decoder.Decode(&normalized) == nil {
			out[key] = normalized
		}
	}
	return out
}
