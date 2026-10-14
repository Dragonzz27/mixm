package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm-edu/backend/internal/dto"
	"mixm-edu/backend/internal/service"
)

type Instances struct {
	Service *service.Instances
	all     bool
}

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
func (h *Instances) List(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), "", h.all)
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Instances) Detail(c *gin.Context) {
	result, err := h.Service.Detail(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
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
func (h *Instances) Delete(c *gin.Context) {
	result, err := h.Service.Delete(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Instances) DeleteAll(c *gin.Context) {
	result, err := h.Service.DeleteAll(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
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
func (h *Instances) Task(c *gin.Context) {
	result, err := h.Service.Task(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Instances) Tasks(c *gin.Context) {
	result, err := h.Service.Tasks(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Instances) Summary(c *gin.Context) {
	result, err := h.Service.Summary(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}

func (h *Instances) Mine(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), "", false)
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
func (h *Instances) ForUser(c *gin.Context) {
	result, err := h.Service.List(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("id"), false)
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, true)
}
