package model

import (
	"encoding/json"
	"time"
)

type AssignmentRole struct {
	AssignmentID string `gorm:"column:assignment_id;type:text;primaryKey" json:"assignment_id"`
	RoleID       string `gorm:"column:role_id;type:text;primaryKey" json:"role_id"`
}

func (AssignmentRole) TableName() string { return "assignment_roles" }

type Assignment struct {
	ID          string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Title       string     `gorm:"column:title;type:text" json:"title"`
	Description *string    `gorm:"column:description;type:text" json:"description"`
	Attachments *string    `gorm:"column:attachments;type:text" json:"attachments"`
	ContentID   *string    `gorm:"column:content_id;type:text" json:"content_id"`
	Deadline    time.Time  `gorm:"column:deadline;type:timestamptz" json:"deadline"`
	Status      *string    `gorm:"column:status;type:text" json:"status"`
	CreatedBy   string     `gorm:"column:created_by;type:text" json:"created_by"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Assignment) TableName() string { return "assignments" }

type Submission struct {
	ID           string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	AssignmentID string     `gorm:"column:assignment_id;type:text" json:"assignment_id"`
	UserID       string     `gorm:"column:user_id;type:text" json:"user_id"`
	FilePath     *string    `gorm:"column:file_path;type:text" json:"file_path"`
	FileName     *string    `gorm:"column:file_name;type:text" json:"file_name"`
	FileSize     *int       `gorm:"column:file_size;type:integer" json:"file_size"`
	Content      *string    `gorm:"column:content;type:text" json:"content"`
	Status       *string    `gorm:"column:status;type:text" json:"status"`
	Grade        *int       `gorm:"column:grade;type:integer" json:"grade"`
	Feedback     *string    `gorm:"column:feedback;type:text" json:"feedback"`
	SubmittedAt  *time.Time `gorm:"column:submitted_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"submitted_at"`
	ReviewedAt   *time.Time `gorm:"column:reviewed_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"reviewed_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Submission) TableName() string { return "submissions" }

type ImageRepository struct {
	ID        string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name      string     `gorm:"column:name;type:text" json:"name"`
	IsPublic  *bool      `gorm:"column:is_public;type:boolean" json:"is_public"`
	UserID    string     `gorm:"column:user_id;type:text" json:"user_id"`
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (ImageRepository) TableName() string { return "image_repositories" }

type User struct {
	ID             string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Username       string     `gorm:"column:username;type:text" json:"username"`
	Email          *string    `gorm:"column:email;type:text" json:"email"`
	Phone          *string    `gorm:"column:phone;type:text" json:"phone"`
	HashedPassword string     `gorm:"column:hashed_password;type:text" json:"-"`
	IsActive       *bool      `gorm:"column:is_active;type:boolean" json:"is_active"`
	IsSuperuser    *bool      `gorm:"column:is_superuser;type:boolean" json:"is_superuser"`
	CreatedAt      *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (User) TableName() string { return "users" }

type Policy struct {
	ID           string          `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name         string          `gorm:"column:name;type:text" json:"name"`
	Description  *string         `gorm:"column:description;type:text" json:"description"`
	Condition    json.RawMessage `gorm:"column:condition;type:jsonb" json:"condition"`
	Action       string          `gorm:"column:action;type:text" json:"action"`
	ResourceType string          `gorm:"column:resource_type;type:text" json:"resource_type"`
	Effect       *string         `gorm:"column:effect;type:text" json:"effect"`
	Priority     *int            `gorm:"column:priority;type:integer" json:"priority"`
	IsActive     *bool           `gorm:"column:is_active;type:boolean" json:"is_active"`
}

func (Policy) TableName() string { return "policies" }

type PracticeRole struct {
	PracticeID string `gorm:"column:practice_id;type:text;primaryKey" json:"practice_id"`
	RoleID     string `gorm:"column:role_id;type:text;primaryKey" json:"role_id"`
}

func (PracticeRole) TableName() string { return "practice_roles" }

type PracticeTag struct {
	PracticeID string `gorm:"column:practice_id;type:text;primaryKey" json:"practice_id"`
	CategoryID string `gorm:"column:category_id;type:text;primaryKey" json:"category_id"`
}

