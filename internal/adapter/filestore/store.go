// Package filestore persists generated images to the local filesystem.
package filestore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"cf-genimage/internal/domain"
)

// Store implements domain.ImageStore on the local filesystem.
type Store struct{}

// New builds a Store.
func New() *Store { return &Store{} }

// Save writes the image bytes to path, creating parent directories as needed.
func (Store) Save(ctx context.Context, path string, img domain.Image) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("filestore: %w", err)
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("filestore: create directories: %w", err)
		}
	}
	if err := os.WriteFile(path, img.Data, 0o644); err != nil {
		return fmt.Errorf("filestore: write %s: %w", path, err)
	}
	return nil
}
