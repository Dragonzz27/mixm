package dto

import (
	"mixm/backend/internal/model"
)

// SubmissionInput 定义当前业务接口允许写入的请求字段。
type SubmissionInput struct {
	Content  Optional[string] `json:"content"`
	Feedback Optional[string] `json:"feedback"`
	FileName Optional[string] `json:"file_name"`
	FilePath Optional[string] `json:"file_path"`
	FileSize Optional[int]    `json:"file_size"`
	Grade    Optional[int]    `json:"grade"`
	Status   Optional[string] `json:"status"`
}

// Fields 提取请求中明确提供的字段，保留空值与零值。
func (input SubmissionInput) Fields() model.Record {
	return requestFields(input)
}