func (PracticeTag) TableName() string { return "practice_tags" }

type Practice struct {
	ID          string          `gorm:"column:id;type:text;primaryKey" json:"id"`
	Title       string          `gorm:"column:title;type:text" json:"title"`
	Description *string         `gorm:"column:description;type:text" json:"description"`
	TypeCounts  json.RawMessage `gorm:"column:type_counts;type:jsonb" json:"type_counts"`
	CreatedBy   string          `gorm:"column:created_by;type:text" json:"created_by"`
	CreatedAt   *time.Time      `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Practice) TableName() string { return "practices" }

type PracticeQuestion struct {
	ID         string `gorm:"column:id;type:text;primaryKey" json:"id"`
	PracticeID string `gorm:"column:practice_id;type:text" json:"practice_id"`
	QuestionID string `gorm:"column:question_id;type:text" json:"question_id"`
	SortOrder  int    `gorm:"column:sort_order;type:integer" json:"sort_order"`
}

func (PracticeQuestion) TableName() string { return "practice_questions" }

type ContentCategorie struct {
	ID         string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	ContentID  string     `gorm:"column:content_id;type:text" json:"content_id"`
	CategoryID string     `gorm:"column:category_id;type:text" json:"category_id"`
	SortOrder  *int       `gorm:"column:sort_order;type:integer" json:"sort_order"`
	CreatedAt  *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
}

func (ContentCategorie) TableName() string { return "content_categories" }

type ContentType struct {
	ID          string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name        string     `gorm:"column:name;type:text" json:"name"`
	Code        string     `gorm:"column:code;type:text" json:"code"`
	Description *string    `gorm:"column:description;type:text" json:"description"`
	IsActive    *bool      `gorm:"column:is_active;type:boolean" json:"is_active"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (ContentType) TableName() string { return "content_types" }

type QuestionRole struct {
	QuestionID string `gorm:"column:question_id;type:text;primaryKey" json:"question_id"`
	RoleID     string `gorm:"column:role_id;type:text;primaryKey" json:"role_id"`
}

func (QuestionRole) TableName() string { return "question_roles" }

type QuestionTag struct {
	QuestionID string `gorm:"column:question_id;type:text;primaryKey" json:"question_id"`
	CategoryID string `gorm:"column:category_id;type:text;primaryKey" json:"category_id"`
}

func (QuestionTag) TableName() string { return "question_tags" }

type Category struct {
	ID          string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name        string     `gorm:"column:name;type:text" json:"name"`
	Description *string    `gorm:"column:description;type:text" json:"description"`
	ParentID    *string    `gorm:"column:parent_id;type:text" json:"parent_id"`
	Level       *int       `gorm:"column:level;type:integer" json:"level"`
	SortOrder   *int       `gorm:"column:sort_order;type:integer" json:"sort_order"`
	IsActive    *bool      `gorm:"column:is_active;type:boolean" json:"is_active"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Category) TableName() string { return "categories" }

type Content struct {
	ID          string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Title       string     `gorm:"column:title;type:text" json:"title"`
	Content     *string    `gorm:"column:content;type:text" json:"content"`
	ContentType *string    `gorm:"column:content_type;type:text" json:"content_type"`
	Url         *string    `gorm:"column:url;type:text" json:"url"`
	Status      *string    `gorm:"column:status;type:text" json:"status"`
	TemplateID  *string    `gorm:"column:template_id;type:text" json:"template_id"`
	IsActive    *bool      `gorm:"column:is_active;type:boolean" json:"is_active"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Content) TableName() string { return "contents" }

type Instance struct {
	ID             string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name           string     `gorm:"column:name;type:text" json:"name"`
	UserID         string     `gorm:"column:user_id;type:text" json:"user_id"`
	Image          string     `gorm:"column:image;type:text" json:"image"`
	RepoID         *string    `gorm:"column:repo_id;type:text" json:"repo_id"`
	RepoTag        *string    `gorm:"column:repo_tag;type:text" json:"repo_tag"`
	Status         *string    `gorm:"column:status;type:text" json:"status"`
	Namespace      *string    `gorm:"column:namespace;type:text" json:"namespace"`
	Replicas       *int       `gorm:"column:replicas;type:integer" json:"replicas"`
	CpuLimit       *float64   `gorm:"column:cpu_limit;type:double precision" json:"cpu_limit"`
	MemoryLimit    *string    `gorm:"column:memory_limit;type:text" json:"memory_limit"`
	MemoryLimitMb  *int       `gorm:"column:memory_limit_mb;type:integer" json:"memory_limit_mb"`
	StorageLimit   *int       `gorm:"column:storage_limit;type:integer" json:"storage_limit"`
	GpuLimit       *int       `gorm:"column:gpu_limit;type:integer" json:"gpu_limit"`
	GpuMemoryLimit *int       `gorm:"column:gpu_memory_limit;type:integer" json:"gpu_memory_limit"`
	CpuUsage       *int       `gorm:"column:cpu_usage;type:integer" json:"cpu_usage"`
	MemoryUsage    *int       `gorm:"column:memory_usage;type:integer" json:"memory_usage"`
	SSHPort        *int       `gorm:"column:ssh_port;type:integer" json:"ssh_port"`
	SSHUsername    *string    `gorm:"column:ssh_username;type:text" json:"ssh_username"`
	Config         *string    `gorm:"column:config;type:text" json:"config"`
	StartedAt      *time.Time `gorm:"column:started_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"started_at"`
	CreatedAt      *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Instance) TableName() string { return "instances" }

type UserRole struct {
	ID     string `gorm:"column:id;type:text;primaryKey" json:"id"`
	UserID string `gorm:"column:user_id;type:text" json:"user_id"`
	RoleID string `gorm:"column:role_id;type:text" json:"role_id"`
}

func (UserRole) TableName() string { return "user_roles" }

type Template struct {
	ID          string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name        string     `gorm:"column:name;type:text" json:"name"`
	Description *string    `gorm:"column:description;type:text" json:"description"`
	Config      *string    `gorm:"column:config;type:text" json:"config"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Template) TableName() string { return "templates" }

type Role struct {
	ID        string          `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name      string          `gorm:"column:name;type:text" json:"name"`
	RouteKeys json.RawMessage `gorm:"column:route_keys;type:jsonb" json:"route_keys"`
	CreatedAt *time.Time      `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt *time.Time      `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Role) TableName() string { return "roles" }

type SimulationAttempt struct {
	ID               string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	UserID           string     `gorm:"column:user_id;type:text" json:"user_id"`
	Status           string     `gorm:"column:status;type:text" json:"status"`
	PaperSource      string     `gorm:"column:paper_source;type:text" json:"paper_source"`
	PracticeID       *string    `gorm:"column:practice_id;type:text" json:"practice_id"`
	QuestionsPayload string     `gorm:"column:questions_payload;type:text" json:"questions_payload"`
	AnswersPayload   string     `gorm:"column:answers_payload;type:text" json:"answers_payload"`
	ResultPayload    *string    `gorm:"column:result_payload;type:text" json:"result_payload"`
	StartedAt        *time.Time `gorm:"column:started_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"started_at"`
	SubmittedAt      *time.Time `gorm:"column:submitted_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"submitted_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (SimulationAttempt) TableName() string { return "simulation_attempt" }

type Question struct {
	ID            string          `gorm:"column:id;type:text;primaryKey" json:"id"`
	Title         string          `gorm:"column:title;type:text" json:"title"`
	Type          string          `gorm:"column:type;type:text" json:"type"`
	Content       *string         `gorm:"column:content;type:text" json:"content"`
	Answer        *string         `gorm:"column:answer;type:text" json:"answer"`
	Explanation   *string         `gorm:"column:explanation;type:text" json:"explanation"`
	Difficulty    int             `gorm:"column:difficulty;type:integer" json:"difficulty"`
	Tags          json.RawMessage `gorm:"column:tags;type:jsonb" json:"tags"`
	Source        *string         `gorm:"column:source;type:text" json:"source"`
	IsAiGenerated *bool           `gorm:"column:is_ai_generated;type:boolean" json:"is_ai_generated"`
	CreatedBy     *string         `gorm:"column:created_by;type:text" json:"created_by"`
	ReviewStatus  string          `gorm:"column:review_status;type:text" json:"review_status"`
	ReviewedBy    *string         `gorm:"column:reviewed_by;type:text" json:"reviewed_by"`
	ReviewedAt    *time.Time      `gorm:"column:reviewed_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"reviewed_at"`
	CreatedAt     *time.Time      `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	UpdatedAt     *time.Time      `gorm:"column:updated_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"updated_at"`
}

func (Question) TableName() string { return "questions" }

type QuestionCategory struct {
	ID        string     `gorm:"column:id;type:text;primaryKey" json:"id"`
	Name      string     `gorm:"column:name;type:text" json:"name"`
	ParentID  *string    `gorm:"column:parent_id;type:text" json:"parent_id"`
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
}

func (QuestionCategory) TableName() string { return "question_categories" }

type QuestionOption struct {
	ID         string `gorm:"column:id;type:text;primaryKey" json:"id"`
	QuestionID string `gorm:"column:question_id;type:text" json:"question_id"`
	OptionKey  string `gorm:"column:option_key;type:text" json:"option_key"`
	Content    string `gorm:"column:content;type:text" json:"content"`
	IsCorrect  *bool  `gorm:"column:is_correct;type:boolean" json:"is_correct"`
}

func (QuestionOption) TableName() string { return "question_options" }

type QuestionAttachment struct {
	ID         string  `gorm:"column:id;type:text;primaryKey" json:"id"`
	QuestionID string  `gorm:"column:question_id;type:text" json:"question_id"`
	FilePath   *string `gorm:"column:file_path;type:text" json:"file_path"`
	FileName   *string `gorm:"column:file_name;type:text" json:"file_name"`
	FileType   *string `gorm:"column:file_type;type:text" json:"file_type"`
}

func (QuestionAttachment) TableName() string { return "question_attachments" }

type InstanceTask struct {
	ID          string          `gorm:"column:id;type:text;primaryKey" json:"id"`
	UserID      string          `gorm:"column:user_id;type:text" json:"user_id"`
	Payload     json.RawMessage `gorm:"column:payload;type:jsonb" json:"payload"`
	Status      string          `gorm:"column:status;type:text" json:"status"`
	InstanceID  *string         `gorm:"column:instance_id;type:text" json:"instance_id"`
	Error       *string         `gorm:"column:error;type:text" json:"error"`
	Attempts    int             `gorm:"column:attempts;type:integer" json:"attempts"`
	CreatedAt   time.Time       `gorm:"column:created_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"created_at"`
	StartedAt   *time.Time      `gorm:"column:started_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"started_at"`
	CompletedAt *time.Time      `gorm:"column:completed_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"completed_at"`
	PublishedAt *time.Time      `gorm:"column:published_at;type:timestamptz;autoCreateTime:false;autoUpdateTime:false" json:"published_at"`
}

func (InstanceTask) TableName() string { return "instance_tasks" }

func Entity(table string) any {
	switch table {
	case "assignment_roles":
		return &AssignmentRole{}
	case "assignments":
		return &Assignment{}
	case "submissions":
		return &Submission{}
	case "image_repositories":
		return &ImageRepository{}
	case "users":
		return &User{}
	case "policies":
		return &Policy{}
	case "practice_roles":
		return &PracticeRole{}
	case "practice_tags":
		return &PracticeTag{}
	case "practices":
		return &Practice{}
	case "practice_questions":
		return &PracticeQuestion{}
	case "content_categories":
		return &ContentCategorie{}
	case "content_types":
		return &ContentType{}
	case "question_roles":
		return &QuestionRole{}
	case "question_tags":
		return &QuestionTag{}
	case "categories":
		return &Category{}
	case "contents":
		return &Content{}
	case "instances":
		return &Instance{}
	case "user_roles":
		return &UserRole{}
	case "templates":
		return &Template{}
	case "roles":
		return &Role{}
	case "simulation_attempt":
		return &SimulationAttempt{}
	case "questions":
		return &Question{}
	case "question_categories":
		return &QuestionCategory{}
	case "question_options":
		return &QuestionOption{}
	case "question_attachments":
		return &QuestionAttachment{}
	case "instance_tasks":
		return &InstanceTask{}
	}
	return nil
}
