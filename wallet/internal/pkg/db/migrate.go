package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

func createMigrationsTable(ctx context.Context, conn *pgxpool.Pool) error {
	createQuery := `CREATE TABLE IF NOT EXISTS migrations (
		id SERIAL PRIMARY KEY,
		name varchar(255) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err := conn.Exec(ctx, createQuery)

	return err
}

func getAppliedMigrations(ctx context.Context, conn *pgxpool.Pool) (map[string]struct{}, error) {
	selectQuery := `SELECT name FROM migrations;`

	rows, err := conn.Query(ctx, selectQuery)
	if err != nil {
		return nil, err
	}

	names := make(map[string]struct{})

	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		names[name] = struct{}{}
	}

	return names, nil
}

func applyMigration(ctx context.Context, migrationName, query string, conn *pgxpool.Pool) error {
	updateAppliedMigrationsQuery := fmt.Sprintf(`INSERT INTO migrations(name) VALUES ('%s');`, migrationName)

	finalQuery := query + updateAppliedMigrationsQuery

	if _, err := conn.Exec(ctx, finalQuery); err != nil {
		return err
	}

	return nil
}

func Migrate(ctx context.Context, conn *pgxpool.Pool) error {
	migrationsRootPath := "./internal/pkg/db/migrations"

	err := createMigrationsTable(ctx, conn)
	if err != nil {
		return err
	}

	appliedMigrations, err := getAppliedMigrations(ctx, conn)
	if err != nil {
		return err
	}

	log.Printf("Applied migrations: %v\n", appliedMigrations)

	files, err := os.ReadDir(migrationsRootPath)
	if err != nil {
		return err
	}

	// TODO sort migrations by name
	for _, f := range files {
		filename := f.Name()
		extension := filepath.Ext(filename)

		if extension == ".sql" {
			migratinoName := strings.TrimSuffix(filename, extension)

			if _, ok := appliedMigrations[migratinoName]; ok {
				continue
			}

			path := filepath.Join(migrationsRootPath, filename)

			file, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			query := string(file)

			err = applyMigration(ctx, migratinoName, query, conn)
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}
