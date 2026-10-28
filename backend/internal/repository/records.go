package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mixm/backend/internal/model"
)

var ErrInvalidValue = errors.New("invalid value")
var ErrNotFound = gorm.ErrRecordNotFound

func Query(ctx context.Context, db *DB, query string, args ...any) ([]model.Record, error) {
	return db.Query(ctx, query, args...)
}
func One(ctx context.Context, db *DB, query string, args ...any) (model.Record, error) {
	rows, err := Query(ctx, db, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	return rows[0], nil
}
func Get(ctx context.Context, db *DB, table, id string) (model.Record, error) {
	if Columns[table] == nil {
		return nil, fmt.Errorf("unknown table: %s", table)
	}
	entity := model.Entity(table)
	if entity == nil {
		return nil, fmt.Errorf("unknown entity: %s", table)
	}
	if err := db.orm.WithContext(ctx).Table(table).Where("id = ?", id).First(entity).Error; err != nil {
		return nil, err
	}
	return model.EntityRecord(entity), nil
}
func Write(ctx context.Context, db *DB, table, id string, data model.Record) (model.Record, error) {
	cols, ok := Columns[table]
	if !ok {
		return nil, fmt.Errorf("unknown table: %s", table)
	}
	fields := map[string]any{}
	for key, v := range data {
		if _, ok := cols[key]; !ok || key == "id" || key == "created_at" || key == "updated_at" {
			continue
		}
		converted, err := value(cols[key], v)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrInvalidValue, key, err)
		}
		fields[key] = converted
	}
	query := db.orm.WithContext(ctx).Table(table)
	if id == "" {
		id = uuid.NewString()
		fields["id"] = id
		if err := query.Clauses(clause.Returning{}).Create(fields).Error; err != nil {
			return nil, err
		}
	} else if len(fields) > 0 {
		if _, ok := cols["updated_at"]; ok {
			fields["updated_at"] = gorm.Expr("now()")
		}
		result := query.Where("id = ?", id).Updates(fields)
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			return nil, ErrNotFound
		}
	}
	return Get(ctx, db, table, id)
}
func Delete(ctx context.Context, db *DB, table, id string) error {
	if Columns[table] == nil {
		return fmt.Errorf("unknown table: %s", table)
	}
	result := db.orm.WithContext(ctx).Table(table).Where("id = ?", id).Delete(map[string]any{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func value(typ string, v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch typ {
	case "double precision":
		if n, ok := v.(json.Number); ok {
			f, err := n.Float64()
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				return nil, fmt.Errorf("must be finite number")
			}
			return f, nil
		}
		switch n := v.(type) {
		case float64:
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return nil, fmt.Errorf("must be finite number")
			}
			return n, nil
		case int:
			return float64(n), nil
		case int32:
			return float64(n), nil
		case int64:
			return float64(n), nil
		}
		return nil, fmt.Errorf("must be number")
	case "text":
		if _, ok := v.(string); !ok {
			return nil, fmt.Errorf("must be string")
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return nil, fmt.Errorf("must be boolean")
		}
	case "jsonb":
		b, err := json.Marshal(v)
		return string(b), err
	case "integer":
		switch n := v.(type) {
		case float64:
			if n != float64(int64(n)) {
				return nil, fmt.Errorf("must be integer")
			}
			return int64(n), nil
		case json.Number:
			return n.Int64()
		case int, int32, int64:
			return n, nil
		}
		return nil, fmt.Errorf("must be integer")
	case "timestamptz":
		if s, ok := v.(string); ok {
			for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02 15:04:05.999999"} {
				if t, err := time.Parse(layout, s); err == nil {
					return t, nil
				}
			}
			return nil, fmt.Errorf("invalid timestamp")
		}
		if _, ok := v.(time.Time); !ok {
			return nil, fmt.Errorf("must be timestamp")
		}
	}
	return v, nil
}
