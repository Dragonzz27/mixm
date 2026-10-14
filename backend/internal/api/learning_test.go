package api

import (
	"math"
	"net/http/httptest"
	"testing"

	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

func TestGrading(t *testing.T) {
	cases := []struct {
		name              string
		q, a              store.Row
		answered, correct bool
	}{{"single", store.Row{"question_type": "single_choice", "correct_index": 1}, store.Row{"selected_index": 1}, true, true}, {"unanswered", store.Row{"question_type": "single_choice", "correct_index": 0}, nil, false, false}, {"multi unordered", store.Row{"question_type": "multi_choice", "correct_indices": []int{0, 2}}, store.Row{"answer_text": "[2,0]"}, true, true}, {"multi partial", store.Row{"question_type": "multi_choice", "correct_indices": []int{0, 2}}, store.Row{"answer_text": "[2]"}, true, false}, {"multi extra", store.Row{"question_type": "multi_choice", "correct_indices": []int{0, 2}}, store.Row{"answer_text": "[0,1,2]"}, true, false}, {"text normalized", store.Row{"question_type": "fill_blank", "correct_answer": "Go Client"}, store.Row{"answer_text": " go\nCLIENT "}, true, true}, {"text empty", store.Row{"question_type": "short_answer", "correct_answer": ""}, store.Row{"answer_text": " "}, false, false}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, b := gradeQuestion(c.q, c.a)
			if a != c.answered || b != c.correct {
				t.Fatalf("got %v,%v", a, b)
			}
		})
	}
}
func TestWeightedScores(t *testing.T) {
	qs := []store.Row{}
	for i := 0; i < 47; i++ {
		qs = append(qs, store.Row{"difficulty": i%5 + 1})
	}
	assignScores(qs)
	sum := 0.0
	for _, q := range qs {
		sum += store.Number(q, "score")
	}
	if math.Abs(sum-100) > 0.000001 {
		t.Fatalf("total score %f", sum)
	}
	if store.Number(qs[4], "score") <= store.Number(qs[0], "score") {
		t.Fatal("difficulty weighting lost")
	}
}
func TestRoutesRequireAuth(t *testing.T) {
	s := New(config.Config{Secret: "test"}, nil, nil, nil)
	for _, path := range []string{"/api/auth/users", "/api/categories/", "/api/instances/queue/tasks", "/api/simulation/attempt/current", "/api/clusters/nodes"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Fatalf("%s returned %d", path, w.Code)
		}
	}
	req := httptest.NewRequest("OPTIONS", "/api/categories", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatalf("preflight %d", w.Code)
	}
}
func TestNavigationCompatibility(t *testing.T) {
	u := store.Row{"is_superuser": false}
	routes, menu := navigationFor(u, []string{"ClientMyInstances"})
	if len(menu) != 1 || menu[0]["key"] != "ClientMyInstances" {
		t.Fatal("menu filter incorrect")
	}
	found := false
	for _, r := range routes {
		for _, c := range store.Objects(r["children"]) {
			if c["name"] == "ClientInstanceDetail" {
				found = true
			}
			if c["name"] == "Users" {
				t.Fatal("admin route leaked")
			}
		}
	}
	if !found {
		t.Fatal("instance detail route missing")
	}
}
