package sentry

import (
	"strings"

	"github.com/getsentry/sentry-go"
	sentryGin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"ely.by/accounts-profiles-endpoint/internal/version"
)

func InitWithConfig(config *viper.Viper) error {
	config.SetDefault("sentry.enable_tracing", false)
	config.SetDefault("sentry.traces_sample_rate", 1.0)
	config.SetDefault("sentry.profiles_sample_rate", 1.0)

	sampleRate := config.GetFloat64("sentry.traces_sample_rate")

	return sentry.Init(sentry.ClientOptions{
		Dsn:           viper.GetString("sentry.dsn"),
		EnableTracing: viper.GetBool("sentry.enable_tracing"),
		TracesSampler: func(ctx sentry.SamplingContext) float64 {
			if !strings.Contains(ctx.Span.Name, "/api") {
				return 0
			}

			return sampleRate
		},
		ProfilesSampleRate: config.GetFloat64("sentry.profiles_sample_rate"),
		Release:            version.Version(),
		Environment:        config.GetString("sentry.environment"),
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			nDeleted := 0
			for i, integration := range integrations {
				if integration.Name() == "Modules" {
					integrations[i] = integrations[len(integrations)-(nDeleted+1)]
					nDeleted++
				}
			}

			return integrations[:len(integrations)-nDeleted]
		},
	})
}

// It seems like this must be a part of the sentrygin package, but it is not, so implement it ourselves
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sentryHub := sentryGin.GetHubFromContext(c)
		if sentryHub != nil {
			return
		}

		for _, err := range c.Errors {
			sentryHub.CaptureException(err)
		}
	}
}
