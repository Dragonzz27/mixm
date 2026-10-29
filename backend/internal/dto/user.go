package dto

import (
	"mixm/backend/internal/model"
)

// UserInput 定义当前业务接口允许写入的请求字段。
type UserInput struct {
	Email       Optional[string] `json:"email"`
	IsActive    Optional[bool]   `json:"is_active"`
	NewPassword Optional[string] `json:"new_password"`
	Password    Optional[string] `json:"password"`
	Phone       Optional[string] `json:"phone"`
	RoleID      Optional[string] `json:"role_id"`
	Username    Optional[string] `json:"username"`
}

// Fields 提取请求中明确提供的字段，保留空值与零值。
func (input UserInput) Fields() model.Record {
	return requestFields(input)
}
