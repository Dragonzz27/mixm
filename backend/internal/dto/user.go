package dto

import (
	"mixm/backend/internal/model"
)

// UserInput contains the writable fields for this business contract.
type UserInput struct {
	Email       Optional[string] `json:"email"`
	IsActive    Optional[bool]   `json:"is_active"`
	NewPassword Optional[string] `json:"new_password"`
	Password    Optional[string] `json:"password"`
	Phone       Optional[string] `json:"phone"`
	RoleID      Optional[string] `json:"role_id"`
	Username    Optional[string] `json:"username"`
}

func (input UserInput) Fields() model.Record { return requestFields(input) }
