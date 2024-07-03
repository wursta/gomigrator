package parser

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestParseMigrationsSuccess(t *testing.T) {
	migrations, err := ParseMigrations("./test/good_migrations")
	require.Nil(t, err)
	require.Len(t, migrations, 4)

	require.Equal(t, "2024_07_01T18_04_42__create_foo_table__wyvmi.sql", migrations[0].Name)
	require.Equal(t, "2024_07_01T21_02_39__create_bar_table__dlfio.sql", migrations[1].Name)
	require.Equal(t, "2024_07_03T21_23_16__alter_bar_table_add_column_name__flxZZ.sql", migrations[2].Name)
	require.Equal(t, "2024_07_03T21_27_21__alter_foo_table_add_column_name__wGYSd.sql", migrations[3].Name)

	db, mock, err := getDBMock()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()

	// migration 0
	mock.ExpectExec(`CREATE TABLE public.foo(
    id SERIAL    
);`)
	mock.ExpectExec("DROP TABLE public.foo;")

	// migration 1
	mock.ExpectExec(`CREATE TABLE public.bar(
		id SERIAL    
	);`)
	mock.ExpectExec("DROP TABLE public.bar;")

	// migration 2
	mock.ExpectExec("ALTER TABLE public.bar ADD COLUMN name varchar(255);")
	mock.ExpectExec("ALTER TABLE public.bar DROP COLUMN name;")

	// migration 3
	mock.ExpectExec("ALTER TABLE public.foo ADD COLUMN name varchar(255);")
	mock.ExpectExec("ALTER TABLE public.foo DROP COLUMN name;")

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	for i := range migrations {
		migrations[i].UpHandlerContext(ctx, tx)
		migrations[i].DownHandlerContext(ctx, tx)
	}

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
