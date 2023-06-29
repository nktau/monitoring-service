package httplayer

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"go.uber.org/zap"
	"net/http"
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
	// /update/*
	api.router.Use(setHeaders)
	api.router.Use(api.withLogging)
	api.router.Post(fmt.Sprintf("/%s/*", handlePathUpdate), api.update)
	// /value/*
	api.router.Get(fmt.Sprintf("/%s/*", handlePathValue), api.value)
	api.router.Get("/", api.root)
	return api
}

func (api *httpAPI) Start(addr string) error {
	err := http.ListenAndServe(addr, api.router)
	return err
}
