package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set(HTTPHeaderContentType, ContentTypeApplicationJson)
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusOK, newAPIResp(true, "", data))
}

func writeUnauthorized(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusUnauthorized, newAPIResp(false, "unauthorized", data))
}

func writeBadRequest(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusBadRequest, newAPIResp(false, "bad request", data))
}

func writeInternalServerError(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusInternalServerError, newAPIResp(false, "internal server error", data))
}
