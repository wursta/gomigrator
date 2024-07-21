/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	migratorApp "github.com/wursta/gomigrator/pkg/app"
)

// downCmd represents the down command.
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback last success migration",
	Run: func(_ *cobra.Command, _ []string) {
		app := migratorApp.New(
			appConfig.MigrationsDir,
			appConfig.DBConnectionDSN,
			migratorApp.DBTypePotgreSQL,
		)
		err := app.Down()
		if err != nil {
			log.Fatalf("down migrations: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
