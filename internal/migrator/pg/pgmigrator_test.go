package pgmigrator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/wursta/gomigrator/internal/migrator"
)

func TestUpSuccess(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New()
	pgMigrator.db = db

	queries := []string{"Test Query 1", "Test Query 2", "Test Query 3"}
	migrations := make([]migrator.Migration, 3)
	ctx := context.Background()
	for i := range queries {
		migrations[i] = migrator.Migration{
			UpHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, queries[i])
				fmt.Println(err)
				if err != nil {
					return err
				}

				return nil
			},
		}
	}

	mock.ExpectBegin()
	mock.ExpectExec("Test Query 1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("Test Query 2").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("Test Query 3").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	pgMigrator.Up(ctx, migrations)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpFailure(t *testing.T) {
	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	pgMigrator := New()
	pgMigrator.db = db

	queries := []string{"Test Query 1", "Test Query 2", "Test Query 3"}
	migrations := make([]migrator.Migration, 3)
	ctx := context.Background()
	for i := range queries {
		migrations[i] = migrator.Migration{
			UpHandlerContext: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, queries[i])
				fmt.Println(err)
				if err != nil {
					return err
				}

				return nil
			},
		}
	}

	mock.ExpectBegin()
	mock.ExpectExec("Test Query 1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("Test Query 2").WillReturnError(errors.New("some DB error"))
	mock.ExpectRollback()

	pgMigrator.Up(ctx, migrations)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func getDBMock() (*sqlx.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock, nil
}
