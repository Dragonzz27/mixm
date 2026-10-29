package repository

import (
	"context"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate 按编号顺序执行尚未应用的数据库迁移。
func Migrate(ctx context.Context, p *DB) error {
	pool, err := p.SQL()
	if err != nil {
		return err
	}

	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(47219025)"); err != nil {
		return err
	}

	if _, err = tx.ExecContext(
		ctx,
		"CREATE TABLE IF NOT EXISTS schema_migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())",
	); err != nil {
		return err
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, f := range entries {
		var exists bool
		if err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name=$1)", f.Name()).Scan(&exists); err != nil {
			return err
		}

		if exists {
			continue
		}

		b, err := migrations.ReadFile("migrations/" + f.Name())
		if err != nil {
			return err
		}

		if _, err = tx.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("migration %s: %w", f.Name(), err)
		}

		if _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(name) VALUES($1)", f.Name()); err != nil {
			return err
		}
	}

	return tx.Commit()
}
