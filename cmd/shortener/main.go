package main

import (
	"github.com/SZabrodskii/url-shortener/internal/config"
	"github.com/SZabrodskii/url-shortener/internal/handler"
	"github.com/SZabrodskii/url-shortener/internal/repository"
	"github.com/SZabrodskii/url-shortener/internal/server"
	"github.com/SZabrodskii/url-shortener/internal/service"
	"go.uber.org/fx"
)

func main() {
	fx.New(createApp()).Run()
}

func createApp() fx.Option {
	return fx.Options(
		config.Provide(),
		repository.Provide(),
		service.Provide(),
		handler.Provide(),
		server.ProvideHTTPModule(),
	)
}
