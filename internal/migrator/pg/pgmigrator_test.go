package pgmigrator

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/wursta/gomigrator/internal/migrator"
)

func TestUpSuccess(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New("testdsn")
	pgMigrator.db = db

	ctx := context.Background()
	migrations := make([]migrator.Migration, 3)
	for i := 0; i < 3; i++ {
		migrations[i] = migrator.Migration{
			FilePath: "./migrations/" + fmt.Sprintf("migration_%v.sql", i),
			Name:     fmt.Sprintf("migration_%v.sql", i),
			UpHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, fmt.Sprint("Test Query ", i))
				if err != nil {
					return err
				}

				return nil
			},
		}
	}

	for i := 0; i < 3; i++ {
		migrationFileName := fmt.Sprintf("migration_%v.sql", i)
		migrationFilePath := "./migrations/" + migrationFileName

		expectStatusCheckError(mock, migrationFileName)
		expectLock(mock, migrationFileName)
		expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrating)
		mock.ExpectBegin()
		mock.ExpectExec(fmt.Sprint("Test Query ", i)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
		expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrated)
		expectUnlock(mock, migrationFileName)
	}

	err = pgMigrator.Up(ctx, migrations)
	require.Nil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpFailure(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New("testdsn")
	pgMigrator.db = db

	ctx := context.Background()
	migrations := make([]migrator.Migration, 3)
	for i := 0; i < 3; i++ {
		migrations[i] = migrator.Migration{
			FilePath: "./migrations/" + fmt.Sprintf("migration_%v.sql", i),
			Name:     fmt.Sprintf("migration_%v.sql", i),
			UpHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, fmt.Sprint("Test Query ", i))
				if err != nil {
					return err
				}

				return nil
			},
		}
	}

	migrationFileName := "migration_0.sql"
	migrationFilePath := "./migrations/migration_0.sql"

	expectStatusCheckError(mock, migrationFileName)
	expectLock(mock, migrationFileName)
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrating)
	mock.ExpectBegin()
	mock.ExpectExec("Test Query 0").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrated)
	expectUnlock(mock, migrationFileName)

	migrationFileName = "migration_1.sql"
	migrationFilePath = "./migrations/migration_1.sql"

	expectStatusCheckError(mock, migrationFileName)
	expectLock(mock, migrationFileName)
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrating)
	mock.ExpectBegin()
	mock.ExpectExec("Test Query 1").WillReturnError(errors.New("some DB error"))
	mock.ExpectRollback()
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusFailed)
	expectUnlock(mock, migrationFileName)

	err = pgMigrator.Up(ctx, migrations)
	require.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDownSuccess(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New("testdsn")
	pgMigrator.db = db

	migrationFileName := "migration.sql"
	migrationFilePath := "./migrations/migration.sql"
	migrations := []migrator.Migration{
		{
			FilePath: migrationFilePath,
			Name:     migrationFileName,
			DownHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, "Test Query")
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	expectLock(mock, migrationFileName)
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrating)
	mock.ExpectBegin()
	mock.ExpectExec("Test Query").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusNew)
	expectUnlock(mock, migrationFileName)

	ctx := context.Background()
	err = pgMigrator.Down(ctx, migrations)
	require.Nil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDownFailure(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New("testdsn")
	pgMigrator.db = db

	migrationFileName := "migration.sql"
	migrationFilePath := "./migrations/migration.sql"
	migrations := []migrator.Migration{
		{
			FilePath: migrationFilePath,
			Name:     migrationFileName,
			DownHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, "Test Query")
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	expectLock(mock, migrationFileName)
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrating)
	mock.ExpectBegin()
	mock.ExpectExec("Test Query").WillReturnError(errors.New("some DB error"))
	mock.ExpectRollback()
	expectUpdateMigrationStatus(mock, migrationFilePath, migrationFileName, migrator.MigrationStatusMigrated)
	expectUnlock(mock, migrationFileName)

	ctx := context.Background()
	err = pgMigrator.Down(ctx, migrations)
	require.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetLastMigations(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New("testdsn")
	pgMigrator.db = db

	createdDT := time.Now()
	migratedDT := time.Now()
	mock.
		ExpectQuery(`SELECT id, file_path, file_name, status, create_dt, migrate_dt 
		FROM public.dbmigrations 
		WHERE status = ?
		ORDER BY id DESC 
		LIMIT ?`).
		WithArgs("migrated", 1).
		WillReturnRows(
			sqlmock.
				NewRowsWithColumnDefinition(
					sqlmock.NewColumn("id").OfType("INTEGER", 0).Nullable(false),
					sqlmock.NewColumn("file_path").OfType("VARCHAR", "").Nullable(false).WithLength(255),
					sqlmock.NewColumn("file_name").OfType("VARCHAR", "").Nullable(false).WithLength(255),
					sqlmock.NewColumn("status").OfType("VARCHAR", "").Nullable(false).WithLength(255),
					sqlmock.NewColumn("create_dt").OfType("TIMESTAMP", time.Now()),
					sqlmock.NewColumn("migrate_dt").OfType("TIMESTAMP", time.Now()),
				).
				AddRow(
					1,
					"./migrations/migration_1.sql",
					"migration_1.sql",
					migrator.MigrationStatusMigrated,
					createdDT,
					migratedDT,
				),
		)

	ctx := context.Background()
	migrations, err := pgMigrator.GetLastMigations(ctx, 1)
	require.Nil(t, err)
	require.Len(t, migrations, 1)
	require.Equal(t, uint(1), *migrations[0].ID)
	require.Equal(t, "./migrations/migration_1.sql", migrations[0].FilePath)
	require.Equal(t, "migration_1.sql", migrations[0].Name)
	require.Equal(t, migrator.MigrationStatus("migrated"), migrations[0].Status)
	require.Equal(t, createdDT, migrations[0].CreateDT)
	require.Equal(t, &migratedDT, migrations[0].MigrateDT)
}

func getDBMock() (*sqlx.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock, nil
}

func expectStatusCheckError(mock sqlmock.Sqlmock, migrationFileName string) {
	mock.
		ExpectQuery("SELECT status FROM public.dbmigrations WHERE file_name = ?").
		WithArgs(migrationFileName).
		WillReturnError(errors.New("no rows"))
}

func expectLock(mock sqlmock.Sqlmock, migrationFileName string) {
	mock.
		ExpectExec("SELECT pg_advisory_lock(id) FROM public.dbmigrations WHERE file_name = ?").
		WithArgs(migrationFileName).
		WillReturnResult(sqlmock.NewResult(0, 0))
}

func expectUnlock(mock sqlmock.Sqlmock, migrationFileName string) {
	mock.
		ExpectExec("SELECT pg_advisory_unlock(id) FROM public.dbmigrations WHERE file_name = ?").
		WithArgs(migrationFileName).
		WillReturnResult(sqlmock.NewResult(0, 0))
}

func expectUpdateMigrationStatus(
	mock sqlmock.Sqlmock,
	migrationFilePath,
	migrationFileName string,
	status migrator.MigrationStatus,
) {
	mock.
		ExpectExec(`INSERT INTO public.dbmigrations (file_path, file_name, status, create_dt, migrate_dt) 
		 VALUES (?, ?, ?, NOW(), NOW())
		 ON CONFLICT (file_name) DO UPDATE SET
		 status = ?,
		 migrate_dt = NOW()`).
		WithArgs(migrationFilePath, migrationFileName, status, status).
		WillReturnResult(sqlmock.NewResult(1, 1))
}
