package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/service"
)

type System struct{ Service *service.System }

func (h *System) Stats(c *gin.Context) {
	result, err := h.Service.Stats(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
func (h *System) Load(c *gin.Context) {
	result, err := h.Service.Load(c.Request.Context(), actor(c))
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, http.StatusOK, result)
}
