package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

func NewServerWithConfig(config *viper.Viper, handler http.Handler) (*http.Server, error) {
	config.SetDefault("http.host", "0.0.0.0")
	config.SetDefault("http.port", 8080)

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.GetString("http.host"), config.GetUint("http.port")),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}, nil
}

func StartServer(ctx context.Context, server *http.Server) error {
	srvErr := make(chan error, 1)
	go func() {
		slog.Debug("Starting the server", slog.String("addr", server.Addr))
		srvErr <- server.ListenAndServe()
		close(srvErr)
	}()

	select {
	case err := <-srvErr:
		return err
	case <-ctx.Done():
		slog.Info("Got stop signal, starting graceful shutdown")

		stopCtx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelFunc()

		_ = server.Shutdown(stopCtx)

		slog.Debug("Graceful shutdown succeed, exiting")
	}

	return nil
}
