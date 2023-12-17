package upload

import (
	"context"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"net/url"
)

// ReleaseArchive uploads the release archive file to the cloud for storage
func ReleaseArchive(ctx context.Context, releaseId uuid.UUID, target, platform, path string, originalPath string, params map[string]string) error {
	if releaseId.IsNil() {
		return fmt.Errorf("invalid release id")
	}

	var (
		fileType = "release-archive"
		fileMime = "application/octet-stream"
	)

	// Try to detect MIME
	pMIME, err := mimetype.DetectFile(path)
	if err == nil {
		fileMime = pMIME.String()
	}

	fileMime = url.QueryEscape(fileMime)

	return uploadEntityFile(ctx, releaseId, fileType, fileMime, target, platform, path, originalPath, params)
}
