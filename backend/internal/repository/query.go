package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"mixm-edu/backend/internal/model"
)

var placeholder = regexp.MustCompile(`\$(\d+)`)

// Named parameters retain repeated PostgreSQL placeholders and JSONB '?' operators.
func bind(query string, args []any) (string, map[string]any) {
	named := map[string]any{}
	for i, arg := range args {
		if values, ok := arg.([]string); ok {
			arg = TextArray(values)
		}
		named[fmt.Sprintf("arg%d", i+1)] = arg
	}
	query = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_.]*)\s+\?\s+(\$\d+)`).ReplaceAllString(query, "jsonb_exists($1, $2)")
	return placeholder.ReplaceAllString(query, "@arg$1"), named
}
func (db *DB) Query(ctx context.Context, query string, args ...any) ([]model.Record, error) {
	query, named := bind(query, args)
	statement := db.orm.WithContext(ctx)
	if len(named) == 0 {
		statement = statement.Raw(query)
	} else {
		statement = statement.Raw(query, named)
	}
	rows, err := statement.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	result := []model.Record{}
	for rows.Next() {
		values := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		record := model.Record{}
		for i, key := range cols {
			v := values[i]
			if b, ok := v.([]byte); ok {
				switch types[i].DatabaseTypeName() {
				case "JSON", "JSONB":
					if err := json.Unmarshal(b, &v); err != nil {
						return nil, err
					}
				default:
					v = string(b)
				}
			} else if text, ok := v.(string); ok && (types[i].DatabaseTypeName() == "JSONB" || types[i].DatabaseTypeName() == "JSON") {
				if err := json.Unmarshal([]byte(text), &v); err != nil {
					return nil, err
				}
			}
			record[key] = v
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

type Result struct{ affected int64 }

func (r Result) RowsAffected() int64 { return r.affected }
func (db *DB) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	query, named := bind(query, args)
	statement := db.orm.WithContext(ctx)
	var result *gorm.DB
	if len(named) == 0 {
		result = statement.Exec(query)
	} else {
		result = statement.Exec(query, named)
	}
	return Result{affected: result.RowsAffected}, result.Error
}

type TextArray []string

func (a TextArray) Value() (driver.Value, error) {
	parts := make([]string, len(a))
	for i, value := range a {
		parts[i] = "\"" + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(value) + "\""
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
