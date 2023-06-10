package httplayer

import (
	"fmt"
	"net/http"
	"strings"
)

func (api *httpAPI) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauge, counter := api.app.GetAll()
	var s []string
	for key, value := range gauge {
		//to do: create a function
		metricValueWithoutTrailingZero := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", value), "0"), ".")
		s = append(s, fmt.Sprintf("<h3>%s: %s</h3>\n", key, metricValueWithoutTrailingZero))
	}
	for key, value := range counter {
		s = append(s, fmt.Sprintf("<h3>%s: %d</h3>\n", key, value))
	}
	for _, i := range s {
		w.Write([]byte(i))
	}
}
