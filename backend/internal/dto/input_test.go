package dto

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

func TestOptionalFields(t *testing.T) {
	var input ResourceInput
	if err := json.Unmarshal([]byte(`{"is_active":false,"sort_order":0,"description":null,"visible_roles":[]}`), &input); err != nil {
		t.Fatal(err)
	}
	fields := input.Fields()
	if fields["is_active"] != false || fields["description"] != nil || !reflect.DeepEqual(fields["visible_roles"], []any{}) {
		t.Fatalf("lost PATCH fields: %#v", fields)
	}
	if _, ok := fields["title"]; ok {
		t.Fatal("absent field was included")
	}
	if _, ok := fields["description"]; !ok {
		t.Fatal("explicit null lost")
	}
}
func TestPagination(t *testing.T) {
	for _, query := range []string{"offset=-1", "offset=x", "limit=0", "limit=1001"} {
		q, _ := url.ParseQuery(query)
		if _, err := ParsePagination(q); err == nil {
			t.Errorf("accepted %s", query)
		}
	}
	p, err := ParsePagination(url.Values{})
	if err != nil || p.Offset != 0 || p.Limit != 100 {
		t.Fatal(p, err)
	}
}
