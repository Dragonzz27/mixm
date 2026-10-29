package dto

import (
	"mixm/backend/internal/model"
)

// BindingInput 定义当前业务接口允许写入的请求字段。
type BindingInput struct {
	CategoryID Optional[string]   `json:"category_id"`
	ContentID  Optional[string]   `json:"content_id"`
	ContentIds Optional[[]string] `json:"content_ids"`
	SortOrder  Optional[int]      `json:"sort_order"`
}

// Fields 提取请求中明确提供的字段，保留空值与零值。
func (input BindingInput) Fields() model.Record {
	return requestFields(input)
}
