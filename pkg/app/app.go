package app

import (
	"context"
	"errors"
	"os"

	creatorsql "github.com/wursta/gomigrator/internal/creator/sql"
	migratorConstants "github.com/wursta/gomigrator/internal/migrator"
	pgmigrator "github.com/wursta/gomigrator/internal/migrator/pg"
	"github.com/wursta/gomigrator/internal/parser"
)

type MigrationFormat int

const (
	MigrationFormatSQL MigrationFormat = iota
	MigrationFormatGo
)

type DBType int

const (
	DBTypePotgreSQL DBType = iota
	DBTypeMySQL
)

type MigratorApp struct {
	MigrationsDir string
	DBType        DBType
}

type Creator interface {
	Create(migrationName string) (*os.File, error)
}

type Migrator interface {
	Connect(context.Context) error
	Up(ctx context.Context, migrations []migratorConstants.Migration) error
	Close() error
}

func New(migrationsDir string) *MigratorApp {
	return &MigratorApp{
		MigrationsDir: migrationsDir,
	}
}

func (a *MigratorApp) CreateMigration(name string, format MigrationFormat) (*os.File, error) {
	creator, err := a.getCreator(format)
	if err != nil {
		return nil, err
	}

	migrationFile, err := creator.Create(name)
	if err != nil {
		return nil, err
	}

	return migrationFile, nil
}

func (a *MigratorApp) Up() error {
	migrator, err := a.getMigrator()
	if err != nil {
		return err
	}

	migrations, err := parser.ParseMigrations(a.MigrationsDir)
	if err != nil {
		return err
	}

	ctx := context.Background()

	migrator.Connect(ctx)
	defer migrator.Close()

	err = migrator.Up(ctx, migrations)
	if err != nil {
		return err
	}

	return nil
}

func (a *MigratorApp) getCreator(format MigrationFormat) (Creator, error) {
	var creator Creator
	switch format {
	case MigrationFormatSQL:
		creator = creatorsql.New(a.MigrationsDir)
	default:
		return nil, errors.New("creator for this format not realized")
	}

	return creator, nil
}

func (a *MigratorApp) getMigrator() (Migrator, error) {
	var migrator Migrator
	switch a.DBType {
	case DBTypePotgreSQL:
		migrator = pgmigrator.New()
	default:
		return nil, errors.New("migrator for this database type not realized")
	}

	return migrator, nil
}
