package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/jmoiron/sqlx"
	migrator "github.com/wursta/gomigrator/internal/migrator"
)

func ParseMigrations(migrationsDir string) ([]migrator.Migration, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	migrations := make([]migrator.Migration, len(files))
	parsingErrors := map[string]error{}
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for i, file := range files {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			filePath := filepath.Join(migrationsDir, file.Name())
			upStmt, downStmt, err := parseMigrationFile(filePath)
			if err != nil {
				parsingErrors[file.Name()] = err
				return
			}

			mu.Lock()
			defer mu.Unlock()
			migrations[i] = migrator.Migration{
				Name:     filepath.Base(file.Name()),
				FilePath: filepath.Join(migrationsDir, file.Name()),
				UpHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
					_, err := tx.ExecContext(ctx, upStmt)
					if err != nil {
						return err
					}

					return nil
				},
				DownHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
					_, err := tx.ExecContext(ctx, downStmt)
					if err != nil {
						return err
					}

					return nil
				},
			}
		}(i)
	}
	wg.Wait()

	if len(parsingErrors) != 0 {
		return nil, fmt.Errorf("error while parsing migration files: %s", parsingErrors)
	}

	return migrations, nil
}

func parseMigrationFile(filePath string) (upStmt, downStmt string, err error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	re := regexp.MustCompile(`(?ms)-- migration: up\n(?P<up_stmt>.*)\n-- migration: down\n(?P<down_stmt>.*)`)
	match := re.FindSubmatch(b)

	for i, name := range re.SubexpNames() {
		if name == "up_stmt" {
			upStmt = string(match[i])
		} else if name == "down_stmt" {
			downStmt = string(match[i])
		}
	}

	return
}
