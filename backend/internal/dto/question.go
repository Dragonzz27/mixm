package dto

// QuestionOptionInput is distinct from the persisted option entity.
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
