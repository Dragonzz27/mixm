package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm-edu/backend/internal/dto"
	"mixm-edu/backend/internal/service"
)

type Resources struct {
	Service *service.Resources
	spec    service.ResourceSpec
}

func (h *Resources) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Resources) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"), c.Request.URL.Query(), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Resources) Save(c *gin.Context) {
	value, ok := bindInput[dto.ResourceInput](c)
	if !ok {
		return
	}
	result, err := h.Service.Save(c.Request.Context(), actor(c), value, c.Param("id"), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Resources) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Resources) CategoryTree(c *gin.Context) {
	result, err := h.Service.CategoryTree(c.Request.Context())
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Resources) QuestionCategoryTree(c *gin.Context) {
	result, err := h.Service.QuestionCategoryTree(c.Request.Context())
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Resources) BatchQuestions(c *gin.Context) {
	var values []dto.ResourceInput
	if !decode(c, &values) {
		return
	}
	result, err := h.Service.BatchQuestions(c.Request.Context(), actor(c), values)
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusCreated, result)
}
func (h *Resources) ReviewQuestion(c *gin.Context) {
	value, ok := bindInput[dto.ResourceInput](c)
	if !ok {
		return
	}
	result, err := h.Service.ReviewQuestion(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}

func (h *Resources) Create(c *gin.Context) {
	value, ok := bindInput[dto.ResourceInput](c)
	if !ok {
		return
	}
	result, err := h.Service.Save(c.Request.Context(), actor(c), value, "", h.spec)
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusCreated, result)
}
