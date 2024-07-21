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
	// MigrationFormatGo.
)

type DBType int

const (
	DBTypePotgreSQL DBType = iota
)

type MigratorApp struct {
	MigrationsDir   string
	DBType          DBType
	DBConnectionDSN string
}

type Creator interface {
	Create(migrationName string) (*os.File, error)
}

type Migrator interface {
	Connect(context.Context) error
	Init(ctx context.Context) error
	Up(ctx context.Context, migrations []migratorConstants.Migration) error
	Down(ctx context.Context, migrations []migratorConstants.Migration) error
	GetLastMigations(
		ctx context.Context,
		status migratorConstants.MigrationStatus,
		count int,
	) ([]migratorConstants.Migration, error)
	GetMigations(ctx context.Context) ([]migratorConstants.Migration, error)
	Close() error
}

func New(migrationsDir, dbConnectionDSN string, dbType DBType) *MigratorApp {
	return &MigratorApp{
		MigrationsDir:   migrationsDir,
		DBType:          dbType,
		DBConnectionDSN: dbConnectionDSN,
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

	err = migrator.Connect(ctx)
	if err != nil {
		return err
	}
	defer migrator.Close()

	err = migrator.Init(ctx)
	if err != nil {
		return err
	}

	err = migrator.Up(ctx, migrations)
	if err != nil {
		return err
	}

	return nil
}

func (a *MigratorApp) Down() error {
	migrator, err := a.getMigrator()
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = migrator.Connect(ctx)
	if err != nil {
		return err
	}
	defer migrator.Close()

	err = migrator.Init(ctx)
	if err != nil {
		return err
	}

	migrations, err := migrator.GetLastMigations(ctx, migratorConstants.MigrationStatusMigrated, 1)
	if err != nil {
		return err
	}

	for i := range migrations {
		upHandler, downHandler, err := parser.GetMigrationFileHandlers(a.MigrationsDir, migrations[i].Name)
		if err != nil {
			return err
		}
		migrations[i].UpHandlerContext = upHandler
		migrations[i].DownHandlerContext = downHandler
	}

	err = migrator.Down(ctx, migrations)
	if err != nil {
		return err
	}

	return nil
}

func (a *MigratorApp) Redo() error {
	migrator, err := a.getMigrator()
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = migrator.Connect(ctx)
	if err != nil {
		return err
	}
	defer migrator.Close()

	err = migrator.Init(ctx)
	if err != nil {
		return err
	}

	migrations, err := migrator.GetLastMigations(ctx, migratorConstants.MigrationStatusMigrated, 1)
	if err != nil {
		return err
	}

	for i := range migrations {
		upHandler, downHandler, err := parser.GetMigrationFileHandlers(a.MigrationsDir, migrations[i].Name)
		if err != nil {
			return err
		}
		migrations[i].UpHandlerContext = upHandler
		migrations[i].DownHandlerContext = downHandler
	}

	err = migrator.Down(ctx, migrations)
	if err != nil {
		return err
	}

	err = migrator.Up(ctx, migrations)
	if err != nil {
		return err
	}

	return nil
}

func (a *MigratorApp) Status() ([]migratorConstants.Migration, error) {
	migrator, err := a.getMigrator()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	err = migrator.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer migrator.Close()

	err = migrator.Init(ctx)
	if err != nil {
		return nil, err
	}

	migrations, err := migrator.GetMigations(ctx)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

func (a *MigratorApp) GetDBVersion() (uint, error) {
	migrator, err := a.getMigrator()
	if err != nil {
		return 0, err
	}

	ctx := context.Background()

	err = migrator.Connect(ctx)
	if err != nil {
		return 0, err
	}
	defer migrator.Close()

	err = migrator.Init(ctx)
	if err != nil {
		return 0, err
	}

	migrations, err := migrator.GetLastMigations(ctx, migratorConstants.MigrationStatusMigrated, 1)
	if err != nil {
		return 0, err
	}

	if len(migrations) == 0 {
		return 0, nil
	}

	return *migrations[0].ID, nil
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
		migrator = pgmigrator.New(a.DBConnectionDSN)
	default:
		return nil, errors.New("migrator for this database type not realized")
	}

	return migrator, nil
}
