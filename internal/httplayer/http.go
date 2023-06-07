package httplayer

import (
	"github.com/nktau/monitoring-service/internal/applayer"
	"log"
	"net/http"
)

type httpAPI struct {
	app applayer.App
}

func New(appLayer applayer.App) {
	api := httpAPI{app: appLayer}
	api.setupRoutesAndStart()
}

func (api httpAPI) setupRoutesAndStart() {
	http.Handle("/update/", http.HandlerFunc(api.update))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))

}
