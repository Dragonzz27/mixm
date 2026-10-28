package dto

import (
	"encoding/json"

	"mixm/backend/internal/model"
)

// ResourceInput contains the writable fields for this business contract.
type ResourceInput struct {
	Answer          Optional[string]                `json:"answer"`
	Attachments     Optional[json.RawMessage]       `json:"attachments"`
	Code            Optional[string]                `json:"code"`
	Config          Optional[json.RawMessage]       `json:"config"`
	Content         Optional[string]                `json:"content"`
	ContentID       Optional[string]                `json:"content_id"`
	ContentType     Optional[string]                `json:"content_type"`
	CreatedBy       Optional[string]                `json:"created_by"`
	Deadline        Optional[string]                `json:"deadline"`
	Description     Optional[string]                `json:"description"`
	Difficulty      Optional[int]                   `json:"difficulty"`
	Explanation     Optional[string]                `json:"explanation"`
	IsActive        Optional[bool]                  `json:"is_active"`
	IsAiGenerated   Optional[bool]                  `json:"is_ai_generated"`
	KnowledgeTagIds Optional[[]string]              `json:"knowledge_tag_ids"`
	Level           Optional[int]                   `json:"level"`
	Name            Optional[string]                `json:"name"`
	Options         Optional[[]QuestionOptionInput] `json:"options"`
	ParentID        Optional[string]                `json:"parent_id"`
	QuestionIds     Optional[[]string]              `json:"question_ids"`
	ReviewStatus    Optional[string]                `json:"review_status"`
	RouteKeys       Optional[[]string]              `json:"route_keys"`
	SortOrder       Optional[int]                   `json:"sort_order"`
	Source          Optional[string]                `json:"source"`
	Status          Optional[string]                `json:"status"`
	Tags            Optional[json.RawMessage]       `json:"tags"`
	TemplateID      Optional[string]                `json:"template_id"`
	Title           Optional[string]                `json:"title"`
	Type            Optional[string]                `json:"type"`
	TypeCounts      Optional[map[string]int]        `json:"type_counts"`
	Url             Optional[string]                `json:"url"`
	VisibleRoles    Optional[[]string]              `json:"visible_roles"`
}

func (input ResourceInput) Fields() model.Record { return requestFields(input) }
