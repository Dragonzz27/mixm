package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/service"
)

type Files struct{ Service *service.Files }

// AssignmentUpload 保存作业附件并返回文件信息。
func (h *Files) AssignmentUpload(c *gin.Context) {
	value, cleanup, err := upload(c, 50<<20)
	if err != nil {
		Failure(c, err)
		return
	}

	defer cleanup()

	result, err := h.Service.AssignmentUpload(value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}

// AssignmentDownload 校验附件路径并返回作业文件。
func (h *Files) AssignmentDownload(c *gin.Context) {
	result, err := h.Service.AssignmentDownload(c.Param("filename"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// ContentUpload 上传教学内容附件到对象存储。
func (h *Files) ContentUpload(c *gin.Context) {
	value, cleanup, err := upload(c, 100<<20)
	if err != nil {
		Failure(c, err)
		return
	}

	defer cleanup()

	result, err := h.Service.ContentUpload(c.Request.Context(), actor(c), value)
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}

// ContentDownload 读取对象存储中的内容附件。
func (h *Files) ContentDownload(c *gin.Context) {
	result, err := h.Service.ContentDownload(c.Request.Context(), strings.TrimPrefix(c.Param("key"), "/"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
