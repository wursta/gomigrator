package intergationtests

import (
	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
)

func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateDatabase(databaseName string) error {
	db, err := Connect("postgres://test:test@db:5432/postgres")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE " + databaseName)
	return err
}

func DropDatabase(databaseName string) error {
	db, err := Connect("postgres://test:test@db:5432/postgres")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE " + databaseName)
	return err
}

func RunStmt(db *sqlx.DB, query string) error {
	_, err := db.Exec(query)
	return err
}

func IsTableExists(db *sqlx.DB, schemaName, tableName string) (bool, error) {
	row := db.QueryRowx(
		`SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE  table_schema = '` + schemaName + `'
			AND    table_name   = '` + tableName + `'
		) as exists`,
	)

	var tableExists bool
	err := row.Scan(&tableExists)
	if err != nil {
		return false, err
	}

	return tableExists, nil
}

func IsColumnExists(db *sqlx.DB, schemaName, tableName, columnName string) (bool, error) {
	row := db.QueryRowx(
		`SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE  table_schema = '` + schemaName + `'
			AND    table_name   = '` + tableName + `'
			AND    column_name   = '` + columnName + `'
		) as exists`,
	)

	var columnExists bool
	err := row.Scan(&columnExists)
	if err != nil {
		return false, err
	}

	return columnExists, nil
}

func IsMigrationRegistered(db *sqlx.DB, migrationName, migrationStatus string) (bool, error) {
	row := db.QueryRowx(
		`SELECT count(id) 
		FROM public.dbmigrations WHERE file_name = '` + migrationName + `' and status = '` + migrationStatus + `'`,
	)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 1, nil
}
