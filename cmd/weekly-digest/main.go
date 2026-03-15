package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tossy-yukky/codex-sample-project/internal/app"
	"github.com/tossy-yukky/codex-sample-project/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("run weekly digest: %v", err)
	}
}
