package intergationtests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
)

func GetUsingConfigFilePattern(configFilePath string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} Using config file: \.` + configFilePath
}

func GetMigrationFileCreatedPattern(migrationsDir, migrationName string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} Migration file created: .+(?P<FILENAME>` +
		migrationsDir +
		`/\d{4}_\d{2}_\d{2}T\d{2}_\d{2}_\d{2}__` + migrationName + `__[a-zA-Z]{5}.sql)`
}

func GetMigrationStepPattern(stepName, migrationName string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} ` + stepName + ` migration: ` + migrationName
}

func GetRollbackStepPattern(stepName, migrationName string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} ` + stepName + ` rollback: ` + migrationName
}

func GetMigrationStatusPattern(status, migrationName string) string {
	return `\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2} ` + status + ` | \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}: | ` + migrationName
}

func runCmd(env map[string]string, args ...string) (cmd *exec.Cmd, stdOut *bytes.Buffer, stdErr *bytes.Buffer) {
	binary := os.Getenv("GOMIGRATOR_TEST_BINARY")
	if _, err := os.Stat(binary); errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	stdOut = &bytes.Buffer{}
	stdErr = &bytes.Buffer{}

	cmd = exec.Command(binary, args...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	cmd.Env = createEnv(env)
	return
}

func execCmd(env map[string]string, args ...string) (returnCode int, stdOut *bytes.Buffer, stdErr *bytes.Buffer, err error) {
	cmd, stdOut, stdErr := runCmd(env, args...)

	if err = cmd.Start(); err != nil {
		returnCode = 1
		return
	}

	if err = cmd.Wait(); err != nil {
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
