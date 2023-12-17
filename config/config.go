// Summary: Configuration for the automation tool.
// Description: This file is used to configure the automation tool. It contains configuration for the API, launcher, Unreal Engine, code signing, supported targets, jobs, platforms and statuses.
// Author: Egor Pristavka
// Date: 2023-03-27

package config

// JobType is a type for job
type JobType int

const (
	// JobTypeRelease is a job type for release jobs (build the Unreal Engine project for a specific platform, configuration and target)
	JobTypeRelease JobType = iota
	// JobTypePackage is a job type for package jobs (build the Unreal Engine UGC package for a specific platform, configuration and target)
	JobTypePackage
)

// TargetType is a type for target
type TargetType int

const (
	// TargetTypeClient is a target type for client targets (build the Unreal Engine project or the ClientLauncher for a client)
	TargetTypeClient TargetType = iota
	// TargetTypeServer is a target type for server targets (build the Unreal Engine project or the ServerLauncher for a server)
	TargetTypeServer
	// TargetTypeEditor is a target type for SDK targets (build the Unreal Engine project for an SDK)
	TargetTypeEditor
	// TargetTypeLauncher is a target type for ClientLauncher targets (build the ClientLauncher)
	TargetTypeLauncher
	// TargetTypeServerLauncher is a target type for ServerLauncher targets (build the ServerLauncher)
	TargetTypeServerLauncher
	// TargetTypePixelStreamingLauncher is a target type for Pixel Streaming ClientLauncher targets (build the Pixel Streaming ClientLauncher)
	TargetTypePixelStreamingLauncher
)

// PlatformType is a type for platform
type PlatformType int

const (
	// PlatformTypeWindows is a platform type for Win64 platform
	PlatformTypeWindows PlatformType = iota
	// PlatformTypeLinux is a platform type for Linux platform
	PlatformTypeLinux
	// PlatformTypeMac is a platform type for Mac platform
	PlatformTypeMac
	// PlatformTypeAndroid is a platform type for Android platform
	PlatformTypeAndroid
	// PlatformTypeIOS is a platform type for iOS platform
	PlatformTypeIOS
)

// JobStatusType is a type for job status
type JobStatusType int

const (
	// JobStatusUnclaimed is a job status for unclaimed jobs (not yet claimed by any automation tool instance)
	JobStatusUnclaimed = iota
	// JobStatusClaimed is a job status for claimed jobs (claimed by the automation tool running on a node)
	JobStatusClaimed
	// JobStatusProcessing is a job status for processing jobs (processing by the automation tool)
	JobStatusProcessing
	// JobStatusUploading is a job status for uploading jobs (uploading to the cloud storage)
	JobStatusUploading
	// JobStatusCompleted is a job status for completed jobs (completed by the automation tool)
	JobStatusCompleted
	// JobStatusError is a job status for errored jobs (errored by the automation tool)
	JobStatusError
	// JobStatusCancelled is a job status for cancelled jobs (cancelled by the owner or admin)
	JobStatusCancelled
)

// ApiConfig is a struct for API configuration
type ApiConfig struct {
	Url             string // URL to the API (supplied via environment variable)
	Email           string // Email for the API (supplied via environment variable)
	Password        string // Password for the API (supplied via environment variable)
	Token           string // Token for the API, issued by the API after successful login, used for all subsequent requests
	CredentialsPath string // Path to the credentials file (used to store credentials if they are not provided via environment variables)
}

// ClientLauncherConfig is a struct for client launcher configuration
type ClientLauncherConfig struct {
	WailsPath string // Path to Wails CLI
	SourceDir string // Path to the client launcher source code
}

// ServerLauncherConfig is a struct for server launcher configuration
type ServerLauncherConfig struct {
	SourceDir string // Path to the server launcher source code
}

// PixelStreamingLauncherConfig is a struct for Pixel Streaming launcher configuration
type PixelStreamingLauncherConfig struct {
	SourceDir string // Path to the Pixel Streaming launcher source code
}

// UnrealEngineVersionConfig is a struct for Unreal Engine version configuration
type UnrealEngineVersionConfig struct {
	AutomationToolPath  string // Path to Unreal Automation Tool executable
	VersionSelectorPath string // Path to Unreal Version Selector executable
	EditorPath          string // Path to UnrealEditor-Cmd executable
	Version             string // Unreal Engine version, used to select the correct version of the Unreal Engine with UVS, used only for Marketplace version
}

// UnrealConfig is a struct for Unreal Engine configuration
type UnrealConfig struct {
	Project struct {
		Directory string // Path to the project
		Name      string // Project name
	}
	Code        UnrealEngineVersionConfig // Code version Unreal Engine version configuration
	Marketplace UnrealEngineVersionConfig // Marketplace version Unreal Engine version configuration
}

