package httplayer

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"net/http"
)

const handlePathUpdate = "update"
const handlePathValue = "value"

type httpAPI struct {
	app    applayer.App
	router chi.Router
}

func New(appLayer applayer.App) httpAPI {

	api := httpAPI{
		app:    appLayer,
		router: chi.NewRouter(),
	}
	// /update/*
	api.router.Handle(fmt.Sprintf("/%s/*", handlePathUpdate), setHeaders(http.HandlerFunc(api.updateAndValueHandler)))
	// /value/*
	api.router.Handle(fmt.Sprintf("/%s/*", handlePathValue), setHeaders(http.HandlerFunc(api.updateAndValueHandler)))
	api.router.Get("/", api.root)
	return api
}

func (api *httpAPI) Start(addr string) error {
	err := http.ListenAndServe(addr, api.router)
	return err
}
