// Package apperror 定义业务服务、外部适配器和 HTTP 处理器共用的应用错误。
package apperror

// Error 保存接口错误映射使用的状态码和公开消息。
type Error struct {
	Status int
	Detail string
}

// Error 返回可公开展示的应用错误消息。
func (err *Error) Error() string {
	return err.Detail
}
