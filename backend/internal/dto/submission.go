package dto

import (
	"mixm/backend/internal/model"
)

// SubmissionInput contains the writable fields for this business contract.
type SubmissionInput struct {
	Content  Optional[string] `json:"content"`
	Feedback Optional[string] `json:"feedback"`
	FileName Optional[string] `json:"file_name"`
	FilePath Optional[string] `json:"file_path"`
	FileSize Optional[int]    `json:"file_size"`
	Grade    Optional[int]    `json:"grade"`
	Status   Optional[string] `json:"status"`
}

func (input SubmissionInput) Fields() model.Record { return requestFields(input) }
