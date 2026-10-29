package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Images struct {
	Service *service.Images
	mine    bool
}

// List 按权限与查询条件返回镜像仓库列表。
func (h *Images) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), h.mine)
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Create 创建镜像仓库并返回结果。
func (h *Images) Create(c *gin.Context) {
	value, ok := bindInput[dto.ImageInput](c)
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

// Detail 读取指定镜像仓库的详情。
func (h *Images) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Tags 查询镜像仓库的标签及清单信息。
func (h *Images) Tags(c *gin.Context) {
	result, err := h.Service.Tags(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, false)
}

// Update 校验并更新指定镜像仓库。
func (h *Images) Update(c *gin.Context) {
	value, ok := bindInput[dto.ImageInput](c)
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

// Delete 删除指定镜像仓库及其必要关联。
func (h *Images) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// DeleteTag 删除镜像仓库中的指定标签。
func (h *Images) DeleteTag(c *gin.Context) {
	result, err := h.Service.DeleteTag(c.Request.Context(), actor(c), c.Param("id"), c.Param("tag"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
