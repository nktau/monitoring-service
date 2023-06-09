package httplayer

import (
	"github.com/nktau/monitoring-service/internal/applayer"
	"log"
	"net/http"
)

type httpAPI struct {
	app applayer.App
}

func New(appLayer applayer.App) httpAPI {
	return httpAPI{app: appLayer}
}

func (api httpAPI) SetupRoutesAndStart() {
	http.Handle("/update/", http.HandlerFunc(api.update))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))

}
