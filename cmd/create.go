package cmd

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	migratorApp "github.com/wursta/gomigrator/pkg/app"
)

// var goFlag bool

var createCmd = &cobra.Command{
	Use:   "create <migration-name>",
	Short: "Create a new migration file",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("argument migration-name is required")
		}

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		var migrationName string
		if len(args) == 0 {
			log.Fatal("Required arguments not passed")
		}
		migrationName = args[0]

		app := migratorApp.New(
			appConfig.MigrationsDir,
			appConfig.DBConnectionDSN,
			migratorApp.DBTypePotgreSQL,
		)

		format := migratorApp.MigrationFormatSQL
		// if goFlag {
		// 	format = migratorApp.MigrationFormatGo
		// }

		createdFile, err := app.CreateMigration(migrationName, format)
		if err != nil {
			log.Fatalf("create migration: %v", err)
		}
		log.Println("Migration file created: ", createdFile.Name())
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
	// createCmd.Flags().BoolVar(&goFlag, "go", false, "Go format for migration file")
}
