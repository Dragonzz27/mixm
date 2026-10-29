package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/service"
)

type Assignments struct{ Service *service.Assignments }

// Mine 读取当前用户的作业列表。
func (h *Assignments) Mine(c *gin.Context) {
	result, err := h.Service.Mine(c.Request.Context(), actor(c), c.Request.URL.Query())
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Submit 校验并提交作业或测评答案。
func (h *Assignments) Submit(c *gin.Context) {
	value, ok := bindInput[dto.SubmissionInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Submit(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusCreated, result)
}

// Submissions 按权限查询作业提交记录。
func (h *Assignments) Submissions(c *gin.Context) {
	result, err := h.Service.Submissions(c.Request.Context(), actor(c), c.Request.URL.Query(), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	FinishPage(c, http.StatusOK, result, true)
}

// Review 绑定批改参数并返回作业评分结果。
func (h *Assignments) Review(c *gin.Context) {
	value, ok := bindInput[dto.SubmissionInput](c)
	if !ok {
		return
	}

	result, err := h.Service.Review(c.Request.Context(), actor(c), value, c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// DeleteSubmission 删除指定作业提交记录。
func (h *Assignments) DeleteSubmission(c *gin.Context) {
	result, err := h.Service.DeleteSubmission(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// Export 生成作业提交记录的导出文件。
func (h *Assignments) Export(c *gin.Context) {
	result, err := h.Service.Export(c.Request.Context(), actor(c), c.Param("id"))
	if err != nil {
		Failure(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}
