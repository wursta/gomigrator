package cmd

import (
	"log"

	"github.com/spf13/cobra"
	migratorApp "github.com/wursta/gomigrator/pkg/app"
)

// upCmd represents the up command.
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply new migrations",
	Run: func(_ *cobra.Command, _ []string) {
		app := migratorApp.New(
			appConfig.MigrationsDir,
			appConfig.DBConnectionDSN,
			migratorApp.DBTypePotgreSQL,
		)
		err := app.Up()
		if err != nil {
			log.Fatalf("up migrations: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
