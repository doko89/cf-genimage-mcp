package usecase_test

import (
	"context"
	"errors"
	"testing"

	"cf-genimage/internal/domain"
	"cf-genimage/internal/usecase"
)

type fakeGenerator struct {
	img   domain.Image
	err   error
	got   string
	calls int
}

func (f *fakeGenerator) Generate(_ context.Context, prompt string) (domain.Image, error) {
	f.got = prompt
	f.calls++
	return f.img, f.err
}

type fakeStore struct {
	err  error
	path string
	img  domain.Image
}

func (f *fakeStore) Save(_ context.Context, path string, img domain.Image) error {
	f.path, f.img = path, img
	return f.err
}

func TestExecute(t *testing.T) {
	img := domain.Image{Data: []byte("jpeg-bytes"), MIMEType: "image/jpeg"}

	tests := []struct {
		name      string
		req       domain.GenerateRequest
		genErr    error
		storeErr  error
		wantErr   bool
		wantSaved string
		wantBytes string
	}{
		{name: "ok without save", req: domain.GenerateRequest{Prompt: "a cat"}, wantBytes: "jpeg-bytes"},
		{name: "ok with save", req: domain.GenerateRequest{Prompt: "a cat", OutputPath: "/tmp/x.jpg"}, wantSaved: "/tmp/x.jpg", wantBytes: "jpeg-bytes"},
		{name: "empty prompt", req: domain.GenerateRequest{Prompt: "  "}, wantErr: true},
		{name: "generator error", req: domain.GenerateRequest{Prompt: "a cat"}, genErr: errors.New("boom"), wantErr: true},
		{name: "store error", req: domain.GenerateRequest{Prompt: "a cat", OutputPath: "/tmp/x.jpg"}, storeErr: errors.New("disk full"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &fakeGenerator{img: img, err: tt.genErr}
			store := &fakeStore{err: tt.storeErr}
			uc := usecase.NewGenerateImage(gen, store)

			res, err := uc.Execute(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(res.Image.Data) != tt.wantBytes {
				t.Fatalf("bytes = %q, want %q", res.Image.Data, tt.wantBytes)
			}
			if res.SavedPath != tt.wantSaved {
				t.Fatalf("savedPath = %q, want %q", res.SavedPath, tt.wantSaved)
			}
			if gen.got != tt.req.Prompt {
				t.Fatalf("prompt = %q, want %q", gen.got, tt.req.Prompt)
			}
		})
	}
}

func TestExecute_EmptyPromptIsDomainError(t *testing.T) {
	uc := usecase.NewGenerateImage(&fakeGenerator{}, &fakeStore{})
	_, err := uc.Execute(context.Background(), domain.GenerateRequest{})
	if !errors.Is(err, domain.ErrEmptyPrompt) {
		t.Fatalf("error = %v, want domain.ErrEmptyPrompt", err)
	}
}

func TestExecute_CancelledContext(t *testing.T) {
	gen := &fakeGenerator{}
	uc := usecase.NewGenerateImage(gen, &fakeStore{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := uc.Execute(ctx, domain.GenerateRequest{Prompt: "a cat"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if gen.calls != 0 {
		t.Fatalf("generator called %d times, want 0", gen.calls)
	}
}
