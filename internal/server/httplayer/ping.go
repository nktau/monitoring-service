package httplayer

import (
	"fmt"
	"net/http"
)

func (api *httpAPI) ping(w http.ResponseWriter, r *http.Request) {
	err := api.app.Ping()
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("ok\n"))
}
