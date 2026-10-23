package mcp

import (
	"context"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cf-genimage/internal/domain"
	"cf-genimage/internal/usecase"
)

type fakeGenerator struct {
	img domain.Image
	err error
}

func (f *fakeGenerator) Generate(_ context.Context, _ string) (domain.Image, error) {
	return f.img, f.err
}

type fakeStore struct {
	path string
}

func (f *fakeStore) Save(_ context.Context, path string, _ domain.Image) error {
	f.path = path
	return nil
}

func TestHandleGenerateImage_Success(t *testing.T) {
	img := domain.Image{Data: []byte("fake-jpeg-bytes"), MIMEType: "image/jpeg"}
	uc := usecase.NewGenerateImage(&fakeGenerator{img: img}, &fakeStore{})
	s := New(uc)

	res, out, err := s.handleGenerateImage(context.Background(), &sdk.CallToolRequest{}, GenerateImageInput{
		Prompt:     "a cat",
		OutputPath: "/tmp/x.jpg",
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatal("IsError = true, want false")
	}
	if len(res.Content) != 2 {
		t.Fatalf("content items = %d, want 2 (text + image)", len(res.Content))
	}
	if _, ok := res.Content[0].(*sdk.TextContent); !ok {
		t.Fatalf("content[0] = %T, want *sdk.TextContent", res.Content[0])
	}
	imgContent, ok := res.Content[1].(*sdk.ImageContent)
	if !ok {
		t.Fatalf("content[1] = %T, want *sdk.ImageContent", res.Content[1])
	}
	if string(imgContent.Data) != "fake-jpeg-bytes" || imgContent.MIMEType != "image/jpeg" {
		t.Fatalf("image content = %+v, want fake bytes + image/jpeg", imgContent)
	}
	if out.Bytes != len(img.Data) || out.Model != domain.ModelID || out.Steps != domain.Steps {
		t.Fatalf("structured output = %+v", out)
	}
	if out.SavedPath != "/tmp/x.jpg" {
		t.Fatalf("savedPath = %q, want /tmp/x.jpg", out.SavedPath)
	}
}

func TestHandleGenerateImage_Error(t *testing.T) {
	uc := usecase.NewGenerateImage(&fakeGenerator{}, &fakeStore{})
	s := New(uc)

	res, _, err := s.handleGenerateImage(context.Background(), &sdk.CallToolRequest{}, GenerateImageInput{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Fatal("IsError = false, want true")
	}
	text, ok := res.Content[0].(*sdk.TextContent)
	if !ok {
		t.Fatalf("content[0] = %T, want *sdk.TextContent", res.Content[0])
	}
	if !strings.Contains(text.Text, "prompt must not be empty") {
		t.Fatalf("text = %q, want prompt error", text.Text)
	}
}
