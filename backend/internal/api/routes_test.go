package api

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"mixm-edu/backend/internal/config"
)

//go:embed routes.json
var requiredRoutes []byte

func canonical(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "{") {
			parts[i] = "{}"
		}
	}
	return strings.Join(parts, "/")
}
func TestAPIRouteCoverage(t *testing.T) {
	s := New(config.Config{}, nil, nil, nil)
	registered := map[string]bool{}
	for _, p := range s.patterns {
		registered[canonical(p)] = true
	}
	var expected []string
	if e := json.Unmarshal(requiredRoutes, &expected); e != nil {
		t.Fatal(e)
	}
	for _, p := range expected {
		if !registered[canonical(p)] {
			t.Errorf("API endpoint missing: %s", p)
		}
	}
}
