package cmd

type Config struct {
	MigrationsDir   string `mapstructure:"migrations_dir"`
	DBConnectionDSN string `mapstructure:"db_dsn"`
}
