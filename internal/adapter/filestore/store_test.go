package filestore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cf-genimage/internal/adapter/filestore"
	"cf-genimage/internal/domain"
)

func TestSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "img.jpg")

	if err := filestore.New().Save(context.Background(), path, domain.Image{Data: []byte("jpeg-bytes")}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != "jpeg-bytes" {
		t.Fatalf("data = %q, want %q", got, "jpeg-bytes")
	}
}

func TestSave_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := filestore.New().Save(ctx, filepath.Join(t.TempDir(), "x.jpg"), domain.Image{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
