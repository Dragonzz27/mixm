package service

import (
	"math"
	"testing"

	"mixm/backend/internal/model"
)

// TestGrading 验证各题型的作答状态与评分规则。
func TestGrading(t *testing.T) {
	cases := []struct {
		name              string
		q, a              model.Record
		answered, correct bool
	}{
		{
			"single",
			model.Record{"question_type": "single_choice", "correct_index": 1},
			model.Record{"selected_index": 1},
			true,
			true,
		},
		{
			"unanswered",
			model.Record{"question_type": "single_choice", "correct_index": 0},
			nil,
			false,
			false,
		},
		{
			"multi unordered",
			model.Record{"question_type": "multi_choice", "correct_indices": []int{0, 2}},
			model.Record{"answer_text": "[2,0]"},
			true,
			true,
		},
		{
			"multi partial",
			model.Record{"question_type": "multi_choice", "correct_indices": []int{0, 2}},
			model.Record{"answer_text": "[2]"},
			true,
			false,
		},
		{
			"multi extra",
			model.Record{"question_type": "multi_choice", "correct_indices": []int{0, 2}},
			model.Record{"answer_text": "[0,1,2]"},
			true,
			false,
		},
		{
			"text normalized",
			model.Record{"question_type": "fill_blank", "correct_answer": "Go Client"},
			model.Record{"answer_text": " go\nCLIENT "},
			true,
			true,
		},
		{
			"text empty",
			model.Record{"question_type": "short_answer", "correct_answer": ""},
			model.Record{"answer_text": " "},
			false,
			false,
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				a, b := gradeQuestion(c.q, c.a)
				if a != c.answered || b != c.correct {
					t.Fatalf("got %v,%v", a, b)
				}
			},
		)
	}
}

// TestWeightedScores 验证难度权重分配后的总分保持一致。
func TestWeightedScores(t *testing.T) {
	qs := []model.Record{}

	for i := 0; i < 47; i++ {
		qs = append(qs, model.Record{"difficulty": i%5 + 1})
	}

	assignScores(qs)
	sum := 0.0

	for _, q := range qs {
		sum += model.Number(q, "score")
	}

	if math.Abs(sum-100) > 0.000001 {
		t.Fatalf("total score %f", sum)
	}

	if model.Number(qs[4], "score") <= model.Number(qs[0], "score") {
		t.Fatal("difficulty weighting lost")
	}
}

// TestNavigationCompatibility 验证菜单筛选与关联路由的权限范围。
func TestNavigationCompatibility(t *testing.T) {
	u := model.Record{"is_superuser": false}
	routes, menu := navigationFor(u, []string{"ClientMyInstances"})
	if len(menu) != 1 || menu[0]["key"] != "ClientMyInstances" {
		t.Fatal("menu filter incorrect")
	}

	found := false

	for _, r := range routes {
		for _, c := range model.Objects(r["children"]) {
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
