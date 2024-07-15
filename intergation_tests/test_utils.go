package intergationtests

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
