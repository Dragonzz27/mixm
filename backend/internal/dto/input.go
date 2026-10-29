package dto

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"mixm/backend/internal/model"
)

// Optional 区分更新请求中缺失、显式空值和零值三种情况。
type Optional[T any] struct {
	Value   T
	Present bool
	Null    bool
}

// UnmarshalJSON 记录字段是否出现，并区分显式空值和有效数据。
func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.Present = true
	o.Null = bytes.Equal(bytes.TrimSpace(b), []byte("null"))
	if o.Null {
		return nil
	}

	return json.Unmarshal(b, &o.Value)
}

// Field 返回字段的存在状态及其实际值。
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

// requestFields 将类型化请求转换为字段映射，保留更新语义。
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
