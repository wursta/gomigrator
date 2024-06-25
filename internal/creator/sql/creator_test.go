package creatorsql_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	creatorsql "github.com/wursta/gomigrator/internal/creator/sql"
)

func TestCreateSuccess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test_migration",
		},
		{
			name: "test migration",
		},
		{
			name: "create test table",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			testCase := tt
			t.Parallel()

			c := creatorsql.New("./test/migrations")
			f, err := c.Create(testCase.name)
			if err == nil {
				defer os.Remove(f.Name())
			}

			require.Nil(t, err)

			pattern := regexp.MustCompile(`^.*/\d{4}_\d{2}_\d{2}T\d{2}_\d{2}_\d{2}__` + testCase.name + `__[a-zA-Z]{5}.sql$`)
			require.Regexp(t, pattern, f.Name())
		})
	}
}

func TestCreateFailure(t *testing.T) {
	tests := map[string]struct {
		name          string
		migrationsDir string
		errPattern    *regexp.Regexp
	}{
		"non-existent dir": {
			name:          "test_migration",
			migrationsDir: "./test/non_existent_dir",
			errPattern: regexp.MustCompile(
				`create migration file: open .*\d{4}_\d{2}_\d{2}T\d{2}_\d{2}_\d{2}__test_migration__[a-zA-Z]{5}.sql: .*`,
			),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %s", i), func(t *testing.T) {
			testCase := tt

			t.Parallel()

			c := creatorsql.New(testCase.migrationsDir)
			_, err := c.Create(testCase.name)

			require.Regexp(t, testCase.errPattern, err)
		})
	}
}
