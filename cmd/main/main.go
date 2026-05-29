package main

import (
	"context"
	app "github.com/Fitnow08/fitnow-backend-sso/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	defer cancel()
	apps, err := app.NewApp(ctx)
	if err != nil {
		log.Fatal(err)
	}
	apps.Log.Info("Starting apps..")
	errCh := make(chan error, 1)
	go func() {
		if err := apps.GRPCServer.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		apps.Log.Error("grpc server failed", "err", err)
	}
	apps.GRPCServer.Stop()
	defer apps.CancelLogger()
	if err := apps.DB.Close(); err != nil {
		apps.Log.Error("failed to close databases", "err", err.Error())
	}
	apps.Log.Info("apps stopped")
}
