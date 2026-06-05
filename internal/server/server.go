package server

import (
	"context"
	"net/http"

	"github.com/SZabrodskii/url-shortener/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type AsHandlerOut struct {
	fx.Out

	Handler *httpbara.Handler `group:"handlers"`
}

type engineIn struct {
	fx.In

	Handlers   []*httpbara.Handler `group:"handlers"`
	ServerAddr config.ServerAddr
	LC         fx.Lifecycle
}

func NewEngine(in engineIn) (httpbara.Engine, error) {
	g := gin.New()
	g.Use(gin.Recovery())

	g.NoRoute(func(c *gin.Context) { c.String(http.StatusBadRequest, "bad request - no such route") })
	g.NoMethod(func(c *gin.Context) { c.String(http.StatusBadRequest, "bad request - no such method") })

	eng, err := httpbara.New(in.Handlers, httpbara.WithGinEngine(g))
	if err != nil {
		return nil, err
	}

	in.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				_ = eng.Run(string(in.ServerAddr))
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return eng, nil
}

func invokeServer() fx.Option {
	return fx.Invoke(func(httpbara.Engine) {})
}

func ProvideHTTPModule() fx.Option {
	return fx.Module("httpbara",
		fx.Provide(NewEngine),
		invokeServer(),
	)
}
