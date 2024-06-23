package app

import (
	"errors"
	"os"

	creatorsql "github.com/wursta/gomigrator/internal/creator/sql"
)

type MigrationFormat int

const (
	MigrationFormatSQL MigrationFormat = iota
	MigrationFormatGo
)

type MigratorApp struct {
	migrationsDir string
}

type Creator interface {
	Create(migrationName string) (*os.File, error)
}

func New(migrationsDir string) *MigratorApp {
	return &MigratorApp{
		migrationsDir: migrationsDir,
	}
}

func (a *MigratorApp) CreateMigration(name string, format MigrationFormat) (*os.File, error) {
	var creator Creator
	switch format {
	case MigrationFormatSQL:
		creator = creatorsql.New(a.migrationsDir)
	default:
		return nil, errors.New("creator for this format not realized")
	}

	migrationFile, err := creator.Create(name)
	if err != nil {
		return nil, err
	}

	return migrationFile, nil
}
