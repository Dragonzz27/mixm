package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Users struct{ Service *service.Users }

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
func (h *Users) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Users) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
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
func (h *Users) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
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
func (h *Users) Permissions(c *gin.Context) {
	result, err := h.Service.Permissions()
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