// CodeSigningConfig is a struct for code signing configuration
type CodeSigningConfig struct {
	ToolPath            string // Path to SignTool
	CertificatePath     string // Path to certificate file
	CertificatePassword string // Password for the certificate
}

// AutomationConfig is a struct for the automation tool configuration
type AutomationConfig struct {
	EnabledTargets   map[string]bool          // Targets enabled for this automation tool (Client, Server, SDK)
	EnabledJobs      map[string]bool          // Job types enabled for this automation tool (ClientLauncher, Release, Package)
	EnabledPlatforms map[string]bool          // Platforms enabled for this automation tool (Windows, Linux, Mac, Android, iOS)
	StatusMapping    map[JobStatusType]string // Job statuses mapped to strings
	JobMapping       map[JobType]string       // Job types mapped to strings
	TargetMapping    map[TargetType]string    // Target types mapped to strings
	PlatformMapping  map[PlatformType]string  // Platform types mapped to strings
}

// GitConfig is a struct for Git configuration
type GitConfig struct {
	BranchMapping map[string]string // Mapping of branch names to target names (what branch to select for a target)
}

// SharedConfig is a struct for shared configuration for the automation tool, stored in the database and allows to change the configuration without restarting the automation tool
type SharedConfig struct {
	Release struct {
		IgnoredFiles []string // List of files to ignore when packing a release into a zip archive
	}
}

var (
	// Api contains configuration for the API
	Api = ApiConfig{
		CredentialsPath: ".credentials",
	}
	// ClientLauncher contains configuration for the launcher
	ClientLauncher = ClientLauncherConfig{}
	// ServerLauncher contains configuration for the launcher
	ServerLauncher = ServerLauncherConfig{}
	// PixelStreamingLauncher contains configuration for the launcher
	PixelStreamingLauncher = PixelStreamingLauncherConfig{}
	// Unreal contains configuration for Unreal Engine
	Unreal = UnrealConfig{}
	// CodeSigning contains configuration for code signing
	CodeSigning = CodeSigningConfig{}
	// Config contains configuration for supported targets, jobs, platforms and statuses
	Config = AutomationConfig{
		EnabledTargets:   map[string]bool{},
		EnabledJobs:      map[string]bool{},
		EnabledPlatforms: map[string]bool{},
		StatusMapping: map[JobStatusType]string{
			JobStatusUnclaimed:  "unclaimed",
			JobStatusClaimed:    "claimed",
			JobStatusProcessing: "processing",
			JobStatusUploading:  "uploading",
			JobStatusCompleted:  "completed",
			JobStatusError:      "error",
			JobStatusCancelled:  "cancelled",
		},
		JobMapping: map[JobType]string{
			JobTypeRelease: "release",
			JobTypePackage: "package",
		},
		TargetMapping: map[TargetType]string{
			TargetTypeClient:                 "client",
			TargetTypeServer:                 "server",
			TargetTypeEditor:                 "editor",
			TargetTypeLauncher:               "launcher",
			TargetTypeServerLauncher:         "server-launcher",
			TargetTypePixelStreamingLauncher: "pixel-streaming-launcher",
		},
	}
	// Git contains configuration for Git
	Git = GitConfig{
		BranchMapping: map[string]string{
			"Debug":       "development",
			"DebugGame":   "development",
			"Development": "development",
			"Test":        "development",
			"Shipping":    "development",
		},
	}
	// Shared contains shared configuration for the automation tool, stored in the database and allows to change the configuration without restarting the automation tool
	Shared = SharedConfig{
		Release: struct {
			IgnoredFiles []string
		}{
			IgnoredFiles: []string{
				".git",
				".gitignore",
				".gitattributes",
				".gitmodules",
				".gitkeep",
				".gitlab-ci.yml",
				".gitlab-ci",
				"/.git",
				".DS_Store",
				".idea",
				".vscode",
				"*.sln",
				"*.suo",
				"*.user",
				"/.idea",
				"Manifest_DebugFiles_Linux.xml",
				"Manifest_NonUFSFiles_Linux.xml",
				"Manifest_UFSFiles.xml",
				"Manifest_DebugFiles_Win64.xml",
				"Manifest_NonUFSFiles_Win64.xml",
				"Manifest_UFSFiles_Win64.xml",
				"MetaverseServer.debug", // Decreases download time and traffic for server
				"MetaverseServer.sym",   // Decreases download time and traffic for server
				"MetaverseServer.pdb",   // Decreases download time and traffic for server
				"Engine/Extras/GPUDumpViewer/OpenGPUDumpViewer.bat",
				"Engine/Extras/GPUDumpViewer/OpenGPUDumpViewer.sh",
				"Engine/Extras/GPUDumpViewer/GPUDumpViewer.html",
				"Samples/PixelStreaming", // Remove Pixel Streaming samples from the release
			},
		},
	}
)
