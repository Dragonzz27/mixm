package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/service"
)

type Cluster struct{ Service *service.Cluster }

// GetPods 读取集群 Pod 列表及运行状态。
func (h *Cluster) GetPods(c *gin.Context) {
	result, err := h.Service.GetPods(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, false)
}

// GetNodes 汇总集群节点状态与资源使用信息。
func (h *Cluster) GetNodes(c *gin.Context) {
	result, err := h.Service.GetNodes(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, false)
}

// GetDeployments 读取 Deployment 副本数和运行状态。
func (h *Cluster) GetDeployments(c *gin.Context) {
	result, err := h.Service.GetDeployments(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, false)
}

// DeleteDeployments 删除指定 Deployment 并保护系统资源。
func (h *Cluster) DeleteDeployments(c *gin.Context) {
	result, err := h.Service.DeleteDeployments(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// DeletePods 删除指定 Pod 并保护系统资源。
func (h *Cluster) DeletePods(c *gin.Context) {
	result, err := h.Service.DeletePods(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// GetPodsLogs 读取指定 Pod 容器的最近日志。
func (h *Cluster) GetPodsLogs(c *gin.Context) {
	result, err := h.Service.GetPodsLogs(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// GetPodsDescribe 汇总 Pod 状态、容器配置及事件。
func (h *Cluster) GetPodsDescribe(c *gin.Context) {
	result, err := h.Service.GetPodsDescribe(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// GetPv 读取持久卷列表及绑定状态。
func (h *Cluster) GetPv(c *gin.Context) {
	result, err := h.Service.GetPv(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, false)
}

// DeletePv 检查绑定和回收策略后删除持久卷。
func (h *Cluster) DeletePv(c *gin.Context) {
	result, err := h.Service.DeletePv(c.Request.Context(), actor(c), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// GetPvc 读取指定范围内的持久卷声明。
func (h *Cluster) GetPvc(c *gin.Context) {
	result, err := h.Service.GetPvc(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, false)
}

// DeletePvc 删除指定持久卷声明并保护系统资源。
func (h *Cluster) DeletePvc(c *gin.Context) {
	result, err := h.Service.DeletePvc(c.Request.Context(), actor(c), c.Param("namespace"), c.Param("name"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
