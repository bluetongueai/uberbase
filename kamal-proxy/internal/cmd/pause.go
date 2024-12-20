package cmd

import (
	"net/rpc"

	"github.com/spf13/cobra"

	"github.com/basecamp/kamal-proxy/internal/server"
)

type pauseCommand struct {
	cmd  *cobra.Command
	args server.PauseArgs
}

func newPauseCommand() *pauseCommand {
	pauseCommand := &pauseCommand{}
	pauseCommand.cmd = &cobra.Command{
		Use:       "pause <service>",
		Short:     "Pause a service",
		RunE:      pauseCommand.run,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"service"},
	}

	pauseCommand.cmd.Flags().DurationVar(&pauseCommand.args.DrainTimeout, "drain-timeout", server.DefaultDrainTimeout, "How long to allow in-flight requests to complete")
	pauseCommand.cmd.Flags().DurationVar(&pauseCommand.args.PauseTimeout, "max-pause", server.DefaultPauseTimeout, "How long to enqueue requests before shedding load")

	return pauseCommand
}

func (c *pauseCommand) run(cmd *cobra.Command, args []string) error {
	var response bool

	c.args.Service = args[0]

	return withRPCClient(globalConfig.SocketPath(), func(client *rpc.Client) error {
		return client.Call("kamal-proxy.Pause", c.args, &response)
	})
}
