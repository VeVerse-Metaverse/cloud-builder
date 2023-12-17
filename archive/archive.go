package archive

import (
	"context"
	sm "dev.hackerman.me/artheon/veverse-shared/model"
	"fmt"
	"github.com/mholt/archiver/v4"
	"l7-cloud-builder/logger"
	"os"
	"path/filepath"
)

func createReleaseArchive(job *sm.JobV2, projectStagingDir string) (string, error) {
	ignore, err := getReleaseIgnoredFiles()
	if err != nil {
		return "", err
	}

	platformStagingDir := filepath.Join(projectStagingDir, getPlatformName(job))
	files, err := listFilesRecursive(platformStagingDir, ignore)

	releaseArchiveFileMap := createReleaseArchiveFileMap(platformStagingDir, files)
	if err := uploadReleaseFiles(job, releaseArchiveFileMap); err != nil {
		return "", err
	}

	zipName := fmt.Sprintf("%s-%s-%s-%s-%s.zip", job.Release.AppName, job.Release.Version, job.Deployment, job.Configuration, job.Platform)
	zip, err := os.Create(zipName)
	if err != nil {
		return "", fmt.Errorf("failed to create a zip file: %v", err)
	}
	defer closeZipFile(zip)

	format := archiver.CompressedArchive{
		Archival: archiver.Zip{},
	}

	releaseArchiveFiles, err := archiver.FilesFromDisk(nil, releaseArchiveFileMap)
	if err != nil {
		return "", fmt.Errorf("failed to enumerate release archive files to zip: %v", err)
	}

	err = format.Archive(context.Background(), zip, releaseArchiveFiles)
	if err != nil {
		return "", fmt.Errorf("failed to zip release archive files: %v", err)
	}

	return zipName, nil
}

func createReleaseArchiveFileMap(platformStagingDir string, files []string) map[string]string {
	releaseArchiveFileMap := make(map[string]string)
	for _, file := range files {
		path := filepath.Join(platformStagingDir, file)
		originalPath := filepath.ToSlash(file)
		releaseArchiveFileMap[path] = originalPath
	}
	return releaseArchiveFileMap
}

func uploadReleaseFiles(job *Job, releaseArchiveFileMap map[string]string) error {
	for path, originalPath := range releaseArchiveFileMap {
		err := uploadReleaseFile(job, path, originalPath, nil)
		if err != nil {
			return fmt.Errorf("failed to upload release file: %v", err)
		}
	}
	return nil
}

func closeZipFile(zip *os.File) {
	err := zip.Close()
	if err != nil {
		logger.Logger.Errorf("failed to close a zip file: %v", err)
	}
}
