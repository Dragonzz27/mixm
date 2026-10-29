package model

// Actor 表示传递给业务服务和仓储的已认证用户身份。
type Actor struct {
	ID          string
	Username    string
	IsSuperuser bool
}
