package upload

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"io"
	"l7-cloud-builder/config"
	"l7-cloud-builder/logger"
	"mime/multipart"
	"net/http"
	"os"
)

// prepareMultipartForm creates a new multipart form writer, adds form fields from params,
// and adds a form file with the specified fileInfo. Returns the content type of the
// multipart form, the opening header, and the closing boundary.
func prepareMultipartForm(params map[string]string, fileInfo os.FileInfo) (string, []byte, []byte, error) {
	multipartFormBuffer := &bytes.Buffer{}
	multipartFormWriter := multipart.NewWriter(multipartFormBuffer)

	for key, value := range params {
		err := multipartFormWriter.WriteField(key, value)
		if err != nil {
			return "", nil, nil, err
		}
	}

	_, err := multipartFormWriter.CreateFormFile("file", fileInfo.Name())
	if err != nil {
		return "", nil, nil, err
	}

	contentType := multipartFormWriter.FormDataContentType()
	headerBytes := multipartFormBuffer.Bytes()

	err = multipartFormWriter.Close()
	if err != nil {
		return "", nil, nil, err
	}

	boundaryBytes := multipartFormBuffer.Bytes()

	return contentType, headerBytes, boundaryBytes, nil
}

// uploadFileInChunks writes the file contents to a pipe writer in chunks.
// The function takes the following arguments:
// - file: *os.File that represents the file to be uploaded
// - pipeWriter: *io.PipeWriter used to write the file contents
// - openingHeader: []byte that represents the opening header of the multipart form
// - closingBoundary: []byte that represents the closing boundary of the multipart form
// - chunkSize: int that represents the size of the chunks to be read from the file and written to the pipe
// The function performs the following steps:
// - Write the openingHeader to the pipeWriter
// - Read the file in chunks and write the chunks to the pipeWriter
// - Write the closingBoundary to the pipeWriter
// - Close the pipeWriter
func uploadFileInChunks(file *os.File, pipeWriter *io.PipeWriter, openingHeader, closingBoundary []byte, chunkSize int) {
	defer func(pipeWriter *io.PipeWriter) {
		err := pipeWriter.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close the pipe writer: %v", err)
		}
	}(pipeWriter)

	_, err := pipeWriter.Write(openingHeader)
	if err != nil {
		logger.Logger.Errorf("failed to write the opening header to the multipart form: %v", err)
	}

	buffer := make([]byte, chunkSize)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logger.Logger.Errorf("failed to read from the file pipe reader: %v", err)
			}
			break
		}
		_, err = pipeWriter.Write(buffer[:n])
		if err != nil {
			logger.Logger.Errorf("failed to write file bytes to the multipart form: %v", err)
		}
	}

	_, err = pipeWriter.Write(closingBoundary)
	if err != nil {
		logger.Logger.Errorf("failed to write the closing boundary to the multipart form: %v", err)
	}
}

// uploadEntityFile uploads a file to the API for storage in association with an entity.
// The function takes the following arguments:
// - job: JobMetadata that contains information about the job
// - entityId: UUID of the entity to associate the file with
// - fileType: a string representing the type of the file
// - fileMime: the MIME type of the file
// - path: the local file system path of the file
// - originalPath: the original path of the file
// - params: a map of string key-value pairs that represents additional parameters to send with the request
// The function performs the following steps:
// - Validate the entity id
// - Construct the request URL
// - Open and read the file
// - Create a multipart form and write the form fields and file contents to a pipe
// - Send an HTTP PUT request with the pipe reader as the request body
// - Handle the response and return any errors
func uploadEntityFile(ctx context.Context, entityId uuid.UUID, fileType, fileMime, target, platform, path, originalPath string, params map[string]string) error {
	const chunkSize = 100 * 1024 * 1024 // 100MiB

	// Validate the entity id
	if entityId.IsNil() {
		return fmt.Errorf("invalid job entity id")
	}

	// Construct request URL
	reqUrl := fmt.Sprintf("%s/entities/%s/files/upload?type=%s&mime=%s&deployment=%s&platform=%s&original-path=%s", config.Api.Url, entityId.String(), fileType, fileMime, target, platform, originalPath)

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	// Defer closing the file
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close file: %v", err)
		}
	}(file)

	// Get the file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	// Prepare the multipart form
	contentType, openingHeader, closingBoundary, err := prepareMultipartForm(params, fileInfo)
	if err != nil {
		return err
	}

	// Create a pipe and write the multipart form to it
	pipeReader, pipeWriter := io.Pipe()
	defer func(pipeReader *io.PipeReader) {
		err := pipeReader.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close pipe reader: %v", err)
		}
	}(pipeReader)

	// Write the multipart form to the pipe in chunks
	go uploadFileInChunks(file, pipeWriter, openingHeader, closingBoundary, chunkSize)

	// Create the request using the pipe reader as the request body
	totalSize := int64(len(openingHeader)) + fileInfo.Size() + int64(len(closingBoundary))
	req, err := http.NewRequest("PUT", reqUrl, pipeReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the request headers, here we expect the valid token to be set in the config (to be logged in) before calling this function
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = totalSize
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Api.Token))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	// Defer closing the response body
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	// Handle the response
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read the response body: %v", err)
		}
		return fmt.Errorf("failed to upload a file, status code: %d, content: %s", resp.StatusCode, string(body))
	}

	return nil
}
