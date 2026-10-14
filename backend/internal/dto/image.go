package dto

import (
	"mixm-edu/backend/internal/model"
)

// ImageInput contains the writable fields for this business contract.
type ImageInput struct {
	IsPublic Optional[bool]   `json:"is_public"`
	Name     Optional[string] `json:"name"`
	UserID   Optional[string] `json:"user_id"`
}

func (input ImageInput) Fields() model.Record { return requestFields(input) }
