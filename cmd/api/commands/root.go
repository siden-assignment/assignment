package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/supernomad/siden/pkg/api"
	"github.com/supernomad/siden/pkg/store"
)

var (
	address  string
	stateDir string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&address, "listen-address", "l", ":8080", "The address for the API to listen on for incoming HTTP requests.")
	rootCmd.PersistentFlags().StringVarP(&stateDir, "state-directory", "d", "./dist/db", "The directory to store data to.")
}

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "A REST API to fulfill the Siden coding assignment.",
	Long:  `A fast and intuative API for filtering duplicate strings from a supplied file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &api.Config{
			Address:   address,
			Directory: stateDir,
			StoreKind: store.BADGER,
		}
		api, err := api.New(cfg)
		if err != nil {
			return err
		}
		return api.Listen()
	},
}

// Execute handles executing the root cobra command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Failed to execute API", err)
		os.Exit(1)
	}
}
