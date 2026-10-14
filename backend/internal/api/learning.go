package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"mixm-edu/backend/internal/store"
)

func (s *Server) selectQuestions(r *http.Request, v store.Row) ([]store.Row, store.Row, error) {
	args := []any{}
	where := []string{"review_status='approved'", s.visibility(r, "questions")}
	add := func(expr string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(expr, len(args)))
	}
	if id := store.String(v, "course_id"); id != "" {
		add("EXISTS(SELECT 1 FROM question_tags t WHERE t.question_id=questions.id AND t.category_id IN (WITH RECURSIVE c AS (SELECT id FROM categories WHERE id=$%d UNION SELECT x.id FROM categories x JOIN c ON x.parent_id=c.id) SELECT id FROM c))", id)
	}
	if tags := store.Strings(v["knowledge_tag_ids"]); len(tags) > 0 {
		add("EXISTS(SELECT 1 FROM question_tags t WHERE t.question_id=questions.id AND t.category_id=ANY($%d))", tags)
	}
	if v["difficulty"] != nil {
		add("difficulty=$%d", int(store.Number(v, "difficulty")))
	}
	types := store.Strings(v["question_types"])
	for i, t := range types {
		types[i] = bankType(t)
	}
	if len(types) > 0 {
		add("type=ANY($%d)", types)
	}
	order := "random()"
	if id := store.String(v, "practice_id"); id != "" {
		if _, e := store.One(r.Context(), s.DB, "SELECT * FROM practices WHERE id=$1 AND "+s.visibility(r, "practices"), id); e != nil {
			return nil, nil, e
		}
		add("id IN (SELECT question_id FROM practice_questions WHERE practice_id=$%d)", id)
		args = append(args, id)
		order = fmt.Sprintf("(SELECT sort_order FROM practice_questions WHERE practice_id=$%d AND question_id=questions.id)", len(args))
	}
	rows, e := store.Query(r.Context(), s.DB, "SELECT * FROM questions WHERE "+strings.Join(where, " AND ")+" ORDER BY "+order+" LIMIT 1000", args...)
	if e != nil {
		return nil, nil, e
	}
	counts := store.Object(v["type_counts"])
	out := []store.Row{}
	shortages := store.Row{}
	if len(counts) > 0 {
		keys := []string{}
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, kind := range keys {
			need := int(store.Number(counts, kind))
			if need < 0 || need > 50 {
				return nil, nil, fail(422, "每种题型数量必须为 0-50")
			}
			for _, q := range rows {
				if need > 0 && store.String(q, "type") == bankType(kind) {
					out = append(out, q)
					need--
				}
			}
			if need > 0 {
				shortages[kind] = need
			}
		}
	} else {
		n := int(store.Number(v, "question_count"))
		if n == 0 {
			n = 10
			if store.String(v, "practice_id") != "" {
				n = len(rows)
			}
		}
		if n < 1 || n > 50 {
			return nil, nil, fail(422, "题目数量必须为 1-50")
		}
		if n > len(rows) {
			n = len(rows)
		}
		out = rows[:n]
	}
	if len(out) > 50 {
		return nil, nil, fail(422, "试卷最多 50 题")
	}
	for i, q := range out {
		out[i], e = s.enrich(r, s.DB, "questions", q)
		if e != nil {
			return nil, nil, e
		}
	}
	return out, shortages, nil
}
func bankType(t string) string {
	switch t {
	case "fill_blank":
		return "blank"
	case "short_answer":
		return "essay"
	}
	return t
}
func simulationQuestion(q store.Row) store.Row {
	kind := store.String(q, "type")
	if kind == "blank" {
		kind = "fill_blank"
	}
	if kind == "essay" {
		kind = "short_answer"
	}
	options := []string{}
	indices := []int{}
	for i, opt := range store.Objects(q["options"]) {
		options = append(options, store.String(opt, "content"))
		if store.Bool(opt, "is_correct") {
			indices = append(indices, i)
		}
	}
	tags := []string{}
	for _, t := range store.Objects(q["knowledge_tags"]) {
		tags = append(tags, store.String(t, "name"))
	}
	if len(tags) == 0 {
		tags = store.Strings(q["tags"])
	}
	if len(tags) == 0 {
		tags = []string{"未分类"}
	}
	index := 0
	if len(indices) > 0 {
		index = indices[0]
	}
	if kind == "true_false" && len(options) == 0 {
		options = []string{"正确", "错误"}
		if strings.EqualFold(store.String(q, "answer"), "false") || store.String(q, "answer") == "错误" {
			index = 1
		}
		indices = []int{index}
	}
	return store.Row{"id": q["id"], "stem": q["title"], "options": options, "correct_index": index, "correct_indices": indices, "correct_answer": q["answer"], "difficulty": q["difficulty"], "knowledge_tag": strings.Join(tags, " / "), "knowledge_tags": tags, "question_type": kind}
}
func assignScores(qs []store.Row) {
	weights := map[int]float64{1: 1, 2: 1.15, 3: 1.3, 4: 1.55, 5: 2}
	total := 0.0
	for _, q := range qs {
		w := weights[int(store.Number(q, "difficulty"))]
		if w == 0 {
			w = 1.3
		}
		total += w
	}
	running := 0.0
	for i, q := range qs {
		w := weights[int(store.Number(q, "difficulty"))]
		if w == 0 {
			w = 1.3
		}
		score := math.Round(100*w/total*100) / 100
		if i == len(qs)-1 {
			score = math.Round((100-running)*100) / 100
		}
		q["score"] = score
		running += score
	}
}
func normalized(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, s)
}
func gradeQuestion(q, a store.Row) (bool, bool) {
	kind := store.String(q, "question_type")
	if kind == "fill_blank" || kind == "short_answer" {
		text := normalized(store.String(a, "answer_text"))
		return text != "", text != "" && text == normalized(store.String(q, "correct_answer"))
	}
	if kind == "multi_choice" {
		picks := []int{}
		_ = json.Unmarshal([]byte(store.String(a, "answer_text")), &picks)
		if len(picks) == 0 && a["selected_index"] != nil {
			picks = []int{int(store.Number(a, "selected_index"))}
		}
		expected := []int{}
		b, _ := json.Marshal(q["correct_indices"])
		_ = json.Unmarshal(b, &expected)
		if len(expected) == 0 {
			expected = []int{int(store.Number(q, "correct_index"))}
		}
		got, want := map[int]bool{}, map[int]bool{}
		for _, x := range picks {
			got[x] = true
		}
		for _, x := range expected {
			want[x] = true
		}
		correct := len(got) == len(want)
		for x := range want {
			correct = correct && got[x]
		}
		return len(got) > 0, correct
	}
	return a["selected_index"] != nil, a["selected_index"] != nil && store.Number(a, "selected_index") == store.Number(q, "correct_index")
}
func (s *Server) learningRoutes() {
	s.route("POST /api/practices/auto-generate", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Practices"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		qs, short, e := s.selectQuestions(r, v)
		return store.Row{"questions": qs, "shortages": short}, e
	})
	s.route("POST /api/simulation/generate", s.generateSimulation)
	s.route("POST /api/simulation/submit", s.submitSimulation)
	s.route("GET /api/simulation/attempt/current", func(w http.ResponseWriter, r *http.Request) (any, error) {
		rows, e := store.Query(r.Context(), s.DB, "SELECT * FROM simulation_attempt WHERE user_id=$1 AND status='in_progress' ORDER BY started_at DESC LIMIT 1", user(r)["id"])
		if e != nil {
			return nil, e
		}
		if len(rows) == 0 {
			return store.Row{"attempt_id": nil, "attempt_status": nil, "questions": []store.Row{}}, nil
		}
		a := rows[0]
		return store.Row{"attempt_id": a["id"], "attempt_status": a["status"], "questions": store.Objects(a["questions_payload"])}, nil
	})
	s.route("GET /api/simulation/attempt/{id}/review", func(w http.ResponseWriter, r *http.Request) (any, error) {
		a, e := store.One(r.Context(), s.DB, "SELECT * FROM simulation_attempt WHERE id=$1 AND user_id=$2", r.PathValue("id"), user(r)["id"])
		if e != nil {
			return nil, e
		}
		if a["status"] != "submitted" {
			return nil, fail(400, "当前试卷尚未交卷")
		}
		return store.Row{"attempt_id": a["id"], "attempt_status": a["status"], "questions": store.Objects(a["questions_payload"]), "answers": store.Objects(a["answers_payload"]), "result": store.Object(a["result_payload"])}, nil
	})
	s.route("GET /api/simulation/traces", func(w http.ResponseWriter, r *http.Request) (any, error) {
		rows, total, e := s.list(r, "simulation_answer_trace", "user_id=$1", []any{user(r)["id"]})
		return store.Row{"total": total, "items": rows}, e
	})
	s.route("GET /api/simulation/kt-status", func(w http.ResponseWriter, r *http.Request) (any, error) {
		out := store.Row{"backend": s.Config.KTBackend, "available": s.Config.KTBackend == "heuristic", "model": "heuristic"}
		if s.Config.KTBackend == "kt_service" {
			req, e := http.NewRequestWithContext(r.Context(), "GET", strings.TrimSuffix(s.Config.KTURL, "/")+"/health", nil)
			if e != nil {
				return nil, e
			}
			resp, e := s.do(req, s.Config.KTTimeout)
			if e == nil {
				defer resp.Body.Close()
				out["available"] = resp.StatusCode == 200
				var data store.Row
				if json.NewDecoder(resp.Body).Decode(&data) == nil {
					out["service"] = data
					out["model"] = data["model"]
				}
			}
		}
		return out, nil
	})
	s.route("POST /api/questions/generate", s.llmStream(false))
	s.route("POST /api/knowledge-tree/analyze", s.llmStream(true))
	s.route("POST /api/knowledge-tree/save", s.saveKnowledgeTree)
}
func (s *Server) generateSimulation(w http.ResponseWriter, r *http.Request) (any, error) {
	v, e := body(r)
	if e != nil {
		return nil, e
	}
	for _, spec := range []struct {
		key    string
		lo, hi float64
	}{{"question_count", 1, 50}, {"difficulty", 1, 5}, {"time_limit_minutes", 1, 180}} {
		if v[spec.key] != nil {
			n := store.Number(v, spec.key)
			if n < spec.lo || n > spec.hi || n != float64(int(n)) {
				return nil, fail(422, spec.key+" 无效")
			}
		}
	}
	if id := store.String(v, "course_id"); id != "" {
		if _, e = store.Get(r.Context(), s.DB, "categories", id); e != nil {
			return nil, e
		}
	}
	tx, e := s.DB.Begin(r.Context())
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", user(r)["id"]); e != nil {
		return nil, e
	}
	active, e := store.Query(r.Context(), tx, "SELECT * FROM simulation_attempt WHERE user_id=$1 AND status='in_progress' ORDER BY started_at DESC LIMIT 1", user(r)["id"])
	if e != nil {
		return nil, e
	}
	if len(active) > 0 {
		return store.Row{"attempt_id": active[0]["id"], "attempt_status": "in_progress", "settings": v, "questions": store.Objects(active[0]["questions_payload"])}, nil
	}
	mode := store.String(v, "generation_mode")
	if mode == "" {
		mode = "random"
	}
	if mode != "random" && mode != "practice" && mode != "ai" {
		return nil, fail(422, "出题模式无效")
	}
	if mode == "practice" && store.String(v, "practice_id") == "" {
		return nil, fail(422, "practice 模式需要 practice_id")
	}
	qs := []store.Row{}
	if mode == "ai" {
		n := int(store.Number(v, "question_count"))
		if n == 0 {
			n = 10
		}
		if n < 1 || n > 50 {
			return nil, fail(422, "题目数量必须为 1-50")
		}
		query := "SELECT id,title,content FROM contents WHERE status='published' AND is_active=true"
		args := []any{}
		if id := store.String(v, "course_id"); id != "" {
			query += " AND EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id IN (WITH RECURSIVE c AS (SELECT id FROM categories WHERE id=$1 UNION SELECT x.id FROM categories x JOIN c ON x.parent_id=c.id) SELECT id FROM c))"
			args = append(args, id)
		}
		contents, err := store.Query(r.Context(), s.DB, query+" LIMIT 100", args...)
		if err != nil {
			return nil, err
		}
		if len(contents) == 0 {
			return nil, fail(400, "暂无已发布教学内容")
		}
		for i := 0; i < n; i++ {
			content := contents[i%len(contents)]
			types := store.Strings(v["question_types"])
			kind := "single_choice"
			if len(types) > 0 {
				kind = bankType(types[i%len(types)])
			}
			difficulty := int(store.Number(v, "difficulty"))
			if difficulty == 0 {
				difficulty = 3
			}
			q, err := s.llm(r, fmt.Sprintf("根据教学内容生成一道题目。只返回 JSON：title,type,difficulty,answer,explanation,options(option_key,content,is_correct)。题型 %s，难度 %d（1-5）。输入：%s", kind, difficulty, store.JSON(content)))
			if err != nil {
				return nil, err
			}
			q["id"] = uuid.NewString()
			if e = validateResource("questions", "", q); e != nil {
				return nil, fail(502, "LLM 题目格式无效")
			}
			tags, err := store.Query(r.Context(), s.DB, "SELECT c.id,c.name FROM categories c JOIN content_categories cc ON cc.category_id=c.id WHERE cc.content_id=$1", content["id"])
			if err != nil {
				return nil, err
			}
			q["knowledge_tags"] = tags
			qs = append(qs, simulationQuestion(q))
		}
	} else {
		selected, _, e := s.selectQuestions(r, v)
		if e != nil {
			return nil, e
		}
		for _, q := range selected {
			qs = append(qs, simulationQuestion(q))
		}
	}
	if len(qs) == 0 {
		return nil, fail(400, "没有符合条件的题目")
	}
	assignScores(qs)
	a, e := store.Write(r.Context(), tx, "simulation_attempt", "", store.Row{"user_id": user(r)["id"], "paper_source": mode, "practice_id": v["practice_id"], "status": "in_progress", "questions_payload": store.JSON(qs), "answers_payload": "[]"})
	if e != nil {
		return nil, e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return nil, e
	}
	return store.Row{"attempt_id": a["id"], "attempt_status": "in_progress", "settings": v, "questions": qs}, nil
}
func (s *Server) submitSimulation(w http.ResponseWriter, r *http.Request) (any, error) {
	v, e := body(r)
	if e != nil {
		return nil, e
	}
	tx, e := s.DB.Begin(r.Context())
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", user(r)["id"]); e != nil {
		return nil, e
	}
	a, e := store.One(r.Context(), tx, "SELECT * FROM simulation_attempt WHERE id=$1 AND user_id=$2 FOR UPDATE", v["attempt_id"], user(r)["id"])
	if e != nil {
		return nil, e
	}
	if a["status"] != "in_progress" {
		return nil, fail(409, "试卷已提交")
	}
	qs := store.Objects(a["questions_payload"])
	if len(qs) == 0 {
		return nil, fail(400, "试卷为空")
	}
	answers := store.Objects(v["answers"])
	answerMap := map[string]store.Row{}
	for _, ans := range answers {
		id := store.String(ans, "question_id")
		if answerMap[id] != nil {
			return nil, fail(422, "重复的题目答案")
		}
		answerMap[id] = ans
	}
	score := 0.0
	answered, correct := 0, 0
	byTag := map[string][]int{}
	byQuestion := store.Row{}
	tagQuestions := map[string][]string{}
	traces := []store.Row{}
	for _, q := range qs {
		qid := store.String(q, "id")
		isAnswered, isCorrect := gradeQuestion(q, answerMap[qid])
		if isAnswered {
			answered++
		}
		if isCorrect {
			correct++
			score += store.Number(q, "score")
		}
		tags := store.Strings(q["knowledge_tags"])
		byQuestion[qid] = tags
		for _, tag := range tags {
			counts := byTag[tag]
			if counts == nil {
				counts = []int{0, 0}
			}
			counts[0]++
			if isCorrect {
				counts[1]++
			}
			byTag[tag] = counts
			tagQuestions[tag] = append(tagQuestions[tag], qid)
			tr := store.Row{"question_id": qid, "knowledge_tag": tag, "selected_index": answerMap[qid]["selected_index"], "correct_index": q["correct_index"], "is_correct": isCorrect}
			traces = append(traces, tr)
			dbTrace := store.Row{}
			for k, x := range tr {
				dbTrace[k] = x
			}
			dbTrace["user_id"] = user(r)["id"]
			if _, e = store.Write(r.Context(), tx, "simulation_answer_trace", "", dbTrace); e != nil {
				return nil, e
			}
		}
	}
	accuracy, mastery, previous := store.Row{}, store.Row{}, store.Row{}
	weak := []string{}
	for tag, counts := range byTag {
		acc := math.Round(float64(counts[1])/float64(counts[0])*10000) / 100
		accuracy[tag] = acc
		if acc < 60 {
			weak = append(weak, tag)
		}
		rows, err := store.Query(r.Context(), tx, "SELECT mastery FROM knowledge_mastery WHERE user_id=$1 AND knowledge_tag=$2", user(r)["id"], tag)
		if err != nil {
			return nil, err
		}
		p := 0.5
		if len(rows) > 0 {
			p = store.Number(rows[0], "mastery")
		}
		previous[tag] = p
		for _, trace := range traces {
			if trace["knowledge_tag"] != tag {
				continue
			}
			if store.Bool(trace, "is_correct") {
				p += 0.18 * (1 - p)
			} else {
				p -= 0.22 * p
			}
		}
		_, err = tx.Exec(r.Context(), "INSERT INTO knowledge_mastery(id,user_id,knowledge_tag,mastery,attempts,correct_attempts,updated_at) VALUES($1,$2,$3,$4,$5,$6,now()) ON CONFLICT(user_id,knowledge_tag) DO UPDATE SET attempts=knowledge_mastery.attempts+EXCLUDED.attempts,correct_attempts=knowledge_mastery.correct_attempts+EXCLUDED.correct_attempts,mastery=EXCLUDED.mastery,updated_at=now()", uuid.NewString(), user(r)["id"], tag, p, counts[0], counts[1])
		if err != nil {
			return nil, err
		}
		mastery[tag] = math.Round(p*10000) / 10000
	}
	sort.Slice(weak, func(i, j int) bool { return store.Number(accuracy, weak[i]) < store.Number(accuracy, weak[j]) })
	if s.Config.KTBackend == "kt_service" {
		payload := store.Row{"user_id": user(r)["id"], "traces": traces, "prev_mastery": previous}
		b, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(r.Context(), "POST", strings.TrimSuffix(s.Config.KTURL, "/")+"/predict", bytes.NewReader(b))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err := s.do(req, s.Config.KTTimeout)
			if err == nil {
				var predicted store.Row
				err = json.NewDecoder(resp.Body).Decode(&predicted)
				resp.Body.Close()
				if err == nil && resp.StatusCode == 200 {
					for tag, value := range store.Object(predicted["mastery_by_tag"]) {
						n := store.Number(store.Row{"v": value}, "v")
						if n >= 0 && n <= 1 {
							mastery[tag] = n
							if _, e = tx.Exec(r.Context(), "UPDATE knowledge_mastery SET mastery=$1 WHERE user_id=$2 AND knowledge_tag=$3", n, user(r)["id"], tag); e != nil {
								return nil, e
							}
						}
					}
				}
			}
		}
	}
	result := store.Row{"attempt_id": a["id"], "attempt_status": "submitted", "total_score": 100, "score": math.Round(score*100) / 100, "total_questions": len(qs), "answered_questions": answered, "correct_count": correct, "accuracy": math.Round(float64(correct)/float64(len(qs))*10000) / 100, "weak_knowledge_points": weak, "knowledge_accuracy": accuracy, "knowledge_mastery": mastery, "q_matrix": traces, "q_matrix_by_question": byQuestion, "q_matrix_by_tag": tagQuestions}
	if _, e = store.Write(r.Context(), tx, "simulation_attempt", store.String(a, "id"), store.Row{"status": "submitted", "answers_payload": store.JSON(answers), "result_payload": store.JSON(result), "submitted_at": time.Now()}); e != nil {
		return nil, e
	}
	return result, tx.Commit(r.Context())
}
