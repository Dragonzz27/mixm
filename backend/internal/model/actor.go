package model

// Actor is the authenticated identity supplied to services and repositories.
type Actor struct {
	ID          string
	Username    string
	IsSuperuser bool
}
