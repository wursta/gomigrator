package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetMigrationFileName(t *testing.T) {
	now := time.Now()
	migrationNameStart := now.Format(MigrationFileDateTime)

	tests := []struct {
		createDatetime        time.Time
		migrationName         string
		migrationNameSuffix   string
		expectedMigrationName string
	}{
		{
			createDatetime:        now,
			migrationName:         "create_some_table",
			migrationNameSuffix:   "abcde",
			expectedMigrationName: migrationNameStart + "__create_some_table__abcde",
		},
		{
			createDatetime:        now,
			migrationName:         "create some table",
			migrationNameSuffix:   "cdefg",
			expectedMigrationName: migrationNameStart + "__create some table__cdefg",
		},
		{
			createDatetime:        now,
			migrationName:         "",
			migrationNameSuffix:   "",
			expectedMigrationName: migrationNameStart + "____",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			testCase := tt
			t.Parallel()

			migrationName := GetMigrationFileName(testCase.createDatetime, testCase.migrationName, testCase.migrationNameSuffix)
			require.Equal(t, testCase.expectedMigrationName, migrationName)
		})
	}
}

func TestCreateMigrationFile(t *testing.T) {
	now := time.Now()
	migrationNameStart := now.Format(MigrationFileDateTime)

	tests := []struct {
		createDatetime            time.Time
		migrationName             string
		migrationNameSuffix       string
		expectedMigrationFileName string
	}{
		{
			createDatetime:            now,
			migrationName:             "create_some_table",
			migrationNameSuffix:       "abcde",
			expectedMigrationFileName: migrationNameStart + "__create_some_table__abcde.sql",
		},
		{
			createDatetime:            now,
			migrationName:             "create some table",
			migrationNameSuffix:       "cdefg",
			expectedMigrationFileName: migrationNameStart + "__create some table__cdefg.sql",
		},
		{
			createDatetime:            now,
			migrationName:             "",
			migrationNameSuffix:       "",
			expectedMigrationFileName: migrationNameStart + "____.sql",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			testCase := tt
			t.Parallel()

			migrationFile, err := CreateMigrationFile(
				testCase.createDatetime,
				"./test/migrations",
				testCase.migrationName,
				testCase.migrationNameSuffix,
				"sql",
			)
			migrationFile.Close()
			defer os.Remove(migrationFile.Name())

			require.Nil(t, err)
			require.Equal(t, testCase.expectedMigrationFileName, filepath.Base(migrationFile.Name()))
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	for i := 0; i <= 5; i++ {
		str := GenerateRandomString(uint8(i))
		pattern := regexp.MustCompile(`^[a-zA-Z]{` + strconv.Itoa(i) + `}$`)
		require.Regexp(t, pattern, str)
	}
}
