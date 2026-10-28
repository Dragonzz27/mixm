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
func (h *Bindings) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Bindings) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Bindings) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Bindings) DeleteForContent(c *gin.Context) {
	result, err := h.Service.DeleteForContent(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
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
func (h *Bindings) Counts(c *gin.Context) {
	result, err := h.Service.Counts(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Bindings) CategoryContents(c *gin.Context) {
	result, err := h.Service.CategoryContents(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Bindings) ContentTree(c *gin.Context) {
	result, err := h.Service.ContentTree(c.Request.Context(), c.Request.URL.Query(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
