package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Assignments struct{ Service *service.Assignments }

func (h *Assignments) Mine(c *gin.Context) {
	result, err := h.Service.Mine(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Assignments) Submit(c *gin.Context) {
	value, ok := bindInput[dto.SubmissionInput](c)
	if !ok {
		return
	}
	result, err := h.Service.Submit(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusCreated, result)
}
func (h *Assignments) Submissions(c *gin.Context) {
	result, err := h.Service.Submissions(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Assignments) Review(c *gin.Context) {
	value, ok := bindInput[dto.SubmissionInput](c)
	if !ok {
		return
	}
	result, err := h.Service.Review(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Assignments) DeleteSubmission(c *gin.Context) {
	result, err := h.Service.DeleteSubmission(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Assignments) Export(c *gin.Context) {
	result, err := h.Service.Export(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
