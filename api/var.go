package api

import (
	"encoding/json"
	"net/http"
)

var PrettifyResponseJSON = false

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	e := json.NewEncoder(w)
	if PrettifyResponseJSON {
		e.SetIndent("", "  ")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	e.Encode(v)
}
