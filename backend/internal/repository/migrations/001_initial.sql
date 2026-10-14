CREATE TABLE "assignment_roles" (
  "assignment_id" text NOT NULL,
  "role_id" text NOT NULL,
  PRIMARY KEY ("assignment_id", "role_id")
);

CREATE TABLE "assignments" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "description" text,
  "attachments" text,
  "content_id" text,
  "deadline" timestamptz NOT NULL,
  "status" text DEFAULT 'active',
  "created_by" text NOT NULL,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "submissions" (
  "id" text NOT NULL,
  "assignment_id" text NOT NULL,
  "user_id" text NOT NULL,
  "file_path" text,
  "file_name" text,
  "file_size" integer,
  "content" text,
  "status" text DEFAULT 'submitted',
  "grade" integer,
  "feedback" text,
  "submitted_at" timestamptz DEFAULT now(),
  "reviewed_at" timestamptz,
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "image_repositories" (
  "id" text NOT NULL,
  "name" text NOT NULL UNIQUE,
  "is_public" boolean DEFAULT false,
  "user_id" text NOT NULL,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "users" (
  "id" text NOT NULL,
  "username" text NOT NULL UNIQUE,
  "email" text UNIQUE,
  "phone" text UNIQUE,
  "hashed_password" text NOT NULL,
  "is_active" boolean DEFAULT true,
  "is_superuser" boolean DEFAULT false,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "policies" (
  "id" text NOT NULL,
  "name" text NOT NULL UNIQUE,
  "description" text,
  "condition" jsonb NOT NULL,
  "action" text NOT NULL,
  "resource_type" text NOT NULL,
  "effect" text,
  "priority" integer DEFAULT 0,
  "is_active" boolean DEFAULT true,
  PRIMARY KEY ("id")
);

CREATE TABLE "practice_roles" (
  "practice_id" text NOT NULL,
  "role_id" text NOT NULL,
  PRIMARY KEY ("practice_id", "role_id")
);

CREATE TABLE "practice_tags" (
  "practice_id" text NOT NULL,
  "category_id" text NOT NULL,
  PRIMARY KEY ("practice_id", "category_id")
);

CREATE TABLE "practices" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "description" text,
  "type_counts" jsonb,
  "created_by" text NOT NULL,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "practice_questions" (
  "id" text NOT NULL,
  "practice_id" text NOT NULL,
  "question_id" text NOT NULL,
  "sort_order" integer NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);

CREATE TABLE "content_categories" (
  "id" text NOT NULL,
  "content_id" text NOT NULL,
  "category_id" text NOT NULL,
  "sort_order" integer DEFAULT 0,
  "created_at" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);

CREATE TABLE "content_types" (
  "id" text NOT NULL,
  "name" text NOT NULL UNIQUE,
  "code" text NOT NULL UNIQUE,
  "description" text,
  "is_active" boolean DEFAULT true,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "question_roles" (
  "question_id" text NOT NULL,
  "role_id" text NOT NULL,
  PRIMARY KEY ("question_id", "role_id")
);

CREATE TABLE "question_tags" (
  "question_id" text NOT NULL,
  "category_id" text NOT NULL,
  PRIMARY KEY ("question_id", "category_id")
);

CREATE TABLE "categories" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "description" text,
  "parent_id" text,
  "level" integer DEFAULT 1,
  "sort_order" integer DEFAULT 0,
  "is_active" boolean DEFAULT true,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "contents" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "content" text,
  "content_type" text DEFAULT 'article',
  "url" text,
  "status" text DEFAULT 'draft',
  "template_id" text,
  "is_active" boolean DEFAULT true,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "instances" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "user_id" text NOT NULL,
  "image" text NOT NULL DEFAULT 'ubuntu:24.04',
  "repo_id" text,
  "repo_tag" text,
  "status" text DEFAULT 'pending',
  "namespace" text DEFAULT 'default',
  "replicas" integer DEFAULT 1,
  "cpu_limit" double precision DEFAULT 1.0,
  "memory_limit" text DEFAULT '1Gi',
  "memory_limit_mb" integer DEFAULT 1024,
  "storage_limit" integer DEFAULT 1,
  "gpu_limit" integer DEFAULT 0,
  "gpu_memory_limit" integer DEFAULT 0,
  "cpu_usage" integer DEFAULT 0,
  "memory_usage" integer DEFAULT 0,
  "ssh_port" integer,
  "ssh_username" text,
  "config" text,
  "started_at" timestamptz,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "user_roles" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "role_id" text NOT NULL,
  PRIMARY KEY ("id")
);

CREATE TABLE "templates" (
  "id" text NOT NULL,
  "name" text NOT NULL UNIQUE,
  "description" text,
  "config" text,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "roles" (
  "id" text NOT NULL,
  "name" text NOT NULL UNIQUE,
  "route_keys" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "knowledge_mastery" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "knowledge_tag" text NOT NULL,
  "mastery" double precision NOT NULL DEFAULT 0.5,
  "attempts" integer NOT NULL DEFAULT 0,
  "correct_attempts" integer NOT NULL DEFAULT 0,
  "updated_at" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);

CREATE TABLE "simulation_answer_trace" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "question_id" text NOT NULL,
  "knowledge_tag" text NOT NULL,
  "selected_index" integer,
  "correct_index" integer NOT NULL,
  "is_correct" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);

CREATE TABLE "simulation_attempt" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "status" text NOT NULL DEFAULT 'in_progress',
  "paper_source" text NOT NULL DEFAULT 'random',
  "practice_id" text,
  "questions_payload" text NOT NULL DEFAULT '[]',
  "answers_payload" text NOT NULL DEFAULT '[]',
  "result_payload" text,
  "started_at" timestamptz DEFAULT now(),
  "submitted_at" timestamptz,
  "updated_at" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);

CREATE TABLE "questions" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "type" text NOT NULL,
  "content" text,
  "answer" text,
  "explanation" text,
  "difficulty" integer NOT NULL,
  "tags" jsonb,
  "source" text,
  "is_ai_generated" boolean DEFAULT false,
  "created_by" text,
  "review_status" text NOT NULL DEFAULT 'pending',
  "reviewed_by" text,
  "reviewed_at" timestamptz,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE TABLE "question_categories" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "parent_id" text,
  "created_at" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);

CREATE TABLE "question_options" (
  "id" text NOT NULL,
  "question_id" text NOT NULL,
  "option_key" text NOT NULL,
  "content" text NOT NULL,
  "is_correct" boolean DEFAULT false,
  PRIMARY KEY ("id")
);

CREATE TABLE "question_attachments" (
  "id" text NOT NULL,
  "question_id" text NOT NULL,
  "file_path" text,
  "file_name" text,
  "file_type" text,
  PRIMARY KEY ("id")
);

ALTER TABLE "assignment_roles" ADD FOREIGN KEY ("assignment_id") REFERENCES "assignments" ("id") ON DELETE CASCADE;

ALTER TABLE "assignment_roles" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE;

ALTER TABLE "assignments" ADD FOREIGN KEY ("content_id") REFERENCES "contents" ("id") ON DELETE SET NULL;

ALTER TABLE "submissions" ADD FOREIGN KEY ("assignment_id") REFERENCES "assignments" ("id") ON DELETE CASCADE;

ALTER TABLE "submissions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "image_repositories" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "practice_roles" ADD FOREIGN KEY ("practice_id") REFERENCES "practices" ("id") ON DELETE CASCADE;

ALTER TABLE "practice_roles" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE;

ALTER TABLE "practice_tags" ADD FOREIGN KEY ("practice_id") REFERENCES "practices" ("id") ON DELETE CASCADE;

ALTER TABLE "practice_tags" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE CASCADE;

ALTER TABLE "practice_questions" ADD FOREIGN KEY ("practice_id") REFERENCES "practices" ("id") ON DELETE CASCADE;

ALTER TABLE "practice_questions" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE;

ALTER TABLE "content_categories" ADD FOREIGN KEY ("content_id") REFERENCES "contents" ("id") ON DELETE SET NULL;

ALTER TABLE "content_categories" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE CASCADE;

ALTER TABLE "question_roles" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE;

ALTER TABLE "question_roles" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE;

ALTER TABLE "question_tags" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE;

ALTER TABLE "question_tags" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE CASCADE;

ALTER TABLE "categories" ADD FOREIGN KEY ("parent_id") REFERENCES "categories" ("id") ON DELETE SET NULL;

ALTER TABLE "contents" ADD FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON DELETE SET NULL;

ALTER TABLE "instances" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "instances" ADD FOREIGN KEY ("repo_id") REFERENCES "image_repositories" ("id") ON DELETE SET NULL;

ALTER TABLE "user_roles" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "user_roles" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE;

ALTER TABLE "question_categories" ADD FOREIGN KEY ("parent_id") REFERENCES "question_categories" ("id") ON DELETE SET NULL;

ALTER TABLE "question_options" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE;

ALTER TABLE "question_attachments" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE;

CREATE INDEX "idx_assignments_id" ON "assignments" ("id");

CREATE INDEX "idx_assignments_title" ON "assignments" ("title");

CREATE INDEX "idx_assignments_status" ON "assignments" ("status");

CREATE INDEX "idx_submissions_id" ON "submissions" ("id");

CREATE INDEX "idx_submissions_status" ON "submissions" ("status");

CREATE INDEX "idx_image_repositories_id" ON "image_repositories" ("id");

CREATE INDEX "idx_image_repositories_name" ON "image_repositories" ("name");

CREATE INDEX "idx_users_id" ON "users" ("id");

CREATE INDEX "idx_users_username" ON "users" ("username");

CREATE INDEX "idx_users_email" ON "users" ("email");

CREATE INDEX "idx_users_phone" ON "users" ("phone");

CREATE INDEX "idx_policies_id" ON "policies" ("id");

CREATE INDEX "idx_policies_resource_type" ON "policies" ("resource_type");

CREATE INDEX "idx_practices_id" ON "practices" ("id");

CREATE INDEX "idx_practices_title" ON "practices" ("title");

CREATE INDEX "idx_practice_questions_id" ON "practice_questions" ("id");

CREATE INDEX "idx_practice_questions_practice_id" ON "practice_questions" ("practice_id");

CREATE INDEX "idx_practice_questions_question_id" ON "practice_questions" ("question_id");

CREATE INDEX "idx_content_categories_id" ON "content_categories" ("id");

CREATE INDEX "idx_content_categories_content_id" ON "content_categories" ("content_id");

CREATE INDEX "idx_content_categories_category_id" ON "content_categories" ("category_id");

CREATE INDEX "idx_content_types_id" ON "content_types" ("id");

CREATE INDEX "idx_content_types_name" ON "content_types" ("name");

CREATE INDEX "idx_content_types_code" ON "content_types" ("code");

CREATE INDEX "idx_categories_id" ON "categories" ("id");

CREATE INDEX "idx_categories_name" ON "categories" ("name");

CREATE INDEX "idx_categories_parent_id" ON "categories" ("parent_id");

CREATE INDEX "idx_categories_level" ON "categories" ("level");

CREATE INDEX "idx_contents_id" ON "contents" ("id");

CREATE INDEX "idx_contents_title" ON "contents" ("title");

CREATE INDEX "idx_contents_status" ON "contents" ("status");

CREATE INDEX "idx_instances_id" ON "instances" ("id");

CREATE INDEX "idx_user_roles_id" ON "user_roles" ("id");

CREATE INDEX "idx_user_roles_user_id" ON "user_roles" ("user_id");

CREATE INDEX "idx_user_roles_role_id" ON "user_roles" ("role_id");

CREATE INDEX "idx_templates_id" ON "templates" ("id");

CREATE INDEX "idx_templates_name" ON "templates" ("name");

CREATE INDEX "idx_roles_id" ON "roles" ("id");

CREATE INDEX "idx_roles_name" ON "roles" ("name");

CREATE INDEX "idx_knowledge_mastery_user_id" ON "knowledge_mastery" ("user_id");

CREATE INDEX "idx_knowledge_mastery_knowledge_tag" ON "knowledge_mastery" ("knowledge_tag");

CREATE INDEX "idx_simulation_answer_trace_user_id" ON "simulation_answer_trace" ("user_id");

CREATE INDEX "idx_simulation_answer_trace_question_id" ON "simulation_answer_trace" ("question_id");

CREATE INDEX "idx_simulation_answer_trace_knowledge_tag" ON "simulation_answer_trace" ("knowledge_tag");

CREATE INDEX "idx_simulation_attempt_user_id" ON "simulation_attempt" ("user_id");

CREATE INDEX "idx_simulation_attempt_status" ON "simulation_attempt" ("status");

CREATE INDEX "idx_questions_id" ON "questions" ("id");

CREATE INDEX "idx_questions_title" ON "questions" ("title");

CREATE INDEX "idx_questions_type" ON "questions" ("type");

CREATE INDEX "idx_questions_review_status" ON "questions" ("review_status");

CREATE INDEX "idx_question_categories_id" ON "question_categories" ("id");

CREATE INDEX "idx_question_options_id" ON "question_options" ("id");

CREATE INDEX "idx_question_attachments_id" ON "question_attachments" ("id");

ALTER TABLE user_roles ADD CONSTRAINT uq_user_single_role UNIQUE(user_id);

ALTER TABLE submissions ADD CONSTRAINT uq_submission UNIQUE(assignment_id,user_id);

ALTER TABLE knowledge_mastery ADD CONSTRAINT uq_mastery UNIQUE(user_id,knowledge_tag);

ALTER TABLE content_categories ADD CONSTRAINT uq_binding UNIQUE(content_id,category_id);

ALTER TABLE questions ADD CHECK (difficulty BETWEEN 1 AND 5), ADD CHECK (type IN ('single_choice','multi_choice','true_false','blank','essay')), ADD CHECK (review_status IN ('pending','approved','rejected'));

ALTER TABLE submissions ADD CHECK (grade BETWEEN 0 AND 100);

CREATE TABLE instance_tasks (id text PRIMARY KEY, user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE, payload jsonb NOT NULL, status text NOT NULL DEFAULT 'pending', instance_id text UNIQUE REFERENCES instances(id) ON DELETE SET NULL, error text, attempts integer NOT NULL DEFAULT 0, created_at timestamptz NOT NULL DEFAULT now(), started_at timestamptz, completed_at timestamptz, published_at timestamptz);

CREATE INDEX idx_tasks_user ON instance_tasks(user_id,created_at);
