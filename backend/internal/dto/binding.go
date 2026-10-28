package dto

import (
	"mixm/backend/internal/model"
)

// BindingInput contains the writable fields for this business contract.
type BindingInput struct {
	CategoryID Optional[string]   `json:"category_id"`
	ContentID  Optional[string]   `json:"content_id"`
	ContentIds Optional[[]string] `json:"content_ids"`
	SortOrder  Optional[int]      `json:"sort_order"`
}

func (input BindingInput) Fields() model.Record { return requestFields(input) }
