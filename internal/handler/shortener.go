package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SZabrodskii/url-shortener/internal/config"
	"github.com/SZabrodskii/url-shortener/internal/server"
	"github.com/SZabrodskii/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type ShortenerDescription struct {
	Shorten httpbara.Route `route:"POST /"`
	Resolve httpbara.Route `route:"GET /:id"`
}

type Shortener struct {
	ShortenerDescription

	svc     *service.Shortener
	baseURL config.BaseURL
}

type newShortenerIn struct {
	fx.In

	Svc     *service.Shortener
	BaseURL config.BaseURL
}

func newShortener(in newShortenerIn) (server.AsHandlerOut, error) {
	h := &Shortener{
		svc:     in.Svc,
		baseURL: in.BaseURL,
	}
	hh, err := httpbara.AsHandler(h)
	if err != nil {
		return server.AsHandlerOut{}, err
	}
	return server.AsHandlerOut{Handler: hh}, nil
}

func (c *Shortener) Shorten(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusBadRequest, "bad request")
		return
	}
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		ctx.String(http.StatusBadRequest, "bad request")
		return
	}
	id, err := c.svc.Shorten(raw)
	if err != nil {
		ctx.String(http.StatusBadRequest, "bad request")
		return
	}
	short := string(c.baseURL) + "/" + id

	ctx.Data(http.StatusCreated, "text/plain", []byte(short))
}

func (c *Shortener) Resolve(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.String(http.StatusBadRequest, "bad request")
		return
	}
	originalURL, err := c.svc.Resolve(id)
	if errors.Is(err, service.ErrNotFound) {
		ctx.String(http.StatusBadRequest, "bad request")
		return
	}
	if err != nil {
		ctx.String(http.StatusBadRequest, "bad request")
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func Provide() fx.Option {
	return fx.Provide(newShortener)
}
