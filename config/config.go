package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func DeclareFlag(v *viper.Viper, c *cobra.Command, name string, shorthand string, defaultVal any, usage string) {
	flags := c.PersistentFlags()

	switch defaultVal := defaultVal.(type) {
	case string:
		flags.StringP(name, shorthand, defaultVal, usage)
	}

	err := v.BindPFlag(name, flags.Lookup(name))
	if err != nil {
		panic(err)
	}
}

func GetAll(v *viper.Viper) map[string]interface{} {
	return v.AllSettings()
}

func Get[T any](v *viper.Viper) (T, error) {
	var t T
	err := v.Unmarshal(&t, func(config *mapstructure.DecoderConfig) {
		config.TagName = "json"
	})
	if err != nil {
		return t, err
	}
	return t, nil
}

func IsDebug() bool {
	s, ok := os.LookupEnv("DEBUG")
	if ok && s != "false" {
		return true
	}

	return false
}

type Options struct {
	EnvNestSeparator string            // 默认为 _，即 DB_HOST=v 这样的 env 会被识别为 {db: {host: v}}
	AutoBindEnv      func(string) bool // 自动加载 env 到 allSetting 中
}

func NewViper(p Options) *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()

	if p.EnvNestSeparator == "" {
		p.EnvNestSeparator = "_"
	}

	if p.EnvNestSeparator != "" {
		v.SetEnvKeyReplacer(strings.NewReplacer(".", p.EnvNestSeparator))
	}

	if p.AutoBindEnv != nil {
		for _, env := range os.Environ() {
			// 分割键值对，如 "KEY=VALUE"
			kv := strings.SplitN(env, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key := kv[0]
			if p.AutoBindEnv(key) {
				if p.EnvNestSeparator != "" {
					key = strings.ReplaceAll(key, p.EnvNestSeparator, ".")
				}
				_ = v.BindEnv(key)
			}
		}
	}

	return v
}
