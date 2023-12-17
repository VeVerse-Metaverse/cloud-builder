package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// expand replaces placeholders in the string s with their corresponding
// values from the match map.
func expand(match map[string]string, s string) string {
	var oldNew []string

	for key, value := range match {
		oldNew = append(oldNew, "{"+key+"}", value)
	}

	replacer := strings.NewReplacer(oldNew...)
	return replacer.Replace(s)
}

// prepareArguments replaces placeholders in the arguments with their corresponding values.
func prepareArguments(commandLine string, keyValueMap map[string]string) []string {
	args := strings.Fields(commandLine)
	for i, arg := range args {
		args[i] = expand(keyValueMap, arg)
	}
	return args
}

// executeCommand runs the given command with the specified arguments and working directory.
func executeCommand(ctx context.Context, command string, arguments []string, workingDir string) ([]byte, error) {
	cmd := exec.Command(command, arguments...)
	cmd.Dir = workingDir
	output, err := cmd.Output()

	if err != nil {
		err = fmt.Errorf("error executing command: %w\n", err)
		fmt.Println(err.Error())
	}

	return output, err
}

type Cmd struct {
	// The Command to run, e.g. "git".
	Command string

	// CommandLine arguments to pass to the command, e.g. "git checkout -b {branchName}".
	CommandLine string

	// The working directory to run the command in, e.g. "C:\myRepo".
	WorkingDir string

	// The key-value pairs to use for placeholder replacement, e.g. {"branchName": "myBranch"}.
	Placeholders map[string]string

	// The output of the command, e.g. "Switched to a new branch 'myBranch'".
	Output []byte

	// The error returned by the command, e.g. "fatal: A branch named 'myBranch' already exists.".
	Error error

	// The exit code of the command, e.g. 1.
	ExitCode int

	// Prepared arguments to pass to the command, e.g. ["checkout", "-b", "myBranch"].
	arguments []string
}

// Run runs the command.
func (c *Cmd) Run(ctx context.Context) error {
	c.arguments = prepareArguments(c.CommandLine, c.Placeholders)

	if _, err := os.Stat(c.WorkingDir); os.IsNotExist(err) {
		c.Error = fmt.Errorf("working directory does not exist: %s", c.WorkingDir)
		return c.Error
	}

	if _, err := exec.LookPath(c.Command); err != nil {
		c.Error = fmt.Errorf("command not found: %s", c.Command)
		return c.Error
	}

	c.Output, c.Error = executeCommand(ctx, c.Command, c.arguments, c.WorkingDir)

	if c.Error != nil {
		if exitError, ok := c.Error.(*exec.ExitError); ok {
			c.ExitCode = exitError.ExitCode()
		}
	}

	return c.Error
}
