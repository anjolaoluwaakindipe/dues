/*
Copyright © 2024 The Dues Authors
*/
package cmd

import (
	"os"

	"github.com/anjolaoluwaakindipe/dues/internal/log"
	dues "github.com/anjolaoluwaakindipe/dues/pkg"
	"github.com/spf13/cobra"
)

var (
	configPath = "dues.json"
	rootCmd    = &cobra.Command{
		Use:           "dues",
		Short:         "A live reloading application made to handle multiple tasks concurrently",
		Long:          ``,
		Args:          cobra.MatchAll(cobra.MinimumNArgs(1)),
		RunE:          rootRun,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVar(&configPath, "config", configPath, "Your dues config path.")
}

// root command execution
func rootRun(cmd *cobra.Command, args []string) error {
	config := dues.DuesConfig{
		Commands:   args,
		ConfigPath: configPath,
	}

	return dues.RunDues(config)
}
