package httplayer

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

const handlePathUpdate = "update"
const handlePathValue = "value"

type httpAPI struct {
	app    applayer.App
	router chi.Router
	logger *zap.Logger
}

func New(appLayer applayer.App, logger *zap.Logger) httpAPI {

	api := httpAPI{
		app:    appLayer,
		router: chi.NewRouter(),
		logger: logger,
	}
	api.router.Use(api.withLogging)
	api.router.Use(middleware.Compress(5, "application/json", "text/html"))
	api.router.Post(fmt.Sprintf("/%s/*", handlePathUpdate), api.whichOfUpdateHandlerUse)
	api.router.Get(fmt.Sprintf("/%s/*", handlePathValue), api.valuePlainText)
	api.router.Post(fmt.Sprintf("/%s/*", handlePathValue), api.valueJSON)
	api.router.Get("/", api.root)
	return api
}

func (api *httpAPI) Start(addr string) error {
	err := http.ListenAndServe(addr, api.router)
	return err
}

var ErrMethodNotAllowed = errors.New("method not allowed")

func getRequestURLSlice(request string) []string {
	return strings.Split(strings.TrimLeft(request, "/"), "/")
}

func (api *httpAPI) readBody(r *http.Request) (io.Reader, error) {
	var reader io.Reader
	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}
	return reader, nil
}

type metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
