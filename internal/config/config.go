package config

import "go.uber.org/fx"

type ServerAddr string
type BaseURL string

type applicationConfig struct {
	fx.Out

	ServerAddr ServerAddr
	BaseURL    BaseURL
}

func newConfig() applicationConfig {
	return applicationConfig{
		ServerAddr: ":8080",
		BaseURL:    "http://localhost:8080",
	}
}

func Provide() fx.Option {
	return fx.Provide(newConfig)
}
