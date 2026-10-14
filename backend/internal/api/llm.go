package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mixm-edu/backend/internal/store"
)

func (s *Server) llmRequest(r *http.Request, prompt string, stream bool) (*http.Response, error) {
	if s.Config.LLMKey == "" {
		return nil, fail(503, "LLM_API_KEY 未配置")
	}
	payload := store.Row{"model": s.Config.LLMModel, "messages": []store.Row{{"role": "user", "content": prompt}}, "response_format": store.Row{"type": "json_object"}, "stream": stream}
	b, e := json.Marshal(payload)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(r.Context(), "POST", strings.TrimSuffix(s.Config.LLMURL, "/")+"/chat/completions", bytes.NewReader(b))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+s.Config.LLMKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := s.do(req, 5*time.Minute)
	if e != nil {
		return nil, e
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fail(502, "LLM 请求失败")
	}
	return resp, nil
}
func (s *Server) llm(r *http.Request, prompt string) (store.Row, error) {
	resp, e := s.llmRequest(r, prompt, false)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if e = json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&out); e != nil {
		return nil, e
	}
	if len(out.Choices) == 0 {
		return nil, fail(502, "LLM 返回空结果")
	}
	var v store.Row
	e = json.Unmarshal([]byte(out.Choices[0].Message.Content), &v)
	if e != nil {
		return nil, fail(502, "LLM 返回格式无效")
	}
	return v, nil
}
func (s *Server) llmStream(knowledge bool) handler {
	return func(w http.ResponseWriter, r *http.Request) (any, error) {
		permission := "Questions"
		if knowledge {
			permission = "Categories"
		}
		if e := s.permission(r, permission); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		prompt := "你是专业教师，根据输入知识点生成一道题，仅返回 JSON: title,type(single_choice/multi_choice/true_false/blank/essay),difficulty(1-5),answer,explanation,options(数组，包含 option_key,content,is_correct)。选择题必须标记正确答案。输入：" + store.JSON(v)
		if knowledge {
			prompt = "根据考试大纲生成五层知识树，仅返回 JSON: course_name,total_chapters,chapters(数组，每项 name,sections 数组，每小节 name,knowledge_points 数组，每知识点 name,description,elements 数组，每知识元 name,description)。输入：" + store.JSON(v)
		}
		resp, e := s.llmRequest(r, prompt, true)
		if e != nil {
			return nil, e
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return nil, fail(500, "流式响应不可用")
		}
		emit := func(x store.Row) { fmt.Fprintf(w, "data: %s\n\n", store.JSON(x)); flusher.Flush() }
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		var content strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if raw == "[DONE]" {
				break
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if e = json.Unmarshal([]byte(raw), &chunk); e != nil {
				emit(store.Row{"type": "error", "message": "LLM 流格式无效"})
				return nil, nil
			}
			if len(chunk.Choices) > 0 {
				part := chunk.Choices[0].Delta.Content
				content.WriteString(part)
				emit(store.Row{"type": "chunk", "content": part})
			}
			if content.Len() > 8<<20 {
				emit(store.Row{"type": "error", "message": "生成结果过大"})
				return nil, nil
			}
		}
		if e = scanner.Err(); e != nil {
			emit(store.Row{"type": "error", "message": "LLM 流中断"})
			return nil, nil
		}
		var data store.Row
		if e = json.Unmarshal([]byte(content.String()), &data); e != nil {
			emit(store.Row{"type": "error", "message": "生成结果不是有效 JSON"})
		} else {
			emit(store.Row{"type": "done", "data": data})
		}
		return nil, nil
	}
}
func (s *Server) saveKnowledgeTree(w http.ResponseWriter, r *http.Request) (any, error) {
	if e := s.permission(r, "Categories"); e != nil {
		return nil, e
	}
	v, e := body(r)
	if e != nil {
		return nil, e
	}
	root := store.Object(v["knowledge_tree"])
	if e = require(root, "course_name"); e != nil {
		return nil, e
	}
	tx, e := s.DB.Begin(r.Context())
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(r.Context())
	course, e := store.Write(r.Context(), tx, "categories", "", store.Row{"name": root["course_name"], "level": 1})
	if e != nil {
		return nil, e
	}
	fields := []string{"chapters", "sections", "knowledge_points", "elements"}
	var save func(store.Row, string, int) error
	save = func(node store.Row, parent string, depth int) error {
		if depth >= len(fields) {
			return nil
		}
		for i, child := range store.Objects(node[fields[depth]]) {
			if e := require(child, "name"); e != nil {
				return e
			}
			row, e := store.Write(r.Context(), tx, "categories", "", store.Row{"name": child["name"], "description": child["description"], "parent_id": parent, "level": depth + 2, "sort_order": i})
			if e != nil {
				return e
			}
			if e = save(child, store.String(row, "id"), depth+1); e != nil {
				return e
			}
		}
		return nil
	}
	if e = save(root, store.String(course, "id"), 0); e != nil {
		return nil, e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return nil, e
	}
	return store.Row{"success": true, "message": "知识树保存成功", "course_id": course["id"]}, nil
}
