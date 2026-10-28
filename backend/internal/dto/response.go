package dto

import (
	"fmt"
	"net/url"
	"strconv"

	"mixm/backend/internal/model"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}
type Page[T any] struct {
	Category model.Record `json:"category,omitempty"`
	Items    []T          `json:"items"`
	Total    int          `json:"total"`
	Offset   int          `json:"offset"`
	Limit    int          `json:"limit"`
}
type Pagination struct{ Offset, Limit int }

func ParsePagination(query url.Values) (Pagination, error) {
	p := Pagination{Limit: 100}
	var err error
	if v := query.Get("offset"); v != "" {
		p.Offset, err = strconv.Atoi(v)
		if err != nil || p.Offset < 0 {
			return p, fmt.Errorf("offset 无效")
		}
	}
	if v := query.Get("limit"); v != "" {
		p.Limit, err = strconv.Atoi(v)
		if err != nil || p.Limit < 1 || p.Limit > 1000 {
			return p, fmt.Errorf("limit 必须为 1-1000")
		}
	}
	return p, nil
}

// Creation distinguishes a new resource from an existing resource resumed by POST.
type Creation struct {
	Data    model.Record
	Created bool
}
