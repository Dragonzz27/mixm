package dto

import (
	"encoding/json"

	"mixm/backend/internal/model"
)

// InstanceInput 定义当前业务接口允许写入的请求字段。
type InstanceInput struct {
	Config         Optional[json.RawMessage] `json:"config"`
	CpuLimit       Optional[json.RawMessage] `json:"cpu_limit"`
	CpuUsage       Optional[int]             `json:"cpu_usage"`
	GpuLimit       Optional[json.RawMessage] `json:"gpu_limit"`
	GpuMemoryLimit Optional[json.RawMessage] `json:"gpu_memory_limit"`
	Image          Optional[string]          `json:"image"`
	MemoryLimit    Optional[string]          `json:"memory_limit"`
	MemoryLimitMb  Optional[int]             `json:"memory_limit_mb"`
	MemoryUsage    Optional[int]             `json:"memory_usage"`
	Name           Optional[string]          `json:"name"`
	Namespace      Optional[string]          `json:"namespace"`
	Replicas       Optional[int]             `json:"replicas"`
	RepoID         Optional[string]          `json:"repo_id"`
	RepoTag        Optional[string]          `json:"repo_tag"`
	SshPort        Optional[int]             `json:"ssh_port"`
	SshUsername    Optional[string]          `json:"ssh_username"`
	Status         Optional[string]          `json:"status"`
	StorageLimit   Optional[json.RawMessage] `json:"storage_limit"`
	UserID         Optional[string]          `json:"user_id"`
	UserIds        Optional[[]string]        `json:"user_ids"`
}

// Fields 提取请求中明确提供的字段，保留空值与零值。
func (input InstanceInput) Fields() model.Record {
	return requestFields(input)
}
