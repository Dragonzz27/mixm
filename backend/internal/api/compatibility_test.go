package api

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"mixm-edu/backend/internal/config"
)

//go:embed legacy_routes.json
var legacyRoutes []byte

func canonical(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "{") {
			parts[i] = "{}"
		}
	}
	return strings.Join(parts, "/")
}
func TestLegacyRouteCoverage(t *testing.T) {
	s := New(config.Config{}, nil, nil, nil)
	registered := map[string]bool{}
	for _, p := range s.patterns {
		registered[canonical(p)] = true
	}
	var old []string
	if e := json.Unmarshal(legacyRoutes, &old); e != nil {
		t.Fatal(e)
	}
	for _, p := range old {
		if !registered[canonical(p)] {
			t.Errorf("legacy endpoint missing: %s", p)
		}
	}
}
