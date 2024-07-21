package intergationtests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDBVersionSuccess(t *testing.T) {
	err := CreateDatabase("migrator_up_test")
	if err != nil {
		t.Fatal(err)
	}
	defer DropDatabase("migrator_up_test")

	db, err := Connect("postgres://test:test@db:5432/migrator_up_test")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, _, _, err = execCmd(nil, "up", "--config=./configs/up_config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	returnCode, stdOut, stdErr, err := execCmd(nil, "dbversion", "--config=./configs/up_config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

	require.Equal(t, "", stdOut.String())
	outputRegex := regexp.MustCompile(`\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} Version: 2`)
	require.Regexp(t, outputRegex, stdErr.String())

	execCmd(nil, "down", "--config=./configs/up_config.yaml")

	returnCode, stdOut, stdErr, err = execCmd(nil, "dbversion", "--config=./configs/up_config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

	require.Equal(t, "", stdOut.String())
	outputRegex = regexp.MustCompile(`\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} Version: 1`)
	require.Regexp(t, outputRegex, stdErr.String())
}
