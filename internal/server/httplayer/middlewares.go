package httplayer

import (
	"fmt"
	"net/http"
)

func setHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Content-Type", "text/plain")
		next.ServeHTTP(w, r)
	})
}
