package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mixm-edu/backend/internal/model"
)

// DB scopes GORM queries to a pool, transaction, or pinned worker connection.
type DB struct{ orm *gorm.DB }

func New(orm *gorm.DB) *DB           { return &DB{orm: orm} }
func (db *DB) GORM() *gorm.DB        { return db.orm }
func (db *DB) SQL() (*sql.DB, error) { return db.orm.DB() }
func (db *DB) Close() error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}
	return pool.Close()
}
func (db *DB) Ping(ctx context.Context) error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}
	return pool.PingContext(ctx)
}
func (db *DB) Begin(ctx context.Context) (*DB, error) {
	tx := db.orm.WithContext(ctx).Begin()
	return New(tx), tx.Error
}
func (db *DB) Commit(ctx context.Context) error   { return db.orm.WithContext(ctx).Commit().Error }
func (db *DB) Rollback(ctx context.Context) error { return db.orm.WithContext(ctx).Rollback().Error }
func (db *DB) Transaction(ctx context.Context, fn func(*DB) error) error {
	return db.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(New(tx)) })
}

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

// WithTaskLock holds a session advisory lock on one dedicated physical connection.
func (db *DB) WithTaskLock(ctx context.Context, id string, fn func(*DB) error) error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}
	conn, err := pool.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var locked bool
	if err = conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtextextended($1,0))", id).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var unlocked bool
		if err := conn.QueryRowContext(unlockCtx, "SELECT pg_advisory_unlock(hashtextextended($1,0))", id).Scan(&unlocked); err != nil || !unlocked {
			// ErrBadConn removes the locked connection instead of returning it to the pool.
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
	}()
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	}
	return fn(New(orm))
}

type TextArray []string

func (a TextArray) Value() (driver.Value, error) {
	parts := make([]string, len(a))
	for i, value := range a {
		parts[i] = "\"" + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(value) + "\""
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
