// Package domain defines the core entities, ports, and fixed generation policy
// of cf-genimage. It has no dependencies on outside layers.
package domain

import (
	"context"
	"errors"
)

// Fixed generation policy: the model and step count are hardcoded by design
// and are never exposed as tool parameters.
const (
	// ModelID is the Cloudflare Workers AI model used for every generation.
	ModelID = "@cf/black-forest-labs/flux-1-schnell"
	// Steps is the number of diffusion steps used for every generation.
	Steps = 4
	// WebPQuality is the lossy WebP quality (0-100) used for every output image.
	WebPQuality = 90
)

// ErrEmptyPrompt is returned when the prompt is blank.
var ErrEmptyPrompt = errors.New("prompt must not be empty")

// Image is a generated raster image with raw bytes.
type Image struct {
	Data     []byte
	MIMEType string
}

// GenerateRequest is the input of the generate-image use case.
type GenerateRequest struct {
	Prompt string
	// OutputPath optionally saves the image to a local file. Empty means no saving.
	OutputPath string
}

// GenerateResult is the output of the generate-image use case.
type GenerateResult struct {
	Image Image
	// SavedPath echoes GenerateRequest.OutputPath when the image was saved.
	SavedPath string
}

// ImageGenerator generates an image from a prompt (driven adapter port).
type ImageGenerator interface {
	Generate(ctx context.Context, prompt string) (Image, error)
}

// ImageStore persists an image to a local path (driven adapter port).
type ImageStore interface {
	Save(ctx context.Context, path string, img Image) error
}
