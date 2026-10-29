package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Resources struct {
	Service *service.Resources
	spec    service.ResourceSpec
}

// List 按权限与查询条件返回通用资源列表。
func (h *Resources) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Detail 读取指定通用资源的详情。
func (h *Resources) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"), c.Request.URL.Query(), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Save 校验并保存通用资源。
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

// Delete 删除指定通用资源及其必要关联。
func (h *Resources) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"), h.spec)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// CategoryTree 返回内容分类树。
func (h *Resources) CategoryTree(c *gin.Context) {
	result, err := h.Service.CategoryTree(c.Request.Context())
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// QuestionCategoryTree 返回题库分类树。
func (h *Resources) QuestionCategoryTree(c *gin.Context) {
	result, err := h.Service.QuestionCategoryTree(c.Request.Context())
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// BatchQuestions 批量创建并保存题目。
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

// ReviewQuestion 记录题目审核结果。
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

// Create 创建通用资源并返回结果。
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
