package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"

	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"mixm-edu/backend/internal/dto"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
)

type Learning struct{ *Core }

func bankType(kind string) string { return model.BankType(kind) }

func simulationQuestion(q model.Record) model.Record {
	kind := model.String(q, "type")
	if kind == "blank" {
		kind = "fill_blank"
	}
	if kind == "essay" {
		kind = "short_answer"
	}
	options := []string{}
	indices := []int{}
	for i, opt := range model.Objects(q["options"]) {
		options = append(options, model.String(opt, "content"))
		if model.Bool(opt, "is_correct") {
			indices = append(indices, i)
		}
	}
	tags := []string{}
	for _, t := range model.Objects(q["knowledge_tags"]) {
		tags = append(tags, model.String(t, "name"))
	}
	if len(tags) == 0 {
		tags = model.Strings(q["tags"])
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
		if strings.EqualFold(model.String(q, "answer"), "false") || model.String(q, "answer") == "错误" {
			index = 1
		}
		indices = []int{index}
	}
	return model.Record{"id": q["id"], "stem": q["title"], "options": options, "correct_index": index, "correct_indices": indices, "correct_answer": q["answer"], "difficulty": q["difficulty"], "knowledge_tag": strings.Join(tags, " / "), "knowledge_tags": tags, "question_type": kind}
}
func assignScores(qs []model.Record) {
	weights := map[int]float64{1: 1, 2: 1.15, 3: 1.3, 4: 1.55, 5: 2}
	total := 0.0
	for _, q := range qs {
		w := weights[int(model.Number(q, "difficulty"))]
		if w == 0 {
			w = 1.3
		}
		total += w
	}
	running := 0.0
	for i, q := range qs {
		w := weights[int(model.Number(q, "difficulty"))]
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
func gradeQuestion(q, a model.Record) (bool, bool) {
	kind := model.String(q, "question_type")
	if kind == "fill_blank" || kind == "short_answer" {
		text := normalized(model.String(a, "answer_text"))
		return text != "", text != "" && text == normalized(model.String(q, "correct_answer"))
	}
	if kind == "multi_choice" {
		picks := []int{}
		_ = json.Unmarshal([]byte(model.String(a, "answer_text")), &picks)
		if len(picks) == 0 && a["selected_index"] != nil {
			picks = []int{int(model.Number(a, "selected_index"))}
		}
		expected := []int{}
		b, _ := json.Marshal(q["correct_indices"])
		_ = json.Unmarshal(b, &expected)
		if len(expected) == 0 {
			expected = []int{int(model.Number(q, "correct_index"))}
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
	return a["selected_index"] != nil, a["selected_index"] != nil && model.Number(a, "selected_index") == model.Number(q, "correct_index")
}
func (s *Learning) generateSimulation(ctx context.Context, actor model.Actor, input dto.LearningInput) (dto.Creation, error) {
	v := input.Fields()
	if input.ID != "" {
		v["attempt_id"] = input.ID
	}
	var err error
	for _, spec := range []struct {
		key    string
		lo, hi float64
	}{{"question_count", 1, 50}, {"difficulty", 1, 5}, {"time_limit_minutes", 1, 180}} {
		if v[spec.key] != nil {
			n := model.Number(v, spec.key)
			if n < spec.lo || n > spec.hi || n != float64(int(n)) {
				return dto.Creation{}, fail(http.StatusUnprocessableEntity, spec.key+" 无效")
			}
		}
	}
	if id := model.String(v, "course_id"); id != "" {
		if _, err = repository.Get(ctx, s.DB, "categories", id); err != nil {
			return dto.Creation{}, err
		}
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return dto.Creation{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = repository.LockSimulationUser(ctx, tx, actor.ID); err != nil {
		return dto.Creation{}, err
	}
	active, err := repository.ActiveSimulationAttempts(ctx, tx, actor.ID)
	if err != nil {
		return dto.Creation{}, err
	}
	if len(active) > 0 {
		return dto.Creation{Data: model.Record{"attempt_id": active[0]["id"], "attempt_status": "in_progress", "settings": v, "questions": model.Objects(active[0]["questions_payload"])}}, nil
	}
	mode := model.String(v, "generation_mode")
	if mode == "" {
		mode = "random"
	}
	if mode != "random" && mode != "practice" && mode != "ai" {
		return dto.Creation{}, fail(http.StatusUnprocessableEntity, "出题模式无效")
	}
	if mode == "practice" && model.String(v, "practice_id") == "" {
		return dto.Creation{}, fail(http.StatusUnprocessableEntity, "practice 模式需要 practice_id")
	}
	qs := []model.Record{}
	if mode == "ai" {
		n := int(model.Number(v, "question_count"))
		if n == 0 {
			n = 10
		}
		if n < 1 || n > 50 {
			return dto.Creation{}, fail(http.StatusUnprocessableEntity, "题目数量必须为 1-50")
		}
		contents, err := repository.SimulationContents(ctx, s.DB, model.String(v, "course_id"))
		if err != nil {
			return dto.Creation{}, err
		}
		if len(contents) == 0 {
			return dto.Creation{}, fail(http.StatusBadRequest, "暂无已发布教学内容")
		}
		for i := 0; i < n; i++ {
			content := contents[i%len(contents)]
			types := model.Strings(v["question_types"])
			kind := "single_choice"
			if len(types) > 0 {
				kind = bankType(types[i%len(types)])
			}
			difficulty := int(model.Number(v, "difficulty"))
			if difficulty == 0 {
				difficulty = 3
			}
			q, err := s.llm(ctx, fmt.Sprintf("根据教学内容生成一道题目。只返回 JSON：title,type,difficulty,answer,explanation,options(option_key,content,is_correct)。题型 %s，难度 %d（1-5）。输入：%s", kind, difficulty, model.JSON(content)))
			if err != nil {
				return dto.Creation{}, err
			}
			q["id"] = uuid.NewString()
			if err = validateResource("questions", "", q); err != nil {
				return dto.Creation{}, fail(http.StatusBadGateway, "LLM 题目格式无效")
			}
			tags, err := repository.ContentKnowledgeTags(ctx, s.DB, model.String(content, "id"))
			if err != nil {
				return dto.Creation{}, err
			}
			q["knowledge_tags"] = tags
			qs = append(qs, simulationQuestion(q))
		}
	} else {
		selected, _, err := s.selectQuestions(ctx, actor, v)
		if err != nil {
			return dto.Creation{}, err
		}
		for _, q := range selected {
			qs = append(qs, simulationQuestion(q))
		}
	}
	if len(qs) == 0 {
		return dto.Creation{}, fail(http.StatusBadRequest, "没有符合条件的题目")
	}
	assignScores(qs)
	a, err := repository.Write(ctx, tx, "simulation_attempt", "", model.Record{"user_id": actor.ID, "paper_source": mode, "practice_id": v["practice_id"], "status": "in_progress", "questions_payload": model.JSON(qs), "answers_payload": "[]"})
	if err != nil {
		return dto.Creation{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return dto.Creation{}, err
	}
	return dto.Creation{Created: true, Data: model.Record{"attempt_id": a["id"], "attempt_status": "in_progress", "settings": v, "questions": qs}}, nil
}
func (s *Learning) submitSimulation(ctx context.Context, actor model.Actor, input dto.LearningInput) (any, error) {
	v := input.Fields()
	if input.ID != "" {
		v["attempt_id"] = input.ID
	}
	var err error
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = repository.LockSimulationUserForSubmission(ctx, tx, actor.ID); err != nil {
		return nil, err
	}
	a, err := repository.SimulationAttemptForUpdate(ctx, tx, model.String(v, "attempt_id"), actor.ID)
	if err != nil {
		return nil, err
	}
	if a["status"] != "in_progress" {
		return nil, fail(http.StatusConflict, "试卷已提交")
	}
	qs := model.Objects(a["questions_payload"])
	if len(qs) == 0 {
		return nil, fail(http.StatusBadRequest, "试卷为空")
	}
	answers := model.Objects(v["answers"])
	answerMap := map[string]model.Record{}
	for _, ans := range answers {
		id := model.String(ans, "question_id")
		if answerMap[id] != nil {
			return nil, fail(http.StatusUnprocessableEntity, "重复的题目答案")
		}
		answerMap[id] = ans
	}
	score := 0.0
	answered, correct := 0, 0
	byTag := map[string][]int{}
	byQuestion := model.Record{}
	tagQuestions := map[string][]string{}
	traces := []model.Record{}
	for _, q := range qs {
		qid := model.String(q, "id")
		isAnswered, isCorrect := gradeQuestion(q, answerMap[qid])
		if isAnswered {
			answered++
		}
		if isCorrect {
			correct++
			score += model.Number(q, "score")
		}
		tags := model.Strings(q["knowledge_tags"])
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
			tr := model.Record{"question_id": qid, "knowledge_tag": tag, "selected_index": answerMap[qid]["selected_index"], "correct_index": q["correct_index"], "is_correct": isCorrect}
			traces = append(traces, tr)
			dbTrace := model.Record{}
			for k, x := range tr {
				dbTrace[k] = x
			}
			dbTrace["user_id"] = actor.ID
			if _, err = repository.Write(ctx, tx, "simulation_answer_trace", "", dbTrace); err != nil {
				return nil, err
			}
		}
	}
	accuracy, mastery, previous := model.Record{}, model.Record{}, model.Record{}
	weak := []string{}
	for tag, counts := range byTag {
		acc := math.Round(float64(counts[1])/float64(counts[0])*10000) / 100
		accuracy[tag] = acc
		if acc < 60 {
			weak = append(weak, tag)
		}
		rows, err := repository.KnowledgeMastery(ctx, tx, actor.ID, tag)
		if err != nil {
			return nil, err
		}
		p := 0.5
		if len(rows) > 0 {
			p = model.Number(rows[0], "mastery")
		}
		previous[tag] = p
		for _, trace := range traces {
			if trace["knowledge_tag"] != tag {
				continue
			}
			if model.Bool(trace, "is_correct") {
				p += 0.18 * (1 - p)
			} else {
				p -= 0.22 * p
			}
		}
		_, err = repository.UpsertKnowledgeMastery(ctx, tx, uuid.NewString(), actor.ID, tag, p, counts[0], counts[1])
		if err != nil {
			return nil, err
		}
		mastery[tag] = math.Round(p*10000) / 10000
	}
	sort.Slice(weak, func(i, j int) bool { return model.Number(accuracy, weak[i]) < model.Number(accuracy, weak[j]) })
	if s.Config.KTBackend == "kt_service" {
		payload := model.Record{"user_id": actor.ID, "traces": traces, "prev_mastery": previous}
		predicted, err := s.Clients.KT.Predict(ctx, payload)
		if err == nil {
			for tag, value := range model.Object(predicted["mastery_by_tag"]) {
				n := model.Number(model.Record{"v": value}, "v")
				if n >= 0 && n <= 1 {
					mastery[tag] = n
					if _, err := repository.UpdateKnowledgeMastery(ctx, tx, n, actor.ID, tag); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	result := model.Record{"attempt_id": a["id"], "attempt_status": "submitted", "total_score": 100, "score": math.Round(score*100) / 100, "total_questions": len(qs), "answered_questions": answered, "correct_count": correct, "accuracy": math.Round(float64(correct)/float64(len(qs))*10000) / 100, "weak_knowledge_points": weak, "knowledge_accuracy": accuracy, "knowledge_mastery": mastery, "q_matrix": traces, "q_matrix_by_question": byQuestion, "q_matrix_by_tag": tagQuestions}
	if _, err = repository.Write(ctx, tx, "simulation_attempt", model.String(a, "id"), model.Record{"status": "submitted", "answers_payload": model.JSON(answers), "result_payload": model.JSON(result), "submitted_at": time.Now()}); err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}
func (s *Learning) saveKnowledgeTree(ctx context.Context, actor model.Actor, input dto.LearningInput) (any, error) {
	if err := s.permission(ctx, actor, "Categories"); err != nil {
		return nil, err
	}
	v := input.Fields()
	if input.ID != "" {
		v["attempt_id"] = input.ID
	}
	var err error
	root := model.Object(v["knowledge_tree"])
	if err = require(root, "course_name"); err != nil {
		return nil, err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	course, err := repository.Write(ctx, tx, "categories", "", model.Record{"name": root["course_name"], "level": 1})
	if err != nil {
		return nil, err
	}
	fields := []string{"chapters", "sections", "knowledge_points", "elements"}
	var save func(model.Record, string, int) error
	save = func(node model.Record, parent string, depth int) error {
		if depth >= len(fields) {
			return nil
		}
		for i, child := range model.Objects(node[fields[depth]]) {
			if err := require(child, "name"); err != nil {
				return err
			}
			row, err := repository.Write(ctx, tx, "categories", "", model.Record{"name": child["name"], "description": child["description"], "parent_id": parent, "level": depth + 2, "sort_order": i})
			if err != nil {
				return err
			}
			if err = save(child, model.String(row, "id"), depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err = save(root, model.String(course, "id"), 0); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return model.Record{"success": true, "message": "知识树保存成功", "course_id": course["id"]}, nil
}
func (s *Learning) stream(ctx context.Context, actor model.Actor, input dto.LearningInput, knowledge bool, emit func(model.Record)) (any, error) {
	permission := "Questions"
	if knowledge {
		permission = "Categories"
	}
	if err := s.permission(ctx, actor, permission); err != nil {
		return nil, err
	}
	v := input.Fields()
	if input.ID != "" {
		v["attempt_id"] = input.ID
	}
	var err error
	prompt := "你是专业教师，根据输入知识点生成一道题，仅返回 JSON: title,type(single_choice/multi_choice/true_false/blank/essay),difficulty(1-5),answer,explanation,options(数组，包含 option_key,content,is_correct)。选择题必须标记正确答案。输入：" + model.JSON(v)
	if knowledge {
		prompt = "根据考试大纲生成五层知识树，仅返回 JSON: course_name,total_chapters,chapters(数组，每项 name,sections 数组，每小节 name,knowledge_points 数组，每知识点 name,description,elements 数组，每知识元 name,description)。输入：" + model.JSON(v)
	}
	resp, err := s.Clients.LLM.Request(ctx, prompt, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var content strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if raw == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err = json.Unmarshal([]byte(raw), &chunk); err != nil {
			emit(model.Record{"type": "error", "message": "LLM 流格式无效"})
			return nil, nil
		}
		if len(chunk.Choices) > 0 {
			part := chunk.Choices[0].Delta.Content
			content.WriteString(part)
			emit(model.Record{"type": "chunk", "content": part})
		}
		if content.Len() > 8<<20 {
			emit(model.Record{"type": "error", "message": "生成结果过大"})
			return nil, nil
		}
	}
	if err = scanner.Err(); err != nil {
		emit(model.Record{"type": "error", "message": "LLM 流中断"})
		return nil, nil
	}
	var data model.Record
	if err = json.Unmarshal([]byte(content.String()), &data); err != nil {
		emit(model.Record{"type": "error", "message": "生成结果不是有效 JSON"})
	} else {
		emit(model.Record{"type": "done", "data": data})
	}
	return nil, nil
}
func (s *Learning) PracticeGeneration(ctx context.Context, actor model.Actor, input dto.LearningInput) (any, error) {
	if err := s.permission(ctx, actor, "Practices"); err != nil {
		return nil, err
	}
	v := input.Fields()
	if input.ID != "" {
		v["attempt_id"] = input.ID
	}
	var err error
	qs, short, err := s.selectQuestions(ctx, actor, v)
	return model.Record{"questions": qs, "shortages": short}, err
}
func (s *Learning) Generate(ctx context.Context, actor model.Actor, input dto.LearningInput) (dto.Creation, error) {
	return s.generateSimulation(ctx, actor, input)
}
func (s *Learning) Submit(ctx context.Context, actor model.Actor, input dto.LearningInput) (any, error) {
	return s.submitSimulation(ctx, actor, input)
}
func (s *Learning) Current(ctx context.Context, actor model.Actor) (any, error) {
	rows, err := repository.CurrentSimulationAttempts(ctx, s.DB, actor.ID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return model.Record{"attempt_id": nil, "attempt_status": nil, "questions": []model.Record{}}, nil
	}
	a := rows[0]
	return model.Record{"attempt_id": a["id"], "attempt_status": a["status"], "questions": model.Objects(a["questions_payload"])}, nil
}
func (s *Learning) Review(ctx context.Context, actor model.Actor, id string) (any, error) {
	a, err := repository.SimulationAttemptForUser(ctx, s.DB, id, actor.ID)
	if err != nil {
		return nil, err
	}
	if a["status"] != "submitted" {
		return nil, fail(http.StatusBadRequest, "当前试卷尚未交卷")
	}
	return model.Record{"attempt_id": a["id"], "attempt_status": a["status"], "questions": model.Objects(a["questions_payload"]), "answers": model.Objects(a["answers_payload"]), "result": model.Object(a["result_payload"])}, nil
}
func (s *Learning) Traces(ctx context.Context, actor model.Actor, query url.Values) (any, error) {
	rows, total, err := s.list(ctx, actor, query, "simulation_answer_trace", "user_id=$1", []any{actor.ID})
	return model.Record{"total": total, "items": rows}, err
}
func (s *Learning) KTStatus(ctx context.Context) (any, error) {
	out := model.Record{"backend": s.Config.KTBackend, "available": s.Config.KTBackend == "heuristic", "model": "heuristic"}
	if s.Config.KTBackend == "kt_service" {
		data, status, err := s.Clients.KT.Health(ctx)
		out["available"] = status == http.StatusOK
		if err == nil {
			out["service"] = data
			out["model"] = data["model"]
		}
	}
	return out, nil
}
func (s *Learning) QuestionGeneration(ctx context.Context, actor model.Actor, input dto.LearningInput, emit func(model.Record)) (any, error) {
	return s.stream(ctx, actor, input, false, emit)
}
func (s *Learning) KnowledgeAnalysis(ctx context.Context, actor model.Actor, input dto.LearningInput, emit func(model.Record)) (any, error) {
	return s.stream(ctx, actor, input, true, emit)
}
func (s *Learning) SaveKnowledge(ctx context.Context, actor model.Actor, input dto.LearningInput) (any, error) {
	return s.saveKnowledgeTree(ctx, actor, input)
}

func (s *Learning) selectQuestions(ctx context.Context, actor model.Actor, v model.Record) ([]model.Record, model.Record, error) {
	rows, err := repository.QuestionCandidates(ctx, s.DB, actor, v)
	if err != nil {
		return nil, nil, err
	}
	counts := model.Object(v["type_counts"])
	out := []model.Record{}
	shortages := model.Record{}
	if len(counts) > 0 {
		keys := []string{}
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, kind := range keys {
			need := int(model.Number(counts, kind))
			if need < 0 || need > 50 {
				return nil, nil, fail(http.StatusUnprocessableEntity, "每种题型数量必须为 0-50")
			}
			for _, q := range rows {
				if need > 0 && model.String(q, "type") == bankType(kind) {
					out = append(out, q)
					need--
				}
			}
			if need > 0 {
				shortages[kind] = need
			}
		}
	} else {
		n := int(model.Number(v, "question_count"))
		if n == 0 {
			n = 10
			if model.String(v, "practice_id") != "" {
				n = len(rows)
			}
		}
		if n < 1 || n > 50 {
			return nil, nil, fail(http.StatusUnprocessableEntity, "题目数量必须为 1-50")
		}
		if n > len(rows) {
			n = len(rows)
		}
		out = rows[:n]
	}
	if len(out) > 50 {
		return nil, nil, fail(http.StatusUnprocessableEntity, "试卷最多 50 题")
	}
	for i, q := range out {
		out[i], err = s.enrich(ctx, s.DB, "questions", q)
		if err != nil {
			return nil, nil, err
		}
	}
	return out, shortages, nil
}

func (s *Learning) llm(ctx context.Context, prompt string) (model.Record, error) {
	resp, err := s.Clients.LLM.Request(ctx, prompt, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fail(http.StatusBadGateway, "LLM 返回空结果")
	}
	var v model.Record
	err = json.Unmarshal([]byte(out.Choices[0].Message.Content), &v)
	if err != nil {
		return nil, fail(http.StatusBadGateway, "LLM 返回格式无效")
	}
	return v, nil
}
