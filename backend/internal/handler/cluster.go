package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/service"
)

type Cluster struct{ Service *service.Cluster }

func (h *Cluster) GetPods(c *gin.Context) {
	result, err := h.Service.GetPods(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, false)
}
func (h *Cluster) GetNodes(c *gin.Context) {
	result, err := h.Service.GetNodes(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, false)
}
func (h *Cluster) GetDeployments(c *gin.Context) {
	result, err := h.Service.GetDeployments(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, false)
}
func (h *Cluster) DeleteDeployments(c *gin.Context) {
	result, err := h.Service.DeleteDeployments(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Cluster) DeletePods(c *gin.Context) {
	result, err := h.Service.DeletePods(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Cluster) GetPodsLogs(c *gin.Context) {
	result, err := h.Service.GetPodsLogs(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Cluster) GetPodsDescribe(c *gin.Context) {
	result, err := h.Service.GetPodsDescribe(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Cluster) GetPv(c *gin.Context) {
	result, err := h.Service.GetPv(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, false)
}
func (h *Cluster) DeletePv(c *gin.Context) {
	result, err := h.Service.DeletePv(c.Request.Context(), actor(c), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *Cluster) GetPvc(c *gin.Context) {
	result, err := h.Service.GetPvc(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}
	FinishPage(c, http.StatusOK, result, false)
}
func (h *Cluster) DeletePvc(c *gin.Context) {
	result, err := h.Service.DeletePvc(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
