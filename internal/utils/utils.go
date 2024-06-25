package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	return time.Now().Format("2006_01_02T15_04_05__" + migrationName + "__" + generateRandomString(5))
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

func generateRandomString(length int) string {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}
