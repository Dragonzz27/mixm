package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Learning struct{ Service *service.Learning }

// PracticeGeneration 绑定练习生成参数并返回选题结果。
func (h *Learning) PracticeGeneration(c *gin.Context) {
	value, ok := bindInput[dto.LearningInput](c)
	if !ok {
		return
	}

	result, err := h.Service.PracticeGeneration(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}

// Generate 绑定测评参数并返回生成或恢复的测试。
func (h *Learning) Generate(c *gin.Context) {
	value, ok := bindInput[dto.LearningInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Generate(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}

	Success(c, status, result.Data)
}

// Submit 绑定答卷参数并返回测评评分结果。
func (h *Learning) Submit(c *gin.Context) {
	value, ok := bindInput[dto.LearningInput](c)
	if !ok {
		return
	}

	value.ID = c.Param("id")
	result, err := h.Service.Submit(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Current 读取当前用户正在进行的模拟测试。
func (h *Learning) Current(c *gin.Context) {
	result, err := h.Service.Current(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Review 读取并返回测评结果及题目解析。
func (h *Learning) Review(c *gin.Context) {
	result, err := h.Service.Review(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// QuestionGeneration 调用大模型生成题目，支持流式返回。
func (h *Learning) QuestionGeneration(c *gin.Context) {
	value, ok := bindInput[dto.LearningInput](c)
	if !ok {
		return
	}

	result, err := h.Service.QuestionGeneration(c.Request.Context(), actor(c), value, emit(c))
	if err != nil {
		Failure(c, err)
		return
	}

	if !c.Writer.Written() {
		Success(c, http.StatusOK, result)
	}
}

// KnowledgeAnalysis 调用大模型分析内容并生成知识树。
func (h *Learning) KnowledgeAnalysis(c *gin.Context) {
	value, ok := bindInput[dto.LearningInput](c)
	if !ok {
		return
	}

	result, err := h.Service.KnowledgeAnalysis(c.Request.Context(), actor(c), value, emit(c))
	if err != nil {
		Failure(c, err)
		return
	}

	if !c.Writer.Written() {
		Success(c, http.StatusOK, result)
	}
}

// SaveKnowledge 校验并保存生成的知识树。
func (h *Learning) SaveKnowledge(c *gin.Context) {
	value, ok := bindInput[dto.LearningInput](c)
	if !ok {
		return
	}

	result, err := h.Service.SaveKnowledge(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}
