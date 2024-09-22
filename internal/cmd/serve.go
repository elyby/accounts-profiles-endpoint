package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/etherlabsio/healthcheck/v2"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"go.uber.org/multierr"

	db "ely.by/sessionserver/internal/db/mysql"
	"ely.by/sessionserver/internal/http"
	"ely.by/sessionserver/internal/logging/sentry"
	"ely.by/sessionserver/internal/services/chrly"
	"ely.by/sessionserver/internal/services/signer"
)

func Serve() error {
	config := initConfig()

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, os.Kill)

	var errors, err error
	err = sentry.InitWithConfig(config)
	if err != nil {
		return fmt.Errorf("unable to initialize Sentry: %w", err)
	}

	mysql, err := db.NewWithConfig(config)
	errors = multierr.Append(errors, err)

	texturesProvider, err := chrly.NewWithConfig(config)
	errors = multierr.Append(errors, err)

	signerService, err := signer.NewWithConfig(config)
	errors = multierr.Append(errors, err)

	if errors != nil {
		return errors
	}

	if config.GetBool("debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	r.Use(sentry.ErrorMiddleware())
	r.Use(http.ErrorMiddleware())

	r.GET("/healthcheck", gin.WrapH(healthcheck.Handler(
		healthcheck.WithChecker("mysql", healthcheck.CheckerFunc(mysql.Ping)),
	)))

	sessionserver := http.NewMojangApi(mysql, texturesProvider, signerService)
	sessionserver.DefineRoutes(r)

	server, err := http.NewServerWithConfig(config, r)
	if err != nil {
		return fmt.Errorf("unable to create a server: %w", err)
	}

	err = http.StartServer(ctx, server)
	if err != nil {
		return fmt.Errorf("unable to start a server: %w", err)
	}

	return nil
}
