package processing

import (
	"context"
	sh "dev.hackerman.me/artheon/veverse-shared/helper"
	sm "dev.hackerman.me/artheon/veverse-shared/model"
	"fmt"
	"l7-cloud-builder/api"
	"l7-cloud-builder/archive"
	"l7-cloud-builder/config"
	"l7-cloud-builder/git"
	"l7-cloud-builder/unreal"
	"l7-cloud-builder/upload"
	"path/filepath"
)

func generateReleaseClientCmdline(ctx context.Context, job *sm.JobV2) (string, map[string]string, error) {
	// Validate job
	if job == nil {
		return "", nil, fmt.Errorf("job is nil")
	}

	// Validate job release
	if job.Release == nil {
		return "", nil, fmt.Errorf("job release is nil")
	}

	// Validate job release id
	if job.Release.Id.IsNil() {
		return "", nil, fmt.Errorf("invalid job release")
	}

	stagingDirectory := filepath.Join(config.Unreal.Project.Directory, "Saved", "StagedBuilds", job.Release.Version)

	// Generate the command line arguments
	cmdline := "BuildCookRun -project={project} -noP4 -unrealexe={unrealexe} -clientconfig={configuration} -platform={platform} -ini:Game:[/Script/UnrealEd.ProjectPackagingSettings]:BlueprintNativizationMethod=Disabled -build -cook -unversionedcookedcontent -SkipCookingEditorContent -map={maps} -pak -compressed -package -createreleaseversion={releaseVersion} -stage -stagingdirectory={stagingDirectory} -VeryVerbose -NoCodeSign -BuildMachine -AllowCommandletRendering -utf8output -debuginfo -debug"

	// Placeholders
	placeholders := map[string]string{
		"project":          config.Unreal.Project.Name + ".uproject",
		"unrealexe":        config.Unreal.Code.EditorPath,
		"configuration":    job.Configuration,
		"platform":         job.Platform,
		"maps":             job.Release.Options.Maps, // TODO: implement maps
		"releaseVersion":   job.Release.Version,
		"stagingDirectory": stagingDirectory,
	}

	// Add additional command line arguments to the shipping configuration (add prerequisites, build for distribution)
	if job.Configuration == "Shipping" {
		cmdline += " -CrashReporter -distribution -prereqs" // -nodebug -nodebuginfo
	}

	return cmdline, placeholders, nil
}

func processReleaseClient(ctx context.Context, job *sm.JobV2) (err error) {
	if job == nil {
		return fmt.Errorf("job is nil")
	}

	// Mark the job as processing
	if err = api.UpdateJobStatus(ctx, job, config.JobStatusProcessing, ""); err != nil {
		return
	}

	//region Validate the received job

	// Validate job type
	if !config.Config.EnabledJobs[job.Type] {
		err = fmt.Errorf("invalid job type: %s", job.Type)
		return
	}

	// Validate job target
	if !config.Config.EnabledTargets[job.Target] {
		err = fmt.Errorf("invalid job target: %s", job.Target)
		return err
	}

	// Validate job platform
	if !config.Config.EnabledPlatforms[job.Platform] {
		err = fmt.Errorf("invalid job platform: %s", job.Platform)
		return err
	}

	// Validate job release
	if job.Release == nil {
		err = fmt.Errorf("job release is nil")
		return err
	}

	// Validate job release id
	if job.Release.Id.IsNil() {
		err = fmt.Errorf("invalid job release")
		return err
	}

	//endregion

	// Update the repo
	if err = git.Fetch(ctx, config.Unreal.Project.Directory); err != nil {
		return fmt.Errorf("failed to update the repo: %w", err)
	}

	// Checkout the tag matching the release code version
	if err = git.CheckoutTag(ctx, config.Unreal.Project.Directory, job.Release.CodeVersion); err != nil {
		return fmt.Errorf("failed to checkout tag %s: %w", job.Release.CodeVersion, err)
	}

	// Switch the project engine version to code version
	if err = unreal.SwitchProjectEngineVersion(ctx, config.Unreal.Project.Directory, config.Unreal.Project.Name, job.Release.CodeVersion); err != nil {
		return fmt.Errorf("failed to switch project engine version: %w", err)
	}

	// Generate the command line arguments
	cmdline, placeholders, err := generateReleaseClientCmdline(ctx, job)

	// Run the source code engine version Unreal Automation Tool to build the client
	if err = unreal.RunAutomationTool(ctx, config.Unreal.Project.Directory, config.Unreal.Code.AutomationToolPath, cmdline, placeholders); err != nil {
		return err
	}

	// Get list of ignored files from the config
	ignoredFiles := config.Shared.Release.IgnoredFiles
	stagingDirectory := filepath.Join(unreal.GetStagingDir(config.Unreal.Project.Directory), job.Release.Version)

	// Get list of files in the staging directory
	files, err := sh.ListFilesRecursive(stagingDirectory, ignoredFiles)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Check if release is a archive
	if job.Release.Options.Archive {
		zipFileName := fmt.Sprintf("%s-%s-%s-%s-%s.zip", job.Release.App.Id.String(), job.Release.Version, job.Target, job.Platform, job.Configuration) // e.g.

		// Create the archive
		err = archive.CreateZipArchive(zipFileName, stagingDirectory, files)
		if err != nil {
			return fmt.Errorf("failed to create a release archive: %w", err)
		}

		// Upload the archive
		err = upload.ReleaseArchive(ctx, job, zipFileName)
	} else {
		// Upload the files one by one
	}

	return nil
}
