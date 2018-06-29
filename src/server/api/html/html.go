package html

import (
	"net/http"
)

func ShowLoginTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "./src/server/api/html/test/ws_login.html")
}
