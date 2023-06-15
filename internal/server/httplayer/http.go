package httplayer

import (
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
	api.router.Handle("/"+handlePathUpdate+"/*", validateUpdateValueHandlersRequest(http.HandlerFunc(api.update)))
	// /value/*
	api.router.Handle("/"+handlePathValue+"/*", validateUpdateValueHandlersRequest(http.HandlerFunc(api.value)))
	api.router.Get("/", api.root)
	return api
}

func (api *httpAPI) Start(addr string) error {
	err := http.ListenAndServe(addr, api.router)
	return err
}
