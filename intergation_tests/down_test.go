package intergationtests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestDownSuccess(t *testing.T) {
	tests := map[string]struct {
		cmdFlags     []string
		envVars      map[string]string
		dsn          string
		databaseName string
	}{
		"config flag": {
			cmdFlags:     []string{"--config=./configs/down_config.yaml"},
			dsn:          "postgres://test:test@db:5432/migrator_down_test",
			databaseName: "migrator_down_test",
		},
		"db-dsn flag": {
			cmdFlags: []string{
				"--migrations-dir=./migrations_down",
				"--db-dsn=postgres://test:test@db:5432/migrator_down_dsn_flag_test",
			},
			dsn:          "postgres://test:test@db:5432/migrator_down_dsn_flag_test",
			databaseName: "migrator_down_dsn_flag_test",
		},
		"db-dsn env": {
			envVars: map[string]string{
				"GOMIGRATOR_MIGRATIONS_DIR": "migrations_down",
				"GOMIGRATOR_DB_DSN":         "postgres://test:test@db:5432/migrator_down_dsn_flag_test",
			},
			dsn:          "postgres://test:test@db:5432/migrator_down_dsn_flag_test",
			databaseName: "migrator_down_dsn_flag_test",
		},
	}

	for tName, tt := range tests {
		t.Run(fmt.Sprintf("case %s", tName), func(t *testing.T) {
			testCase := tt

			err := CreateDatabase(testCase.databaseName)
			if err != nil {
				t.Fatal(err)
			}
			defer DropDatabase(testCase.databaseName)

			db, err := Connect(testCase.dsn)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			err = initDBTestData(db)
			if err != nil {
				t.Fatal(err)
			}

			cmdArgs := []string{"down"}
			cmdArgs = append(cmdArgs, testCase.cmdFlags...)
			returnCode, stdOut, stdErr, err := execCmd(testCase.envVars, cmdArgs...)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

			require.Equal(t, "", stdOut.String())

			outputRegex := regexp.MustCompile(GetRollbackStepPattern(
				"Start",
				"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
			) + "\n" +
				GetRollbackStepPattern(
					"Success",
					"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
				) + "\n")
			require.Regexp(t, outputRegex, stdErr)

			tableExists, err := IsTableExists(db, "public", "foo")
			if err != nil {
				t.Fatal(err)
			}
			require.True(t, tableExists)

			columnExists, err := IsColumnExists(db, "public", "foo", "name")
			if err != nil {
				t.Fatal(err)
			}
			require.False(t, columnExists)

			migrationRegistered, err := IsMigrationRegistered(
				db,
				"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
				"new",
			)
			if err != nil {
				t.Fatal(err)
			}
			require.True(t, migrationRegistered)

			migrationRegistered, err = IsMigrationRegistered(
				db,
				"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
				"migrated",
			)
			if err != nil {
				t.Fatal(err)
			}
			require.True(t, migrationRegistered)

			cmdArgs = []string{"down"}
			cmdArgs = append(cmdArgs, testCase.cmdFlags...)
			returnCode, stdOut, stdErr, err = execCmd(testCase.envVars, cmdArgs...)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

			require.Equal(t, "", stdOut.String())
			outputRegex = regexp.MustCompile(GetRollbackStepPattern(
				"Start",
				"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
			) + "\n" +
				GetRollbackStepPattern(
					"Success",
					"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
				) + "\n")
			require.Regexp(t, outputRegex, stdErr)

			tableExists, err = IsTableExists(db, "public", "foo")
			if err != nil {
				t.Fatal(err)
			}
			require.False(t, tableExists)

			migrationRegistered, err = IsMigrationRegistered(
				db,
				"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
				"new",
			)
			if err != nil {
				t.Fatal(err)
			}
			require.True(t, migrationRegistered)
		})
	}
}

func initDBTestData(db *sqlx.DB) error {
	initDBStatements := []string{
		"CREATE TABLE public.foo(id SERIAL)",
		"ALTER TABLE public.foo ADD COLUMN name varchar(255)",
		`CREATE TABLE public.dbmigrations (
			id SERIAL NOT NULL,
			file_path VARCHAR(255) NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			status VARCHAR(10) NOT NULL,
			create_dt timestamp NOT NULL,
			migrate_dt timestamp NOT NULL,
			CONSTRAINT dbmigrations_pk PRIMARY KEY (id)
		)`,
		`CREATE UNIQUE INDEX unq_dbmigrations_file_name ON public.dbmigrations (file_name)`,
		`INSERT INTO public.dbmigrations (file_path, file_name, status, create_dt, migrate_dt) 
		  VALUES (
			 '/somepath/2024_07_05T18_51_07__create_table_foo__hKnRd.sql', 
			'2024_07_05T18_51_07__create_table_foo__hKnRd.sql', 
			'migrated', 
			NOW(), 
			NOW()
		 )`,
		`INSERT INTO public.dbmigrations (file_path, file_name, status, create_dt, migrate_dt) 
		  VALUES (
			 '/somepath/2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql', 
			'2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql', 
			'migrated', 
			NOW(), 
			NOW()
		 )`,
	}

	for i := range initDBStatements {
		err := RunStmt(db, initDBStatements[i])
		if err != nil {
			return err
		}
	}

	return nil
}
