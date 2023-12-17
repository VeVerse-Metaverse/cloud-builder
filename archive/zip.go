package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"l7-cloud-builder/logger"
	"os"
	"path/filepath"
)

// addToZip takes a zip.Writer, a basePath, and a path of a file.
// It adds the file to the zip archive using the zip.Writer, preserving the
// relative path of the file with respect to basePath.
// zipWriter: *zip.Writer - The zip writer used to add files to the archive.
// basePath: string - The base path to calculate the relative path of the file.
// path: string - The path of the file to be added to the zip archive.
func addToZip(zipWriter *zip.Writer, basePath, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close file: %v", err)
		}
	}(file)

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return err
	}

	header.Name = filepath.ToSlash(relPath)
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// CreateZipArchive takes an output path, a basePath, and a list of files.
// It creates a new zip archive at the specified output path and adds the files
// from the files list to the archive.
// output: string - The output path for the zip archive.
// basePath: string - The path to the directory containing the files.
// files: []string - The list of file paths relative to the basePath to be added to the archive.
func CreateZipArchive(output, basePath string, files []string) error {
	zipFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close zip file: %v", err)
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		err := zipWriter.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close zip writer: %v", err)
		}
	}(zipWriter)

	for _, file := range files {
		filePath := filepath.Join(basePath, file)
		err = addToZip(zipWriter, basePath, filePath)
		if err != nil {
			return fmt.Errorf("failed to add file %s to zip: %v", filePath, err)
		}
	}

	return nil
}
