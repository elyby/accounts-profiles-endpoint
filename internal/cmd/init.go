package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

func initConfig() *viper.Viper {
	v := viper.GetViper()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return v
}
