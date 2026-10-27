package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm-edu/backend/internal/dto"
	"mixm-edu/backend/internal/service"
)

type Learning struct{ Service *service.Learning }

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
func (h *Learning) Current(c *gin.Context) {
	result, err := h.Service.Current(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Learning) Review(c *gin.Context) {
	result, err := h.Service.Review(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
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
