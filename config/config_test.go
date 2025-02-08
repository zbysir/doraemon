package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestAutoBindEnv(t *testing.T) {
	os.Setenv("DB_PORT", "5432")
	v := NewViper(Options{
		EnvNestSeparator: "_",
		AutoBindEnv:      nil,
	})

	//v.BindEnv("db.port")

	all := v.AllSettings()
	t.Log(all) // empty if not v.BindEnv("db.port")
	assert.Equal(t, map[string]any{}, all)

	{
		v := NewViper(Options{
			EnvNestSeparator: "_",
			AutoBindEnv: func(s string) bool {
				return strings.HasPrefix(s, "DB_")
			}, // auto bind env if key HasPrefix
		})

		all := v.AllSettings()
		t.Log(all) // map[db:map[port:5432]]

		assert.Equal(t, map[string]any{"port": "5432"}, all["db"])
	}

}
