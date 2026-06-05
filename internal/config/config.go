package config

import (
	"flag"
	"os"

	"go.uber.org/fx"
)

type ServerAddr string
type BaseURL string

type applicationConfig struct {
	fx.Out

	ServerAddr ServerAddr
	BaseURL    BaseURL
}

func newConfig() (applicationConfig, error) {
	var serverAddr string
	var baseURL string

	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)
	fs.StringVar(&serverAddr, "a", ":8080", "...")
	fs.StringVar(&baseURL, "b", "http://localhost:8080", "...")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return applicationConfig{}, err
	}

	return applicationConfig{
		ServerAddr: ServerAddr(serverAddr),
		BaseURL:    BaseURL(baseURL),
	}, nil
}

func Provide() fx.Option {
	return fx.Provide(newConfig)
}
