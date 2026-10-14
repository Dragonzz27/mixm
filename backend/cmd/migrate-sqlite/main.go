package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
	_ "modernc.org/sqlite"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func convert(typ string, v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	if b, ok := v.([]byte); ok {
		v = string(b)
	}
	switch typ {
	case "boolean":
		switch x := v.(type) {
		case int64:
			return x != 0, nil
		case string:
			return x == "1" || x == "true", nil
		}
	case "jsonb":
		if s, ok := v.(string); ok {
			if s == "" {
				return nil, nil
			}
			if !json.Valid([]byte(s)) {
				return nil, fmt.Errorf("invalid JSON")
			}
			return s, nil
		}
	case "timestamptz":
		if s, ok := v.(string); ok {
			for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999-07:00", "2006-01-02 15:04:05.999999", "2006-01-02T15:04:05.999999"} {
				if t, e := time.Parse(layout, s); e == nil {
					return t, nil
				}
			}
			return nil, fmt.Errorf("invalid timestamp %q", s)
		}
	}
	return v, nil
}
func run() error {
	path := flag.String("sqlite", "", "源 SQLite 文件（只读）")
	dry := flag.Bool("dry-run", false, "验证后回滚 PostgreSQL 数据导入")
	flag.Parse()
	if *path == "" {
		return fmt.Errorf("usage: migrate-sqlite -sqlite /path/to/app.db [-dry-run]")
	}
	abs, e := filepath.Abs(*path)
	if e != nil {
		return e
	}
	if _, e = os.Stat(abs); e != nil {
		return e
	}
	source, e := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: abs, RawQuery: "mode=ro"}).String())
	if e != nil {
		return e
	}
	defer source.Close()
	source.SetMaxOpenConns(1)
	snapshot, e := source.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if e != nil {
		return e
	}
	defer snapshot.Rollback()
	c, e := config.Load()
	if e != nil {
		return e
	}
	ctx := context.Background()
	db, e := store.Open(ctx, c.DatabaseURL)
	if e != nil {
		return e
	}
	defer db.Close()
	if e = store.Migrate(ctx, db); e != nil {
		return e
	}
	tx, e := db.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SET CONSTRAINTS ALL DEFERRED"); e != nil {
		return e
	}
	tables := []string{}
	for t := range store.Columns {
		if t != "instance_tasks" {
			tables = append(tables, t)
		}
	}
	sort.Strings(tables)
	for _, table := range tables {
		var n int64
		if e = tx.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); e != nil {
			return e
		}
		if n != 0 {
			return fmt.Errorf("目标表 %s 非空，导入已取消", table)
		}
	}
	total := 0
	for _, table := range tables {
		var exists int
		if e = snapshot.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&exists); e != nil {
			return e
		}
		if exists == 0 {
			continue
		}
		rows, e := snapshot.Query("SELECT * FROM \"" + table + "\"")
		if e != nil {
			return e
		}
		cols, e := rows.Columns()
		if e != nil {
			rows.Close()
			return e
		}
		names := []string{}
		indices := []int{}
		for i, col := range cols {
			if _, ok := store.Columns[table][col]; ok {
				names = append(names, col)
				indices = append(indices, i)
			}
		}
		values := [][]any{}
		for rows.Next() {
			raw := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range raw {
				ptrs[i] = &raw[i]
			}
			if e = rows.Scan(ptrs...); e != nil {
				rows.Close()
				return e
			}
			item := []any{}
			for _, i := range indices {
				value, err := convert(store.Columns[table][cols[i]], raw[i])
				if err != nil {
					rows.Close()
					return fmt.Errorf("%s.%s: %w", table, cols[i], err)
				}
				if strings.HasSuffix(cols[i], "_id") && value == "" {
					value = nil
				}
				item = append(item, value)
			}
			values = append(values, item)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		if len(values) > 0 {
			n, err := tx.CopyFrom(ctx, pgx.Identifier{table}, names, pgx.CopyFromRows(values))
			if err != nil {
				return fmt.Errorf("导入 %s: %w", table, err)
			}
			total += int(n)
			fmt.Printf("%s: %d\n", table, n)
		}
	}
	if _, e = tx.Exec(ctx, "SET CONSTRAINTS ALL IMMEDIATE"); e != nil {
		return fmt.Errorf("关联数据校验失败（已回滚）: %w", e)
	}
	if *dry {
		fmt.Printf("验证通过，共 %d 行；dry-run 已回滚\n", total)
		return nil
	}
	if e = tx.Commit(ctx); e == nil {
		fmt.Printf("导入完成，共 %d 行，源 SQLite 未修改\n", total)
	}
	return e
}
