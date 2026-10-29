package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Users struct{ Service *service.Users }

// Create 创建用户并返回结果。
func (h *Users) Create(c *gin.Context) {
	value, ok := bindInput[dto.UserInput](c)
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

// Login 校验登录凭据并返回用户身份和令牌。
func (h *Users) Login(c *gin.Context) {
	value, ok := bindInput[dto.UserInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Login(c.Request.Context(), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Refresh 校验刷新令牌并签发新令牌。
func (h *Users) Refresh(c *gin.Context) {
	value, ok := bindInput[dto.UserInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Refresh(c.Request.Context(), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// List 按权限与查询条件返回用户列表。
func (h *Users) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Detail 读取指定用户的详情。
func (h *Users) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Update 校验并更新指定用户。
func (h *Users) Update(c *gin.Context) {
	value, ok := bindInput[dto.UserInput](c)
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

// Password 校验并更新用户登录密码。
func (h *Users) Password(c *gin.Context) {
	value, ok := bindInput[dto.UserInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Password(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Delete 删除指定用户及其必要关联。
func (h *Users) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Import 读取 Excel 并批量导入用户。
func (h *Users) Import(c *gin.Context) {
	value, cleanup, err := upload(c, 20<<20)
	if err != nil {
		Failure(c, err)
		return
	}

	defer cleanup()

	result, err := h.Service.Import(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}

// Permissions 返回前端可配置的角色权限选项。
func (h *Users) Permissions(c *gin.Context) {
	result, err := h.Service.Permissions()
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
