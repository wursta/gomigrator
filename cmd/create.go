package cmd

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	migratorApp "github.com/wursta/gomigrator/pkg/app"
)

var createCmd = &cobra.Command{
	Use:   "create <migration-name>",
	Short: "Creates a new migration",
	Long:  `Log description...`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("argument migration-name is required")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var migrationName string
		if len(args) == 0 {
			log.Fatal("Required arguments not passed")
		}

		migrationName = args[0]

		migrationsDir, err := cmd.Flags().GetString("migrations-dir")
		if err != nil {
			log.Fatalf("get \"migrations-dir\" flag error: %v", err)
		}

		app := migratorApp.New(migrationsDir)

		var format migratorApp.MigrationFormat
		sqlFlag, err := cmd.Flags().GetBool("sql")
		if err != nil {
			log.Fatalf("get \"sql\" flag error: %v", err)
		}
		goFlag, err := cmd.Flags().GetBool("go")
		if err != nil {
			log.Fatalf("get \"go\" flag error: %v", err)
		}
		if sqlFlag && goFlag {
			log.Fatal("flags \"sql\" and \"go\" cannot be applied at the same time")
		}

		if sqlFlag {
			format = migratorApp.MigrationFormatSQL
		} else if goFlag {
			format = migratorApp.MigrationFormatGo
		} else {
			log.Fatal("unknown migration file format")
		}
		_, err = app.CreateMigration(migrationName, format)
		if err != nil {
			log.Fatalf("create migration: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createCmd.Flags().Bool("sql", true, "SQL format for migration file")
	createCmd.Flags().Bool("go", false, "Go format for migration file")
}
