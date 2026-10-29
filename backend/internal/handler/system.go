package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/service"
)

type System struct{ Service *service.System }

// Stats 读取宿主机 CPU、内存、磁盘和进程统计。
func (h *System) Stats(c *gin.Context) {
	result, err := h.Service.Stats(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Load 读取宿主机近期负载。
func (h *System) Load(c *gin.Context) {
	result, err := h.Service.Load(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
