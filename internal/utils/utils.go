package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

func GetAbsoluteMigrationsDir(dir string) (string, error) {
	migrationsDir := dir
	if !filepath.IsAbs(migrationsDir) {
		var err error
		migrationsDir, err = filepath.Abs(migrationsDir)
		if err != nil {
			return "", fmt.Errorf("get dir absolute path: %w", err)
		}
	}

	return migrationsDir, nil
}

func GetMigrationFileName(migrationName string) string {
	return time.Now().Format("2006_01_02T15_04_05__" + migrationName)
}

func CreateMigrationFile(dir, migrationName, extension string) (*os.File, error) {
	migrationsDir, err := GetAbsoluteMigrationsDir(dir)
	if err != nil {
		return nil, fmt.Errorf("get migration dir absolute path: %w", err)
	}

	f, err := os.Create(path.Join(migrationsDir, GetMigrationFileName(migrationName)+"."+extension))
	if err != nil {
		return nil, fmt.Errorf("create migration file: %w", err)
	}

	return f, nil
}
