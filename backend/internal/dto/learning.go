package dto

import (
	"encoding/json"

	"mixm/backend/internal/model"
)

// LearningInput 定义当前业务接口允许写入的请求字段。
type LearningInput struct {
	ID               string                            `json:"-"`
	Answers          Optional[[]SimulationAnswerInput] `json:"answers"`
	CourseID         Optional[string]                  `json:"course_id"`
	CourseName       Optional[string]                  `json:"course_name"`
	Difficulty       Optional[int]                     `json:"difficulty"`
	GenerationMode   Optional[string]                  `json:"generation_mode"`
	KnowledgeItems   Optional[json.RawMessage]         `json:"knowledge_items"`
	KnowledgeTagIds  Optional[[]string]                `json:"knowledge_tag_ids"`
	KnowledgeTree    Optional[json.RawMessage]         `json:"knowledge_tree"`
	OutlineText      Optional[string]                  `json:"outline_text"`
	PracticeID       Optional[string]                  `json:"practice_id"`
	QuestionCount    Optional[int]                     `json:"question_count"`
	QuestionType     Optional[string]                  `json:"question_type"`
	QuestionTypes    Optional[[]string]                `json:"question_types"`
	TimeLimitMinutes Optional[int]                     `json:"time_limit_minutes"`
	TypeCounts       Optional[map[string]int]          `json:"type_counts"`
}

// Fields 提取请求中明确提供的字段，保留空值与零值。
func (input LearningInput) Fields() model.Record {
	return requestFields(input)
}
