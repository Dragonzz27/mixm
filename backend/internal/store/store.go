package store

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Row = map[string]any

var ErrInvalidValue = errors.New("invalid value")

type DB interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	c, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return nil, e
	}
	c.MaxConns = 20
	c.MaxConnLifetime = time.Hour
	p, e := pgxpool.NewWithConfig(ctx, c)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, e
	}
	return p, nil
}
func Migrate(ctx context.Context, p *pgxpool.Pool) error {
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(47219025)"); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); e != nil {
		return e
	}
	entries, e := migrations.ReadDir("migrations")
	if e != nil {
		return e
	}
	for _, f := range entries {
		var exists bool
		if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name=$1)", f.Name()).Scan(&exists); e != nil {
			return e
		}
		if exists {
			continue
		}
		b, e := migrations.ReadFile("migrations/" + f.Name())
		if e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, string(b)); e != nil {
			return fmt.Errorf("migration %s: %w", f.Name(), e)
		}
		if _, e = tx.Exec(ctx, "INSERT INTO schema_migrations(name) VALUES($1)", f.Name()); e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
func Query(ctx context.Context, db DB, sql string, args ...any) ([]Row, error) {
	rows, e := db.Query(ctx, sql, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	result := []Row{}
	for rows.Next() {
		vals, e := rows.Values()
		if e != nil {
			return nil, e
		}
		r := Row{}
		for i, f := range rows.FieldDescriptions() {
			r[f.Name] = vals[i]
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func One(ctx context.Context, db DB, sql string, args ...any) (Row, error) {
	rs, e := Query(ctx, db, sql, args...)
	if e != nil {
		return nil, e
	}
	if len(rs) == 0 {
		return nil, pgx.ErrNoRows
	}
	return rs[0], nil
}
func Get(ctx context.Context, db DB, table, id string) (Row, error) {
	if Columns[table] == nil {
		return nil, fmt.Errorf("unknown table")
	}
	return One(ctx, db, "SELECT * FROM "+table+" WHERE id=$1", id)
}
func Write(ctx context.Context, db DB, table, id string, data Row) (Row, error) {
	cols, ok := Columns[table]
	if !ok {
		return nil, fmt.Errorf("unknown table")
	}
	keys := []string{}
	for k := range data {
		if _, ok := cols[k]; ok && k != "id" && k != "created_at" && k != "updated_at" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	args := []any{}
	values := []string{}
	assign := []string{}
	quoted := []string{}
	for i, k := range keys {
		v, e := value(cols[k], data[k])
		if e != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrInvalidValue, k, e)
		}
		args = append(args, v)
		ph := fmt.Sprintf("$%d", i+1)
		values = append(values, ph)
		quoted = append(quoted, "\""+k+"\"")
		assign = append(assign, "\""+k+"\"="+ph)
	}
	if id == "" {
		id = uuid.NewString()
		args = append(args, id)
		quoted = append(quoted, "id")
		values = append(values, fmt.Sprintf("$%d", len(args)))
		return One(ctx, db, "INSERT INTO "+table+" ("+strings.Join(quoted, ",")+") VALUES ("+strings.Join(values, ",")+") RETURNING *", args...)
	}
	if len(assign) == 0 {
		return Get(ctx, db, table, id)
	}
	if _, ok := cols["updated_at"]; ok {
		assign = append(assign, "updated_at=now()")
	}
	args = append(args, id)
	return One(ctx, db, "UPDATE "+table+" SET "+strings.Join(assign, ",")+fmt.Sprintf(" WHERE id=$%d RETURNING *", len(args)), args...)
}
func value(typ string, v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch typ {
	case "double precision":
		if n, ok := v.(json.Number); ok {
			f, e := n.Float64()
			if e != nil || math.IsNaN(f) || math.IsInf(f, 0) {
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
		b, e := json.Marshal(v)
		return string(b), e
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
				if t, e := time.Parse(layout, s); e == nil {
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
func Delete(ctx context.Context, db DB, table, id string) error {
	if Columns[table] == nil {
		return fmt.Errorf("unknown table")
	}
	tag, e := db.Exec(ctx, "DELETE FROM "+table+" WHERE id=$1", id)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
func String(r Row, k string) string { s, _ := r[k].(string); return s }
func Bool(r Row, k string) bool     { v, _ := r[k].(bool); return v }
func Number(r Row, k string) float64 {
	switch v := r[k].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		n, _ := v.Float64()
		return n
	}
	return 0
}
func Strings(v any) []string {
	out := []string{}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		for _, a := range x {
			if s, ok := a.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}
func JSON(v any) string { b, _ := json.Marshal(v); return string(b) }
func Object(v any) Row {
	if r, ok := v.(map[string]any); ok {
		return r
	}
	if s, ok := v.(string); ok {
		var r Row
		if json.Unmarshal([]byte(s), &r) == nil && r != nil {
			return r
		}
	}
	return Row{}
}
func Objects(v any) []Row {
	out := []Row{}
	if s, ok := v.(string); ok {
		_ = json.Unmarshal([]byte(s), &out)
		return out
	}
	switch a := v.(type) {
	case []Row:
		return a
	case []any:
		for _, r := range a {
			out = append(out, Object(r))
		}
	}
	return out
}
