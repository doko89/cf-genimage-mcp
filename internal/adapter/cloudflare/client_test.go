package cloudflare_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cf-genimage/internal/adapter/cloudflare"
	"cf-genimage/internal/domain"
)

func TestNew_Validation(t *testing.T) {
	if _, err := cloudflare.New("", "tok"); err == nil {
		t.Fatal("expected error for empty accountID")
	}
	if _, err := cloudflare.New("acc", ""); err == nil {
		t.Fatal("expected error for empty apiToken")
	}
}

func TestGenerate_Success(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody struct {
		Prompt string `json:"prompt"`
		Steps  int    `json:"steps"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"result":{"image":"` + base64.StdEncoding.EncodeToString([]byte("fake-jpeg")) + `"}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := cloudflare.New("acc123", "tok", cloudflare.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	img, err := c.Generate(context.Background(), "a cat")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if string(img.Data) != "fake-jpeg" {
		t.Fatalf("data = %q, want %q", img.Data, "fake-jpeg")
	}
	if img.MIMEType != "image/jpeg" {
		t.Fatalf("mime = %q, want image/jpeg", img.MIMEType)
	}
	wantPath := "/accounts/acc123/ai/run/" + domain.ModelID
	if gotPath != wantPath {
		t.Fatalf("path = %q, want %q", gotPath, wantPath)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("auth = %q, want %q", gotAuth, "Bearer tok")
	}
	if gotBody.Prompt != "a cat" {
		t.Fatalf("prompt = %q, want %q", gotBody.Prompt, "a cat")
	}
	if gotBody.Steps != domain.Steps {
		t.Fatalf("steps = %d, want %d", gotBody.Steps, domain.Steps)
	}
}

func TestGenerate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{name: "api error", status: 200, body: `{"success":false,"errors":[{"code":10000,"message":"bad request"}]}`, wantErr: "bad request"},
		{name: "http error", status: 500, body: `internal boom`, wantErr: "HTTP 500"},
		{name: "missing image", status: 200, body: `{"success":true,"result":{}}`, wantErr: "no result.image"},
		{name: "bad base64", status: 200, body: `{"success":true,"result":{"image":"!!!"}}`, wantErr: "decode image"},
		{name: "bad json", status: 200, body: `not-json`, wantErr: "decode response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)

			c, err := cloudflare.New("acc", "tok", cloudflare.WithBaseURL(srv.URL))
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			_, err = c.Generate(context.Background(), "a cat")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}
