package httplayer

import (
	"github.com/nktau/monitoring-service/internal/applayer"
	"log"
	"net/http"
)

type httpApi struct {
	app applayer.App
}

func New(appLayer applayer.App) {
	api := httpApi{app: appLayer}
	api.setupRoutesAndStart()
}

func (self httpApi) setupRoutesAndStart() {
	http.Handle("/update/", http.HandlerFunc(self.update))
	log.Fatal(http.ListenAndServe("localhost:8080", nil))

}
