package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"
const defaultDBPath = "default.db"

var rootCmd = &cobra.Command{
	Use:   "gograph [command] [database_path]",
	Short: "GoGraph is a pure Go embedded graph database",
	Long:  `GoGraph is a lightweight, zero-dependency embedded graph database supporting a core subset of Cypher.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := defaultDBPath
		if len(args) > 0 {
			dbPath = args[0]
		}
		return runTUI(dbPath)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(queryCmd)
}
