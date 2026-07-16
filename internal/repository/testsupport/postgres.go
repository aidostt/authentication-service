// Package testsupport provides shared helpers for repository integration tests.
package testsupport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// SetupPostgres starts a throwaway PostgreSQL, applies the migrations and
// returns a connected pool plus a cleanup func. A non-nil error means the
// environment is unavailable (for example Docker is not running); callers should
// skip rather than fail in that case. Migrations are resolved relative to the
// caller's working directory at ../../migrations, which holds for the
// internal/repository package.
func SetupPostgres(ctx context.Context) (*pgxpool.Pool, func(), error) {
	container, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("reservista"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start postgres: %w", err)
	}
	terminate := func() { _ = container.Terminate(ctx) }

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		terminate()
		return nil, nil, err
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		terminate()
		return nil, nil, err
	}
	cleanup := func() {
		pool.Close()
		terminate()
	}

	if err := applyMigrations(ctx, pool); err != nil {
		cleanup()
		return nil, nil, err
	}
	return pool, cleanup, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no migrations found")
	}
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
	}
	return nil
}
