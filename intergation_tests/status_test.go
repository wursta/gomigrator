package intergationtests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusSuccess(t *testing.T) {
	err := CreateDatabase("migrator_up_test")
	if err != nil {
		t.Fatal(err)
	}
	defer DropDatabase("migrator_up_test")

	db, err := Connect("postgres://test:test@localhost:5432/migrator_up_test")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	execCmd(nil, "up", "--config=./configs/up_config.yaml")

	returnCode, stdOut, stdErr := execCmd(nil, "status", "--config=./configs/up_config.yaml")
	require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

	require.Equal(t, "", stdOut.String())

	outputRegex := regexp.MustCompile(GetMigrationStatusPattern(
		"migrated",
		"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
	) + "\n" +
		GetMigrationStatusPattern(
			"migrated",
			"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
		))
	require.Regexp(t, outputRegex, stdErr)

	execCmd(nil, "down", "--config=./configs/up_config.yaml")

	returnCode, stdOut, stdErr = execCmd(nil, "status", "--config=./configs/up_config.yaml")
	require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

	outputRegex = regexp.MustCompile(GetMigrationStatusPattern(
		"new",
		"2024_07_09T20_34_36__alter_table_foo_add_column_name__oypjB.sql",
	) + "\n" +
		GetMigrationStatusPattern(
			"migrated",
			"2024_07_05T18_51_07__create_table_foo__hKnRd.sql",
		))
	require.Regexp(t, outputRegex, stdErr)
}
