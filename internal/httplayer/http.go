package httplayer

import (
	"github.com/go-chi/chi/v5"
	"github.com/nktau/monitoring-service/internal/applayer"
	"net/http"
)

type httpAPI struct {
	app    applayer.App
	router chi.Router
}

func New(appLayer applayer.App) httpAPI {

	api := httpAPI{
		app:    appLayer,
		router: chi.NewRouter(),
	}
	api.router.Post("/update/*", api.update)
	api.router.Get("/value/*", api.value)
	api.router.Get("/", api.root)
	return api
}

func (api *httpAPI) Start() {
	http.ListenAndServe("localhost:8080", api.router)
}
