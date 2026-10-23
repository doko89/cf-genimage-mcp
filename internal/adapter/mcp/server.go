// Package mcp adapts the generate-image use case to the Model Context Protocol
// over stdio. The model and step count stay hardcoded server-side and are not
// exposed as tool parameters.
package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cf-genimage/internal/domain"
	"cf-genimage/internal/usecase"
)

// Server identity.
const (
	ServerName    = "cf-genimage"
	ServerVersion = "0.1.0"
)

// GenerateImageInput is the JSON schema of the generate_image tool.
type GenerateImageInput struct {
	Prompt     string `json:"prompt" jsonschema:"Detailed description of the image to generate, e.g. 'flat vector illustration of a laptop on a desk, pastel colors'"`
	OutputPath string `json:"output_path,omitempty" jsonschema:"Optional local file path to save the WebP image (e.g. /tmp/cat.webp). When empty the image is only returned in the response."`
}

// GenerateImageOutput is the structured output of the generate_image tool.
type GenerateImageOutput struct {
	MIMEType  string `json:"mime_type" jsonschema:"MIME type of the generated image"`
	Bytes     int    `json:"bytes" jsonschema:"Size of the generated image in bytes"`
	SavedPath string `json:"saved_path,omitempty" jsonschema:"File path where the image was saved; empty when not saved"`
	Model     string `json:"model" jsonschema:"Model used for generation"`
	Steps     int    `json:"steps" jsonschema:"Number of diffusion steps used"`
}

// Server is the cf-genimage MCP server.
type Server struct {
	uc  *usecase.GenerateImage
	srv *sdk.Server
}

// New builds the MCP server and registers the generate_image tool.
func New(uc *usecase.GenerateImage) *Server {
	s := &Server{uc: uc}
	s.srv = sdk.NewServer(&sdk.Implementation{Name: ServerName, Version: ServerVersion}, nil)
	sdk.AddTool(s.srv, &sdk.Tool{
		Name:        "generate_image",
		Description: "Generate an image from a text prompt with Cloudflare Workers AI (flux-1-schnell, 4 steps). Returns the image as WebP inline and optionally saves it to output_path.",
	}, s.handleGenerateImage)
	return s
}

// Run serves MCP over stdin/stdout until ctx is cancelled or the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	if err := s.srv.Run(ctx, &sdk.StdioTransport{}); err != nil {
		return fmt.Errorf("mcp server: %w", err)
	}
	return nil
}

func (s *Server) handleGenerateImage(ctx context.Context, _ *sdk.CallToolRequest, in GenerateImageInput) (*sdk.CallToolResult, GenerateImageOutput, error) {
	res, err := s.uc.Execute(ctx, domain.GenerateRequest{Prompt: in.Prompt, OutputPath: in.OutputPath})
	if err != nil {
		return &sdk.CallToolResult{
			Content: []sdk.Content{&sdk.TextContent{Text: "Image generation failed: " + err.Error()}},
			IsError: true,
		}, GenerateImageOutput{}, nil
	}
	out := GenerateImageOutput{
		MIMEType:  res.Image.MIMEType,
		Bytes:     len(res.Image.Data),
		SavedPath: res.SavedPath,
		Model:     domain.ModelID,
		Steps:     domain.Steps,
	}
	summary := fmt.Sprintf("Generated %d-byte %s image with %s (steps=%d).", out.Bytes, out.MIMEType, out.Model, out.Steps)
	if out.SavedPath != "" {
		summary += " Saved to " + out.SavedPath + "."
	}
	return &sdk.CallToolResult{
		Content: []sdk.Content{
			&sdk.TextContent{Text: summary},
			&sdk.ImageContent{Data: res.Image.Data, MIMEType: res.Image.MIMEType},
		},
	}, out, nil
}
