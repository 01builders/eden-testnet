package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	rollcmd "github.com/evstack/ev-node/pkg/cmd"
	"github.com/evstack/ev-node/pkg/config"

	"github.com/01builders/eden-testnet/cmd"
)

func main() {
	// Initiate the root command
	rootCmd := &cobra.Command{
		Use:   "eden-testnet",
		Short: "Eden Testnet Single Sequencer",
	}

	config.AddGlobalFlags(rootCmd, "eden-testnet")

	// Add configuration flags to NetInfoCmd so it can read RPC address
	config.AddFlags(rollcmd.NetInfoCmd)

	rootCmd.AddCommand(
		cmd.InitCmd(),
		cmd.RunCmd,
		cmd.NewRollbackCmd(),
		rollcmd.VersionCmd,
		rollcmd.NetInfoCmd,
		rollcmd.StoreUnsafeCleanCmd,
		rollcmd.KeysCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		// Print to stderr and exit with error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
