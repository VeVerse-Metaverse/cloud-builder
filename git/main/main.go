package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
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

// createKeyValueMap creates a map from the given key-value pairs.
func createKeyValueMap(keyValuePairs []string) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(keyValuePairs); i += 2 {
		m[keyValuePairs[i]] = keyValuePairs[i+1]
	}
	return m
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
func executeCommand(command string, arguments []string, workingDir string) ([]byte, error) {
	cmd := exec.Command(command, arguments...)
	cmd.Dir = workingDir
	output, err := cmd.Output()

	if err != nil {
		logrus.Errorf("error running git: %s", err.Error())
	}

	return output, err
}

// createNewDir creates a new directory and returns the updated directory path.
func createNewDir(newDir, currentDir string) string {
	var err error
	var fullPath string

	if filepath.IsAbs(newDir) {
		fullPath = newDir
	} else {
		fullPath = filepath.Join(currentDir, newDir)
	}

	err = os.Mkdir(fullPath, fs.ModePerm)

	if err != nil {
		logrus.Printf("Error creating directory: %v\n", err)
	}

	return fullPath
}

// processInternalFlags processes internal flags and updates the arguments and directory accordingly.
func processInternalFlags(arguments []string, dir string) ([]string, string) {
	for len(arguments) >= 2 {
		switch arguments[0] {
		case "-go-internal-mkdir":
			dir = createNewDir(arguments[1], dir)
			arguments = arguments[2:]

		case "-go-internal-cd":
			dir = changeDir(arguments[1], dir)
			arguments = arguments[2:]

		default:
			return arguments, dir
		}
	}
	return arguments, dir
}

// changeDir changes the working directory and returns the updated directory path.
func changeDir(newDir, currentDir string) string {
	if filepath.IsAbs(newDir) {
		return newDir
	} else {
		return filepath.Join(currentDir, newDir)
	}
}

func main() {
	// Parse command line arguments.
	keyValuePairs := []string{"goos", "linux", "goarch", "amd64"}
	commandLine := "go build -o {goos}-{goarch} main.go"
	dir := "."

	keyValueMap := createKeyValueMap(keyValuePairs)
	arguments := prepareArguments(commandLine, keyValueMap)

	// Process internal flags.
	arguments, dir = processInternalFlags(arguments, dir)

	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintf(os.Stderr,
			"go: missing %s command. See https://golang.org/s/gogetcmd\n",
			"git")
		os.Exit(2)
	}

	b, err := executeCommand("git", arguments, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go: %v\n", err)
		os.Exit(2)
	}

	fmt.Fprintf(os.Stdout, "%s", b)
}
