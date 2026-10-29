package dto

// QuestionOptionInput 描述题目选项的请求字段，与持久化实体分别维护。
type QuestionOptionInput struct {
	OptionKey string `json:"option_key"`
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
}

type SimulationAnswerInput struct {
	QuestionID    string `json:"question_id"`
	SelectedIndex *int   `json:"selected_index"`
	AnswerText    string `json:"answer_text"`
}
