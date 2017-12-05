package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/cmd"
	"github.com/zero-os/0-stor/proxy"
)

// daemonCfg represents the daemon subcommand configuration
var daemonCfg struct {
	ListenAddress string
	Config        string
}

// daemonCmd represents the daemon subcommand
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the client API as a network-connected GRPC client.",
	RunE:  daemonFunc,
}

func daemonFunc(*cobra.Command, []string) error {
	cmd.LogVersion()

	confFile, err := os.Open(daemonCfg.Config)
	if err != nil {
		return err
	}
	defer confFile.Close()

	policy, err := client.NewPolicyFromReader(confFile)
	if err != nil {
		return err
	}

	prox, err := proxy.New(policy)
	if err != nil {
		return err
	}
	return prox.Listen(daemonCfg.ListenAddress)
}

func init() {
	daemonCmd.Flags().StringVar(
		&daemonCfg.Config, "config", "config.yaml", "Configuration file.")
	daemonCmd.Flags().StringVar(
		&daemonCfg.ListenAddress, "listen", "localhost:12121", "Listen address")
}
