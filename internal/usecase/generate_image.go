// Package usecase implements the application business rules of cf-genimage.
package usecase

import (
	"context"
	"fmt"
	"strings"

	"cf-genimage/internal/domain"
)

// GenerateImage orchestrates image generation and optional persistence.
type GenerateImage struct {
	gen   domain.ImageGenerator
	store domain.ImageStore
}

// NewGenerateImage builds the generate-image use case.
func NewGenerateImage(gen domain.ImageGenerator, store domain.ImageStore) *GenerateImage {
	return &GenerateImage{gen: gen, store: store}
}

// Execute validates the request, generates the image, and saves it when OutputPath is set.
func (u *GenerateImage) Execute(ctx context.Context, req domain.GenerateRequest) (domain.GenerateResult, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return domain.GenerateResult{}, domain.ErrEmptyPrompt
	}
	if err := ctx.Err(); err != nil {
		return domain.GenerateResult{}, fmt.Errorf("generate image: %w", err)
	}
	img, err := u.gen.Generate(ctx, req.Prompt)
	if err != nil {
		return domain.GenerateResult{}, fmt.Errorf("generate image: %w", err)
	}
	res := domain.GenerateResult{Image: img}
	if req.OutputPath == "" {
		return res, nil
	}
	if err := u.store.Save(ctx, req.OutputPath, img); err != nil {
		return domain.GenerateResult{}, fmt.Errorf("save image to %s: %w", req.OutputPath, err)
	}
	res.SavedPath = req.OutputPath
	return res, nil
}
