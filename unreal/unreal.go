package unreal

import (
	"context"
	"fmt"
	"l7-cloud-builder/cmd"
	"l7-cloud-builder/config"
	"path/filepath"
)

func GetStagingDir(projectDir string) string {
	return filepath.Join(projectDir, "Saved", "StagedBuilds")
}

func RunAutomationTool(ctx context.Context, workdir string, command string, cmdline string, placeholders map[string]string) error {
	var uat = &cmd.Cmd{
		Command:      command,
		CommandLine:  cmdline,
		WorkingDir:   workdir,
		Placeholders: placeholders,
	}

	if err := uat.Run(ctx); err != nil {
		return fmt.Errorf("failed to run Unreal Automation Tool: %w", err)
	}

	return nil
}

// SwitchProjectEngineVersion switches the engine version for the project
// workdir: the project directory
// version: the engine version to switch to (marketplace version or source version directory)
func SwitchProjectEngineVersion(ctx context.Context, workdir string, project string, version string) error {
	// Get path to the project descriptor file
	projectDescriptorPath := filepath.Join(workdir, project+".uproject")

	// Select existing engine version selector tool from code or marketplace config
	var versionSelectorPath string
	if config.Unreal.Code.VersionSelectorPath != "" {
		versionSelectorPath = config.Unreal.Code.VersionSelectorPath
	} else if config.Unreal.Marketplace.VersionSelectorPath != "" {
		versionSelectorPath = config.Unreal.Marketplace.VersionSelectorPath
	} else {
		return fmt.Errorf("failed to find engine version selector tool")
	}

	// Switch engine version
	var uvs = &cmd.Cmd{
		Command:      versionSelectorPath,
		CommandLine:  "-switchversionsilent {project} {version}",
		WorkingDir:   workdir,
		Placeholders: map[string]string{"project": projectDescriptorPath, "version": version},
	}

	if err := uvs.Run(ctx); err != nil {
		return fmt.Errorf("failed to switch engine version for the project %s: %w", projectDescriptorPath, err)
	}

	return nil
}
