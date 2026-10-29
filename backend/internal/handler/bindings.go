package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Bindings struct {
	Service *service.Bindings
	unbind  bool
}

// Create 创建内容绑定并返回结果。
func (h *Bindings) Create(c *gin.Context) {
	value, ok := bindInput[dto.BindingInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Create(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}

// List 按权限与查询条件返回内容绑定列表。
func (h *Bindings) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Detail 读取指定内容绑定的详情。
func (h *Bindings) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Delete 删除指定内容绑定及其必要关联。
func (h *Bindings) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// DeleteForContent 清除指定内容的分类绑定。
func (h *Bindings) DeleteForContent(c *gin.Context) {
	result, err := h.Service.DeleteForContent(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Batch 绑定批量关联参数并处理分类内容绑定请求。
func (h *Bindings) Batch(c *gin.Context) {
	value, ok := bindInput[dto.BindingInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Batch(c.Request.Context(), actor(c), value, c.Param("id"), h.unbind)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Counts 统计分类中的内容绑定数量。
func (h *Bindings) Counts(c *gin.Context) {
	result, err := h.Service.Counts(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// CategoryContents 读取指定分类关联的内容列表。
func (h *Bindings) CategoryContents(c *gin.Context) {
	result, err := h.Service.CategoryContents(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// ContentTree 返回绑定内容与分类的树形结果。
func (h *Bindings) ContentTree(c *gin.Context) {
	result, err := h.Service.ContentTree(c.Request.Context(), c.Request.URL.Query(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
