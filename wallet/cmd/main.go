package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wallet_service/internal/app/api"

	wal "wallet_service/internal/app/wallet"
)

func main() {
	ctx := context.Background()
	app := wal.NewWalletApp(ctx)
	s := api.NewServer(app)

	defer s.Shutdown()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		app.Logger.Println("Starting server")

		if err := s.Run(ctx); err != nil && err != http.ErrServerClosed {
			s.App.Logger.Info(err)

			return
		}
	}()

	<-stop

	app.Logger.Println("Server shutdown")
}
