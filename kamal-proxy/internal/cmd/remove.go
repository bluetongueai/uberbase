package cmd

import (
	"net/rpc"

	"github.com/spf13/cobra"

	"github.com/basecamp/kamal-proxy/internal/server"
)

type removeCommand struct {
	cmd  *cobra.Command
	args server.RemoveArgs
}

func newRemoveCommand() *removeCommand {
	removeCommand := &removeCommand{}
	removeCommand.cmd = &cobra.Command{
		Use:       "remove <service>",
		Short:     "Remove the service",
		RunE:      removeCommand.run,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"service"},
		Aliases:   []string{"rm"},
	}

	return removeCommand
}

func (c *removeCommand) run(cmd *cobra.Command, args []string) error {
	var response bool

	c.args.Service = args[0]

	return withRPCClient(globalConfig.SocketPath(), func(client *rpc.Client) error {
		return client.Call("kamal-proxy.Remove", c.args, &response)
	})
}
