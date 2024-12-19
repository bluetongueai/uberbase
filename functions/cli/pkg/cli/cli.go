package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
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
	Help() (CommandHelp, bool)
}

type CommandHelp struct {
	Description string
	Usage       string
	Category    CommandCategory
}

type StartCommand struct {
	binary string
}

func NewStartCommand() *StartCommand {
	return &StartCommand{
		binary: "./bin/start",
	}
}

func (c *StartCommand) Execute(args []string) error {
	cmd := exec.Command(c.binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return err
	}
	return nil
}

func (c *StartCommand) Help() (CommandHelp, bool) {
	return CommandHelp{
		Description: "Start Uberbase services",
		Usage:       "start",
		Category:    commonCategory,
	}, true
}

type StopCommand struct {
	binary string
}

func NewStopCommand() *StopCommand {
	return &StopCommand{
		binary: "./bin/stop",
	}
}

func (c *StopCommand) Execute(args []string) error {
	cmd := exec.Command(c.binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return err
	}
	return nil
}

func (c *StopCommand) Help() (CommandHelp, bool) {
	return CommandHelp{
		Description: "Stop all services",
		Usage:       "stop",
		Category:    commonCategory,
	}, true
}

type ComposeCommand struct {
	binary string
}

func NewComposeCommand() *ComposeCommand {
	return &ComposeCommand{
		binary: "podman-compose",
	}
}

func (c *ComposeCommand) Execute(args []string) error {
	cmd := exec.Command(c.binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return err
	}
	return nil
}

type DeployCommand struct {
	binary string
}

func NewDeployCommand() *DeployCommand {
	return &DeployCommand{
		binary: "./bin/kamal",
	}
}

func captureAndReplaceOutput(cmd *exec.Cmd, replacements map[string]string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	output := stdout.String()
	if output == "" {
		output = stderr.String()  // Some commands write help to stderr
	}
	
	for old, new := range replacements {
		output = strings.ReplaceAll(output, old, new)
	}
	
	return output, err
}

func (c *DeployCommand) Execute(args []string) error {
	cmd := exec.Command(c.binary, args...)
	
	// If this is a help request
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		replacements := map[string]string{
			"Kamal": "deploy",
			"kamal": "deploy",
		}
		output, _ := captureAndReplaceOutput(cmd, replacements)  // Ignore error for help output
		fmt.Print(output)
		return nil
	}
	
	// Normal execution
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *DeployCommand) Help() (CommandHelp, bool) {
	return CommandHelp{}, false
}

func (c *ComposeCommand) Help() (CommandHelp, bool) {
	return CommandHelp{}, false
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
		if help, hasHelp := cmd.Help(); hasHelp {
			categories[help.Category.Name] = append(categories[help.Category.Name], name)
		}
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
			if help, hasHelp := cmd.Help(); hasHelp {
				fmt.Printf("  %-15s%s\n", name, help.Description)
			}
		}
		fmt.Println()
	}

	// Add section for commands that pass through
	fmt.Printf("Management Commands:\n")
	fmt.Printf("  %-15s%s\n", "deploy", "Deploy using Kamal (see: deploy --help)")
	fmt.Printf("  %-15s%s\n", "compose", "Manage containers using Podman Compose (see: compose --help)\n")
	fmt.Printf("\nCommands:\n")
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

	// Only show help if it's the global help flag
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
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

	// First check if it's a registered command
	if cmd, ok := cli.commands[command]; ok {
		// Check if user wants help
		wantsHelp := len(args) > 0 && (args[0] == "-h" || args[0] == "--help") || len(args) == 0 && (command == "-h" || command == "--help")
		if wantsHelp {
			// Check if command has custom help
			if help, hasHelp := cmd.Help(); hasHelp {
				fmt.Printf("\n%s - %s\n", command, help.Description)
				fmt.Printf("\nUsage: %s %s\n", os.Args[0], help.Usage)
				return nil
			}
			// No custom help, fall through to execute with help flag
		}
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
