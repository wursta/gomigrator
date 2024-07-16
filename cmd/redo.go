/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	migratorApp "github.com/wursta/gomigrator/pkg/app"
)

// redoCmd represents the redo command.
var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Redo last success migration",
	Run: func(_ *cobra.Command, _ []string) {
		app := migratorApp.New(
			appConfig.MigrationsDir,
			appConfig.DBConnectionDSN,
			migratorApp.DBTypePotgreSQL,
		)

		err := app.Redo()
		if err != nil {
			log.Fatalf("redo last migration: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(redoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// redoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// redoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
