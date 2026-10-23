package webpconv_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"

	"cf-genimage/internal/adapter/webpconv"
	"cf-genimage/internal/domain"
)

type fakeGenerator struct {
	img domain.Image
	err error
}

func (f *fakeGenerator) Generate(_ context.Context, _ string) (domain.Image, error) {
	return f.img, f.err
}

// testJPEG encodes a small JPEG fixture with the standard library.
func testJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := range 16 {
		for x := range 16 {
			img.Set(x, y, color.NRGBA{R: uint8(x * 16), G: uint8(y * 16), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

// isWebP checks the RIFF....WEBP container magic.
func isWebP(b []byte) bool {
	return len(b) > 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP"
}

func TestGenerate_Success(t *testing.T) {
	inner := &fakeGenerator{img: domain.Image{Data: testJPEG(t), MIMEType: "image/jpeg"}}

	out, err := webpconv.New(inner).Generate(context.Background(), "a cat")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.MIMEType != "image/webp" {
		t.Fatalf("mime = %q, want image/webp", out.MIMEType)
	}
	if !isWebP(out.Data) {
		t.Fatal("output bytes are not a WebP container")
	}
}

func TestGenerate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		img     domain.Image
		genErr  error
		wantErr string
	}{
		{name: "inner error", genErr: errors.New("boom"), wantErr: "boom"},
		{name: "bad bytes", img: domain.Image{Data: []byte("not-an-image")}, wantErr: "decode source"},
		{name: "empty bytes", img: domain.Image{}, wantErr: "decode source"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inner := &fakeGenerator{img: tt.img, err: tt.genErr}
			_, err := webpconv.New(inner).Generate(context.Background(), "a cat")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestGenerate_Cancelled(t *testing.T) {
	inner := &fakeGenerator{img: domain.Image{Data: testJPEG(t)}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := webpconv.New(inner).Generate(ctx, "a cat"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
