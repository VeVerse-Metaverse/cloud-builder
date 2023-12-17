// Summary: Main entry point for the application.
// Description: This file is used to setup the application and run the root command.
// Author: Egor Pristavka
// Date: 2023-03-27

package main

import (
	"context"
	sl "dev.hackerman.me/artheon/veverse-shared/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"l7-cloud-builder/api"
	"l7-cloud-builder/config"
	"l7-cloud-builder/database"
	"l7-cloud-builder/logger"
	"l7-cloud-builder/processing"
	"os"
	"strings"
)

var rootCmd *cobra.Command
var ctx context.Context
var cancel context.CancelFunc

func init() {
	var err error

	// Create a application scoped context with cancel function.
	ctx, cancel = context.WithCancel(context.Background())

	//region Setup logging

	// Setup Clickhouse client (used for logging).
	ctx, err = database.SetupClickhouse(ctx)
	if err != nil {
		logger.Logger.Errorf("failed to setup clickhouse: %s\n", err.Error())
	}

	// Setup logrus hook.
	var hook logrus.Hook
	hook, err = sl.NewHook(ctx)
	if err != nil {
		logger.Logger.Errorf("failed to setup logrus hook: %s\n", err.Error())
	} else {
		logger.Logger.AddHook(hook)
	}

	//endregion

	//region Load API configuration

	// Load API URL.
	config.Api.Url = os.Getenv("API_URL")
	if config.Api.Url == "" {
		logger.Logger.Fatalln("required env VAT_API2_URL is not defined")
	}

	// Load API credentials file path.
	config.Api.CredentialsPath = os.Getenv("API_CREDENTIALS_PATH")
	if config.Api.CredentialsPath == "" {
		config.Api.CredentialsPath = ".credentials"
	}

	// Load API credentials.
	config.Api.Email = os.Getenv("API_EMAIL")
	config.Api.Password = os.Getenv("API_PASSWORD")
	if config.Api.Email == "" || config.Api.Password == "" {
		logger.Logger.Infof("loading credentials from file: %s\n", config.Api.CredentialsPath)
		b, err := os.ReadFile(config.Api.CredentialsPath)
		if err != nil {
			logger.Logger.Fatalln(err)
		}
		c := string(b)
		t := strings.Split(c, ":")
		config.Api.Email = t[0]
		config.Api.Password = t[1]
	}

	// Load shared configuration from the API.
	err = api.LoadSharedConfiguration(ctx)
	if err != nil {
		logger.Logger.Errorf("failed to load shared configuration: %s, continuing with default values\n", err.Error())
	}

	//endregion

	rootCmd = &cobra.Command{
		Use: "process",
		Run: func(cmd *cobra.Command, args []string) {
			//region Enabled jobs, targets and platforms

			// Load enabled job types.
			enabledJobs := os.Getenv("ENABLED_JOBS")
			if enabledJobs == "" {
				logger.Logger.Fatalln("required env ENABLED_JOBS is not defined")
			}
			for _, t := range strings.Split(enabledJobs, ",") {
				config.Config.EnabledJobs[t] = true
			}

			// Load enabled target types (e.g. Client, Server, Editor).
			enabledTargets := os.Getenv("ENABLED_TARGETS")
			if enabledTargets == "" {
				logger.Logger.Fatalln("required env ENABLED_TARGETS is not defined")
			}
			for _, d := range strings.Split(enabledTargets, ",") {
				config.Config.EnabledTargets[d] = true
			}

			// Load enabled platforms (e.g. Windows, Linux, Android, iOS).
			enabledPlatforms := os.Getenv("ENABLED_PLATFORMS")
			if enabledPlatforms == "" {
				logger.Logger.Fatalln("required env ENABLED_PLATFORMS is not defined")
			}
			for _, p := range strings.Split(enabledPlatforms, ",") {
				config.Config.EnabledPlatforms[p] = true
			}

			// Load Unreal Engine source code Unreal Automation Tool path. Required for UGC Package and Client/Server Release jobs.
			config.Unreal.Code.AutomationToolPath = os.Getenv("UNREAL_CODE_AUTOMATION_TOOL_PATH")
			if config.Unreal.Code.AutomationToolPath == "" {
				// Check if the job type is required.
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypePackage]] {
					// Code UAT path is required for the Package jobs.
					logger.Logger.Fatalln("required env UNREAL_CODE_AUTOMATION_TOOL_PATH is not defined")
				} else if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] {
					if config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeClient]] ||
						config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeServer]] {
						// Code UAT required for the Release jobs with non-Editor Targets.
						logger.Logger.Fatalln("required env UNREAL_CODE_AUTOMATION_TOOL_PATH is not defined")
					}
				}
			}

			//endregion

			//region Version Selector Tool

			// Load Unreal Engine source code Version Selector Tool path.
			config.Unreal.Code.VersionSelectorPath = os.Getenv("UNREAL_CODE_VERSION_SELECTOR_PATH")
			if config.Unreal.Code.VersionSelectorPath == "" {
				logger.Logger.Fatalln("required env UNREAL_CODE_VERSION_SELECTOR_PATH is not defined")
			}

			// Load Unreal Engine marketplace Version Selector Tool path.
			config.Unreal.Marketplace.VersionSelectorPath = os.Getenv("UNREAL_MARKETPLACE_VERSION_SELECTOR_PATH")
			if config.Unreal.Marketplace.VersionSelectorPath == "" {
				logger.Logger.Fatalln("required env UNREAL_MARKETPLACE_VERSION_SELECTOR_PATH is not defined")
			}

			// At least one of the Version Selector Tool paths must be defined, required for all jobs related to Unreal Engine.
			if config.Unreal.Code.VersionSelectorPath == "" && config.Unreal.Marketplace.VersionSelectorPath == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypePackage]] ||
					config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] {
					logger.Logger.Fatalln("required env UNREAL_CODE_VERSION_SELECTOR_PATH or UNREAL_MARKETPLACE_VERSION_SELECTOR_PATH is not defined")
				}
			}

			//endregion

			//region Unreal Engine Editor

			// Load Unreal Engine source code Editor path.
			config.Unreal.Code.EditorPath = os.Getenv("UNREAL_CODE_EDITOR_PATH")
			// Code editor path is required for the UGC Package and Client/Server Release jobs.
			if config.Unreal.Code.EditorPath == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypePackage]] ||
					(config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
						!config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeEditor]]) {
					logger.Logger.Fatalln("required env UNREAL_CODE_EDITOR_PATH is not defined")
				}
			}

			// Load Unreal Engine marketplace Editor path.
			config.Unreal.Marketplace.EditorPath = os.Getenv("UNREAL_MARKETPLACE_EDITOR_PATH")
			// Marketplace editor path is required for the Editor Release jobs.
			if config.Unreal.Marketplace.EditorPath == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeEditor]] {
					logger.Logger.Fatalln("required env UNREAL_MARKETPLACE_EDITOR_PATH is not defined")
				}
			}

			//endregion

			//region Unreal Engine Version

			// Load Unreal Engine marketplace version, required for Editor Release jobs.
			config.Unreal.Marketplace.Version = os.Getenv("UNREAL_MARKETPLACE_VERSION")
			if config.Unreal.Marketplace.Version == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeEditor]] {
					logger.Logger.Fatalln("required env UNREAL_MARKETPLACE_VERSION is not defined")
				}
			}

			//endregion

			//region Unreal Engine Automation Tool

			// Load Unreal Engine source code Unreal Automation Tool path, required for UGC Package and Client/Server Release jobs.
			config.Unreal.Code.AutomationToolPath = os.Getenv("UNREAL_CODE_AUTOMATION_TOOL_PATH")
			if config.Unreal.Code.AutomationToolPath == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypePackage]] ||
					(config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
						!config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeEditor]]) {
					logger.Logger.Fatalln("required env UNREAL_CODE_AUTOMATION_TOOL_PATH is not defined")
				}
			}

			// Load Unreal Engine marketplace Unreal Automation Tool path.
			config.Unreal.Marketplace.AutomationToolPath = os.Getenv("UNREAL_MARKETPLACE_AUTOMATION_TOOL_PATH")
			// Required for Editor Release jobs.
			if config.Unreal.Marketplace.AutomationToolPath == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeEditor]] {
					logger.Logger.Fatalln("required env UNREAL_MARKETPLACE_AUTOMATION_TOOL_PATH is not defined")
				}
			}

			//endregion

			//region Client Launcher (Wails)

			// Load Wails path.
			config.ClientLauncher.WailsPath = os.Getenv("LAUNCHER_WAILS_PATH")
			// Required for the ClientLauncher job.
			if config.ClientLauncher.WailsPath == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeLauncher]] {
					logger.Logger.Fatalln("required env LAUNCHER_WAILS_PATH is not defined")
				}
			}

			// Load launcher source code path.
			config.ClientLauncher.SourceDir = os.Getenv("LAUNCHER_SOURCE_DIR")
			// Required for the ClientLauncher job.
			if config.ClientLauncher.SourceDir == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeLauncher]] {
					logger.Logger.Fatalln("required env LAUNCHER_SOURCE_DIR is not defined")
				}
			}

			//endregion

			//region Server Launcher

			// Load server launcher source code path.
			config.ServerLauncher.SourceDir = os.Getenv("SERVER_LAUNCHER_SOURCE_DIR")
			// Required for the ServerLauncher job.
			if config.ServerLauncher.SourceDir == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeServerLauncher]] {
					logger.Logger.Fatalln("required env SERVER_LAUNCHER_SOURCE_DIR is not defined")
				}
			}

			//endregion

			//region Pixel Streaming Launcher

			// Load pixel streaming launcher source code path.
			config.PixelStreamingLauncher.SourceDir = os.Getenv("PIXEL_STREAMING_LAUNCHER_SOURCE_DIR")
			// Required for the PixelStreamingLauncher job.
			if config.PixelStreamingLauncher.SourceDir == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypePixelStreamingLauncher]] {
					logger.Logger.Fatalln("required env PIXEL_STREAMING_LAUNCHER_SOURCE_DIR is not defined")
				}
			}

			//endregion

			//region Code Signing

			// Load code signing tool path.
			config.CodeSigning.ToolPath = os.Getenv("CODE_SIGNING_TOOL_PATH")
			// Load code signing tool certificate path.
			config.CodeSigning.CertificatePath = os.Getenv("CODE_SIGNING_CERTIFICATE_PATH")
			// Load code signing tool certificate password.
			config.CodeSigning.CertificatePassword = os.Getenv("CODE_SIGNING_CERTIFICATE_PASSWORD")
			// Optional for the Client and Launcher Release job on Win64 platform.
			if config.CodeSigning.ToolPath == "" || config.CodeSigning.CertificatePath == "" || config.CodeSigning.CertificatePassword == "" {
				if config.Config.EnabledJobs[config.Config.JobMapping[config.JobTypeRelease]] &&
					config.Config.EnabledPlatforms[config.Config.PlatformMapping[config.PlatformTypeWindows]] &&
					(config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeClient]] ||
						config.Config.EnabledTargets[config.Config.TargetMapping[config.TargetTypeLauncher]]) {
					logger.Logger.Warningf("optional envs CODE_SIGNING_TOOL_PATH, CODE_SIGNING_CERTIFICATE_PATH, CODE_SIGNING_CERTIFICATE_PASSWORD are not defined, code signing will be skipped")
				}
			}

			//endregion

			for {
				if err := processing.Process(ctx); err != nil {
					logger.Logger.Errorf("failed to process a job: %v", err)
				}
			}
		},
	}
}

func main() {
	// Parse command line arguments.
	if err := rootCmd.Execute(); err != nil {
		logger.Logger.Fatalln(err)
	}
}
