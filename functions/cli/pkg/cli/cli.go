package cli

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"syscall"
)

type CommandCategory struct {
	Name        string
	Description string
}

type CLI struct {
	commands map[string]Command
	debug    bool
}

type Command interface {
	Execute(args []string) error
	Help() CommandHelp
}

type CommandHelp struct {
	Description string
	Usage       string
	Examples    []string
	Category    CommandCategory
}

// Basic command that shells out to a binary
type ShellCommand struct {
	binary      string
	args        []string
	description string
	usage       string
	examples    []string
	category    CommandCategory
}

func (c *ShellCommand) Execute(args []string) error {
	cmd := exec.Command(c.binary, append(c.args, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *ShellCommand) Help() CommandHelp {
	return CommandHelp{
		Description: c.description,
		Usage:       c.usage,
		Examples:    c.examples,
		Category:    c.category,
	}
}

// Specific commands
type StartCommand struct {
	ShellCommand
}

// Define categories
var (
	managementCategory = CommandCategory{
		Name:        "Management Commands",
		Description: "Commands to manage Uberbase services",
	}
	commonCategory = CommandCategory{
		Name:        "Common Commands",
		Description: "Most commonly used commands",
	}
)

func NewStartCommand() *StartCommand {
	return &StartCommand{
		ShellCommand{
			binary:      "./bin/start",
			description: "Start Uberbase services",
			usage:       "start",
			examples:    []string{"start"},
			category:    commonCategory,
		},
	}
}

type StopCommand struct {
	ShellCommand
}

func NewStopCommand() *StopCommand {
	return &StopCommand{
		ShellCommand{
			binary:      "./bin/stop",
			description: "Stop all services",
			usage:       "stop",
			examples:    []string{"stop"},
			category:    commonCategory,
		},
	}
}

type DeployCommand struct {
	ShellCommand
}

func NewDeployCommand() *DeployCommand {
	return &DeployCommand{
		ShellCommand{
				binary:      "./bin/kamal",
				description: "Deploy the application",
				usage:       "deploy [environment] [options]",
				examples:    []string{
					"deploy production",
					"deploy staging --version=v1",
				},
				category: managementCategory,
		},
	}
}

type ComposeCommand struct {
	ShellCommand
}

func NewComposeCommand() *ComposeCommand {
	return &ComposeCommand{
		ShellCommand{
			binary:      "podman-compose",
			description: "Multi-container management",
			usage:       "compose [command] [options]",
			examples:    []string{
				"compose up -d",
				"compose down",
				"compose logs -f",
			},
			category: managementCategory,
		},
	}
}

func NewCLI() *CLI {
	cli := &CLI{
		commands: make(map[string]Command),
		debug:    false,
	}

	// Register commands
	cli.commands["start"] = NewStartCommand()
	cli.commands["stop"] = NewStopCommand()
	cli.commands["deploy"] = NewDeployCommand()
	cli.commands["compose"] = NewComposeCommand()

	return cli
}

func (cli *CLI) ShowHelp() {
	fmt.Printf("\nUsage: %s [OPTIONS] COMMAND\n\n", os.Args[0])
	fmt.Printf("A full-featured platform-in-a-box.\n\n")

	// Group commands by category
	categories := make(map[string][]string)
	for name, cmd := range cli.commands {
		help := cmd.Help()
		categories[help.Category.Name] = append(categories[help.Category.Name], name)
	}

	// Sort category names
	var categoryNames []string
	for name := range categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	// Print each category and its commands
	for _, categoryName := range categoryNames {
		fmt.Printf("%s:\n", categoryName)
		
		// Sort commands within category
		commands := categories[categoryName]
		sort.Strings(commands)
		
		// Print each command in the category
		for _, name := range commands {
			cmd := cli.commands[name]
			help := cmd.Help()
			fmt.Printf("  %-15s%s\n", name, help.Description)
		}
		fmt.Println()
	}

	// Add section for commands that pass through to podman
	fmt.Printf("Commands:\n")
	fmt.Printf("  All other commands are passed directly to Podman (e.g., ps, images, build)\n\n")

	fmt.Printf("Run '%s COMMAND --help' for more information on a command.\n\n", os.Args[0])
}

func (cli *CLI) showDebugInfo() {
	fmt.Println("\n=== Debug Information ===")
	
	// Show user info
	uid := os.Getuid()
	gid := os.Getgid()
	fmt.Printf("Running as UID: %d, GID: %d\n", uid, gid)

	// Find podman path
	podmanPath, err := exec.LookPath("podman")
	if err != nil {
		fmt.Printf("Podman not found in PATH: %v\n", err)
	} else {
		fmt.Printf("Podman path: %s\n", podmanPath)
	}

	// Find podman-compose path
	composePath, err := exec.LookPath("podman-compose")
	if err != nil {
		fmt.Printf("Podman-compose not found in PATH: %v\n", err)
	} else {
		fmt.Printf("Podman-compose path: %s\n", composePath)
	}
	
	fmt.Println("=====================")
}

func (cli *CLI) Execute() error {
	if len(os.Args) < 2 {
		cli.ShowHelp()
		return fmt.Errorf("no command specified")
	}

	// Check for debug flag anywhere in the arguments
	for _, arg := range os.Args {
		if arg == "--debug" {
			cli.debug = true
			break
		}
	}

	// Show debug info if flag is set
	if cli.debug {
		cli.showDebugInfo()
	}

	// Check for help flags
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		cli.ShowHelp()
		return nil
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Filter out --debug from args if it exists
	if cli.debug {
		filteredArgs := make([]string, 0, len(args))
		for _, arg := range args {
			if arg != "--debug" {
				filteredArgs = append(filteredArgs, arg)
			}
		}
		args = filteredArgs
	}

	// Show help for specific command
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		if cmd, ok := cli.commands[command]; ok {
			help := cmd.Help()
			fmt.Printf("\n%s - %s\n", command, help.Description)
			fmt.Printf("\nUsage: %s %s\n", os.Args[0], help.Usage)
			if len(help.Examples) > 0 {
				fmt.Println("\nExamples:")
				for _, example := range help.Examples {
					fmt.Printf("  %s %s\n", os.Args[0], example)
				}
			}
			return nil
		}
	}

	// First check if it's a registered command
	if cmd, ok := cli.commands[command]; ok {
		return cmd.Execute(args)
	}

	// If not found in our registered commands, 
	// pass through to podman directly
	podmanCmd := exec.Command("podman", append([]string{command}, args...)...)
	podmanCmd.Stdout = os.Stdout
	podmanCmd.Stderr = os.Stderr
	podmanCmd.Stdin = os.Stdin

	if err := podmanCmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Preserve the exit code from podman
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return err
	}

	return nil
}
