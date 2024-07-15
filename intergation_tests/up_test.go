package intergationtests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpSuccess(t *testing.T) {
	tests := map[string]struct {
		cmdFlags     []string
		envVars      map[string]string
		dsn          string
		databaseName string
	}{
		"config flag": {
			cmdFlags:     []string{"--config=./configs/up_config.yaml"},
			dsn:          "postgres://test:test@localhost:5432/migrator_up_test",
			databaseName: "migrator_up_test",
		},
		"db-dsn flag": {
			cmdFlags: []string{
				"--migrations-dir=./migrations_up",
				"--db-dsn=postgres://test:test@localhost:5432/migrator_up_dsn_flag_test",
			},
			dsn:          "postgres://test:test@localhost:5432/migrator_up_dsn_flag_test",
			databaseName: "migrator_up_dsn_flag_test",
		},
		"db-dsn env": {
			envVars: map[string]string{
				"GOMIGRATOR_MIGRATIONS_DIR": "migrations_up",
				"GOMIGRATOR_DB_DSN":         "postgres://test:test@localhost:5432/migrator_up_dsn_env_test",
			},
			dsn:          "postgres://test:test@localhost:5432/migrator_up_dsn_env_test",
			databaseName: "migrator_up_dsn_env_test",
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

			cmdArgs := []string{"up"}
			cmdArgs = append(cmdArgs, testCase.cmdFlags...)
			returnCode, stdOut, stdErr := execCmd(testCase.envVars, cmdArgs...)
			require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

			require.Equal(t, "", stdOut.String())

			outputRegex := regexp.MustCompile(GetMigrationStepPattern(
				"Start",
				"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
			) + "\n" +
				GetMigrationStepPattern(
					"Success",
					"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
				) + "\n" +
				GetMigrationStepPattern(
					"Start",
					"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
				) + "\n" +
				GetMigrationStepPattern(
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
			require.True(t, columnExists)

			migrationRegistered, err := IsMigrationRegistered(
				db,
				"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
				"migrated",
			)
			if err != nil {
				t.Fatal(err)
			}
			require.True(t, migrationRegistered)

			migrationRegistered, err = IsMigrationRegistered(
				db,
				"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
				"migrated",
			)
			if err != nil {
				t.Fatal(err)
			}
			require.True(t, migrationRegistered)
		})
	}
}
