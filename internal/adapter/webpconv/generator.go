// Package webpconv adapts any domain.ImageGenerator to emit WebP images,
// the preferred format for generated assets.
package webpconv

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // decode source images from the API
	_ "image/png"  // tolerate PNG sources too

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"

	"cf-genimage/internal/domain"
)

// Generator decorates an ImageGenerator, transcoding every result to WebP.
type Generator struct {
	inner   domain.ImageGenerator
	quality float32
}

// New wraps inner so Generate returns WebP-encoded images.
func New(inner domain.ImageGenerator) *Generator {
	return &Generator{inner: inner, quality: domain.WebPQuality}
}

// Generate implements domain.ImageGenerator.
func (g *Generator) Generate(ctx context.Context, prompt string) (domain.Image, error) {
	if err := ctx.Err(); err != nil {
		return domain.Image{}, fmt.Errorf("webpconv: %w", err)
	}
	src, err := g.inner.Generate(ctx, prompt)
	if err != nil {
		return domain.Image{}, fmt.Errorf("webpconv: generate: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(src.Data))
	if err != nil {
		return domain.Image{}, fmt.Errorf("webpconv: decode source image: %w", err)
	}
	opts, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, g.quality)
	if err != nil {
		return domain.Image{}, fmt.Errorf("webpconv: encoder options: %w", err)
	}
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, opts); err != nil {
		return domain.Image{}, fmt.Errorf("webpconv: encode webp: %w", err)
	}
	return domain.Image{Data: buf.Bytes(), MIMEType: "image/webp"}, nil
}
