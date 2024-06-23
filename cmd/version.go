package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	release   string
	buildDate string
	gitHash   string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		if err := json.NewEncoder(os.Stdout).Encode(struct {
			Release   string
			BuildDate string
			GitHash   string
		}{
			Release:   release,
			BuildDate: buildDate,
			GitHash:   gitHash,
		}); err != nil {
			fmt.Printf("error while decode version info: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
