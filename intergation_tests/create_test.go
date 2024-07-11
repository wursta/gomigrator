package intergationtests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func getUsingConfigFilePattern(configFilePath string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} Using config file: \.` + configFilePath
}

func getMigrationFileCreatedPattern(migrationsDir, migrationName string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} Migration file created: .+(?P<FILENAME>` +
		migrationsDir +
		`/\d{4}_\d{2}_\d{2}T\d{2}_\d{2}_\d{2}__` + migrationName + `__[a-zA-Z]{5}.sql)`
}

func TestCreateSuccess(t *testing.T) {
	tests := map[string]struct {
		cmdFlags      []string
		envVars       map[string]string
		migrationName string
		migrationsDir string
		regex         *regexp.Regexp
	}{
		"config flag": {
			cmdFlags:      []string{"--config=./configs/config.yaml"},
			migrationName: "create_foo_table",
			migrationsDir: "migrations",
			regex: regexp.MustCompile("^" +
				getUsingConfigFilePattern(`/configs/config\.yaml`) +
				"\n" +
				getMigrationFileCreatedPattern("migrations", "create_foo_table")),
		},
		"migrations-dir flag": {
			cmdFlags:      []string{"--migrations-dir=./migrations_flag_test"},
			migrationName: "create_test_table",
			migrationsDir: "migrations_flag_test",
			regex: regexp.MustCompile("^" +
				getMigrationFileCreatedPattern("migrations_flag_test", "create_test_table")),
		},
		"migrations-dir env": {
			envVars:       map[string]string{"GOMIGRATOR_MIGRATIONS_DIR": "migrations_env_test"},
			migrationName: "create_bar_table",
			migrationsDir: "migrations_env_test",
			regex: regexp.MustCompile("^" +
				getMigrationFileCreatedPattern("migrations_env_test", "create_bar_table")),
		},
	}

	for tName, tt := range tests {
		t.Run(fmt.Sprintf("case %s", tName), func(t *testing.T) {
			testCase := tt

			err := createMigrationsDir(testCase.migrationsDir, 0o755)
			if err != nil {
				t.Error(err)
			}
			defer clearMigrationsDir(testCase.migrationsDir)

			cmdArgs := []string{"create", testCase.migrationName}
			cmdArgs = append(cmdArgs, testCase.cmdFlags...)
			returnCode, stdOut, stdErr := execCmd(testCase.envVars, cmdArgs...)
			require.Equal(t, 0, returnCode, fmt.Sprintf("stdout: %s\nstderr: %s", stdOut, stdErr))

			require.Equal(t, "", stdOut.String())
			require.Regexp(t, testCase.regex, stdErr.String())
		})
	}
}

func runCmd(env map[string]string, args ...string) (cmd *exec.Cmd, stdOut *bytes.Buffer, stdErr *bytes.Buffer) {
	stdOut = &bytes.Buffer{}
	stdErr = &bytes.Buffer{}

	cmd = exec.Command("./bin/gomigrator", args...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	cmd.Env = createEnv(env)
	return
}

func execCmd(env map[string]string, args ...string) (returnCode int, stdOut *bytes.Buffer, stdErr *bytes.Buffer) {
	cmd, stdOut, stdErr := runCmd(env, args...)

	if err := cmd.Start(); err != nil {
		returnCode = 1
		return
	}

	if err := cmd.Wait(); err != nil {
		var exitErrType *exec.ExitError
		if errors.As(err, &exitErrType) {
			returnCode = exitErrType.ExitCode()
		} else {
			returnCode = 1
		}
	}

	return
}

func createEnv(env map[string]string) []string {
	envStrings := make([]string, 0, len(env))
	for key, val := range env {
		envStrings = append(envStrings, key+"="+val)
	}
	return envStrings
}

func createMigrationsDir(dirName string, perm os.FileMode) error {
	return os.Mkdir(dirName, perm)
}

func clearMigrationsDir(dirName string) {
	os.RemoveAll(dirName)
}