package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Instances struct {
	Service *service.Instances
	all     bool
}

// Create 创建实例并返回结果。
func (h *Instances) Create(c *gin.Context) {
	value, ok := bindInput[dto.InstanceInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Create(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusAccepted, result)
}

// List 按权限与查询条件返回实例列表。
func (h *Instances) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), "", h.all)
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Detail 读取指定实例的详情。
func (h *Instances) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Update 校验并更新指定实例。
func (h *Instances) Update(c *gin.Context) {
	value, ok := bindInput[dto.InstanceInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Update(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Delete 删除指定实例及其必要关联。
func (h *Instances) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// DeleteAll 删除当前管理范围内的全部实例。
func (h *Instances) DeleteAll(c *gin.Context) {
	result, err := h.Service.DeleteAll(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Batch 绑定模板和用户列表，返回批量实例创建任务。
func (h *Instances) Batch(c *gin.Context) {
	value, ok := bindInput[dto.InstanceInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Batch(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusAccepted, result)
}

// Task 读取单个实例创建任务的进度。
func (h *Instances) Task(c *gin.Context) {
	result, err := h.Service.Task(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Tasks 读取用户可见的实例创建任务列表。
func (h *Instances) Tasks(c *gin.Context) {
	result, err := h.Service.Tasks(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Summary 汇总实例创建任务的状态数量。
func (h *Instances) Summary(c *gin.Context) {
	result, err := h.Service.Summary(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Mine 读取当前用户的实例列表。
func (h *Instances) Mine(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), "", false)
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// ForUser 读取指定用户的实例列表。
func (h *Instances) ForUser(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("id"), false)
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}
