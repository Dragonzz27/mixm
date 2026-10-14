package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm-edu/backend/internal/dto"
	"mixm-edu/backend/internal/service"
)

type Images struct {
	Service *service.Images
	mine    bool
}

func (h *Images) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), h.mine)
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
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
func (h *Images) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Images) Tags(c *gin.Context) {
	result, err := h.Service.Tags(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, false)
}
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
func (h *Images) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Images) DeleteTag(c *gin.Context) {
	result, err := h.Service.DeleteTag(c.Request.Context(), actor(c), c.Param("id"), c.Param("tag"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
