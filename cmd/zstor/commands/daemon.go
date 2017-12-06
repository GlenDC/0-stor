package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/cmd"
	"github.com/zero-os/0-stor/daemon"
)

// daemonCfg represents the daemon subcommand configuration
var daemonCfg struct {
	ListenAddress cmd.ListenAddress
	Config        string
	MaxMsgSize    int
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

	daem, err := daemon.New(policy, daemonCfg.MaxMsgSize)
	if err != nil {
		return err
	}
	return daem.Listen(daemonCfg.ListenAddress.String())
}

func init() {
	daemonCmd.Flags().StringVar(
		&daemonCfg.Config, "config", "config.yaml", "Configuration file.")
	daemonCmd.Flags().VarP(
		&daemonCfg.ListenAddress, "listen", "L", "Bind the proxy to the given host and port. Format has to be host:port, with host optional (default :8080)")
	daemonCmd.Flags().IntVar(
		&daemonCfg.MaxMsgSize, "max-msg-size", 32, "Configure the maximum size of the message this daemon can receive, in MiB")

}
