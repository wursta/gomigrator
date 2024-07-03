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

const (
	MigrationFileDateTime = "2006_01_02T15_04_05"
	letters               = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
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

func GetMigrationFileName(createDatetime time.Time, migrationName, suffix string) string {
	return createDatetime.Format(MigrationFileDateTime + "__" + migrationName + "__" + suffix)
}

func CreateMigrationFile(
	createDatetime time.Time,
	dir,
	migrationName,
	migrationSuffix,
	fileExtension string,
) (*os.File, error) {
	migrationsDir, err := GetAbsoluteMigrationsDir(dir)
	if err != nil {
		return nil, fmt.Errorf("get migration dir absolute path: %w", err)
	}

	f, err := os.Create(path.Join(
		migrationsDir,
		GetMigrationFileName(createDatetime, migrationName, migrationSuffix)+"."+fileExtension),
	)
	if err != nil {
		return nil, fmt.Errorf("create migration file: %w", err)
	}

	return f, nil
}

func GenerateRandomString(length uint8) string {
	ret := make([]byte, length)
	var i uint8
	for i = 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}
