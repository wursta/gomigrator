/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	migratorApp "github.com/wursta/gomigrator/pkg/app"
)

// dbversionCmd represents the dbversion command.
var dbversionCmd = &cobra.Command{
	Use:   "dbversion",
	Short: "Get current database version",
	Run: func(_ *cobra.Command, _ []string) {
		app := migratorApp.New(
			appConfig.MigrationsDir,
			appConfig.DBConnectionDSN,
			migratorApp.DBTypePotgreSQL,
		)
		dbVersion, err := app.GetDBVersion()
		if err != nil {
			log.Fatalf("error while get databse version: %v", err)
		}
		log.Println("Version: ", dbVersion)
	},
}

func init() {
	rootCmd.AddCommand(dbversionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dbversionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dbversionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
