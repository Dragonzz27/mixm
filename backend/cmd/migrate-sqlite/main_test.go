package main

import (
	"testing"
	"time"
)

func TestLegacyValues(t *testing.T) {
	for _, stamp := range []string{"2026-10-13 12:03:04.123456", "2026-10-13T12:03:04Z", "2026-10-13 12:03:04"} {
		v, e := convert("timestamptz", stamp)
		if e != nil {
			t.Fatal(e)
		}
		if v.(time.Time).Location() != time.UTC {
			t.Fatal("legacy timestamps must be interpreted as UTC")
		}
	}
	v, e := convert("boolean", int64(1))
	if e != nil || v != true {
		t.Fatal("SQLite boolean conversion")
	}
	if _, e = convert("jsonb", `["ClientHome"]`); e != nil {
		t.Fatal(e)
	}
	if _, e = convert("jsonb", "not json"); e == nil {
		t.Fatal("invalid JSON accepted")
	}
}
