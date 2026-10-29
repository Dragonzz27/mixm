package dto

import (
	"mixm/backend/internal/model"
)

// ImageInput 定义当前业务接口允许写入的请求字段。
type ImageInput struct {
	IsPublic Optional[bool]   `json:"is_public"`
	Name     Optional[string] `json:"name"`
	UserID   Optional[string] `json:"user_id"`
}

// Fields 提取请求中明确提供的字段，保留空值与零值。
func (input ImageInput) Fields() model.Record {
	return requestFields(input)
}
