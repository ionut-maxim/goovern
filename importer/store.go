package importer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"

	"github.com/ionut-maxim/goovern/ckan"
)

type ResourceStore interface {
	Save(ctx context.Context, resource ckan.Resource) error
	Load(ctx context.Context, resource ckan.Resource) (io.ReadCloser, error)
}

type FSResourceStore struct {
	path   string
	logger *slog.Logger
}

func NewFSResourceStore(path string, logger *slog.Logger) (*FSResourceStore, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	return &FSResourceStore{
		path:   path,
		logger: logger,
	}, nil
}

func (s *FSResourceStore) Save(ctx context.Context, resource ckan.Resource) error {
	logger := s.logger.With("resource_id", resource.Id, "resource_name", resource.Name)

	if !resource.PackageId.Valid {
		return fmt.Errorf("invalid package id")
	}
	path := filepath.Join(s.path, resource.PackageId.UUID.String())
	filePath := filepath.Join(path, resource.Name)

	// Check if file is already fully downloaded
	if _, err := os.Stat(filePath); err == nil {
		logger.Info("File already exists, skipping download")
		return nil
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		logger.Error("Failed to create directory", "path", path, "error", err)
		return err
	}

	// Use a temporary file for partial downloads
	tempPath := filePath + ".tmp"

	// Check if we have a partial download to resume
	var existingSize int64
	if stat, err := os.Stat(tempPath); err == nil {
		existingSize = stat.Size()
		logger.Info("Found partial download", "existing_size", humanize.Bytes(uint64(existingSize)))
	}

	// Create HTTP request with Range header for resumable download
	logger.Debug("Creating HTTP request")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resource.Url, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request", "error", err)
		return err
	}

	// If we have a partial file, request only the remaining bytes
	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
		logger.Debug("Requesting resume from byte offset", "offset", existingSize)
	}

	logger.Debug("Sending HTTP request")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("HTTP request failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	// Ensure response body is closed when context is cancelled
	go func() {
		<-ctx.Done()
		resp.Body.Close()
	}()

	logger.Debug("Received HTTP response", "status_code", resp.StatusCode)

	// Check response status
	// 200 = full content, 206 = partial content (resume), 416 = already complete
	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		logger.Info("File is already complete on server, renaming temp file")
		if err = os.Rename(tempPath, filePath); err != nil {
			logger.Error("Failed to rename file", "error", err)
			return err
		}
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		logger.Error("Unexpected HTTP status", "status_code", resp.StatusCode)
		return err
	}

	// Open file for appending if resuming, or create new if starting fresh
	var file *os.File
	if existingSize > 0 && resp.StatusCode == http.StatusPartialContent {
		// Resuming - open for append
		file, err = os.OpenFile(tempPath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
	} else {
		// Starting fresh - create new file
		file, err = os.Create(tempPath)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// Use a context-aware copy that can be cancelled
	if err = copyWithContext(ctx, file, resp.Body, existingSize, s.logger.With("resource_name", resource.Name, "resource_id", resource.Id)); err != nil {
		// Leave partial file in place for resume on next retry
		return err
	}

	// Only rename to final path if download completed successfully
	if err = os.Rename(tempPath, filePath); err != nil {
		return err
	}

	return nil
}

func (s *FSResourceStore) Load(ctx context.Context, resource ckan.Resource) (io.ReadCloser, error) {
	if !resource.PackageId.Valid {
		return nil, fmt.Errorf("invalid package id")
	}
	path := filepath.Join(s.path, resource.PackageId.UUID.String())
	return os.Open(filepath.Join(path, resource.Name))
}

// copyWithContext copies from src to dst while respecting context cancellation
// existingSize is the number of bytes already downloaded (for resume tracking)
func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, existingSize int64, logger *slog.Logger) error {
	size := 10 * humanize.MByte // 10MB buffer
	buf := make([]byte, size)
	var totalWritten int64
	var lastLoggedAt int64
	logInterval := int64(10 * humanize.MByte) // Log every 10 MB

	if existingSize > 0 {
		logger.Info("Resuming download", "already_downloaded", humanize.Bytes(uint64(existingSize)))
	}

	for {
		select {
		case <-ctx.Done():
			logger.Warn("Download cancelled", "downloaded_this_session", humanize.Bytes(uint64(totalWritten)), "total_on_disk", humanize.Bytes(uint64(existingSize+totalWritten)))
			return ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
			totalWritten += int64(nw)

			// Only log every 10 MB
			if totalWritten-lastLoggedAt >= logInterval {
				logger.Info("Download progress",
					"session", humanize.Bytes(uint64(totalWritten)),
					"total", humanize.Bytes(uint64(existingSize+totalWritten)))
				lastLoggedAt = totalWritten
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			}
			break
		}
	}

	logger.Info("Download complete", "total", humanize.Bytes(uint64(existingSize+totalWritten)))
	return nil
}

type NoopResourceStore struct {
	logger *slog.Logger
}

func NewNoopResourceStore(logger *slog.Logger) *NoopResourceStore {
	return &NoopResourceStore{logger: logger.With("store", "noop")}
}

func (s *NoopResourceStore) Save(ctx context.Context, resource ckan.Resource) error {
	s.logger.Info("Saving resourceGetter", "package_id", resource.PackageId, "resource_name", resource.Name)
	return nil
}

func (s *NoopResourceStore) Load(ctx context.Context, resource ckan.Resource) (io.ReadCloser, error) {
	s.logger.Info("Loading resourceGetter", "package_id", resource.PackageId, "resource_name", resource.Name)
	return &os.File{}, nil
}
