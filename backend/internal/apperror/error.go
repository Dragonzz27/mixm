// Package apperror defines application errors shared by services, adapters and HTTP handlers.
package apperror

// Error carries the HTTP status and public message used by the API error mapping.
type Error struct {
	Status int
	Detail string
}

func (err *Error) Error() string { return err.Detail }
