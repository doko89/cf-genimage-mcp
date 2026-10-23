// Command cf-genimage is an MCP server (stdio) that generates images with
// Cloudflare Workers AI (flux-1-schnell, 4 steps).
//
// Credentials come from CF_ACCOUNT_ID and CF_API_TOKEN environment variables.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"cf-genimage/internal/adapter/cloudflare"
	"cf-genimage/internal/adapter/filestore"
	"cf-genimage/internal/adapter/mcp"
	"cf-genimage/internal/adapter/webpconv"
	"cf-genimage/internal/config"
	"cf-genimage/internal/usecase"
)

func main() {
	// log defaults to stderr: never write logs to stdout, it carries MCP frames.
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatalf("cf-genimage: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cf, err := cloudflare.New(cfg.AccountID, cfg.APIToken)
	if err != nil {
		return err
	}
	uc := usecase.NewGenerateImage(webpconv.New(cf), filestore.New())
	if err := mcp.New(uc).Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
