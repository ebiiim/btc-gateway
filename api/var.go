package api

import (
	"encoding/json"
	"errors"
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

var (
	ErrInvalidRequestBody     = errors.New("btcgw:invalid_request_body")
	ErrInvalidRequestBodyDesc = "Request body should be a JSON."

	ErrInvalidID     = errors.New("btcgw::invalid_id")
	ErrInvalidIDDesc = "ID should be a 32 bytes binary in hexadecimal string."

	ErrTxNotFound     = errors.New("btcgw::tx_not_found")
	ErrTxNotFoundDesc = "Transaction not found."

	ErrRegisterFailed     = errors.New("btcgw::register_failed")
	ErrRegisterFailedDesc = "Could not register. There may be a system error."

	ErrTxAlreadyExists     = errors.New("btcgw::tx_already_exists")
	ErrTxAlreadyExistsDesc = "Transaction already exists."

	ErrAPIKeyCreationFailed     = errors.New("btcgw::apikey_creation_failed")
	ErrAPIKeyCreationFailedDesc = "Could not create API Key. There may be a system error."

	ErrAPIKeyDeletionFailed     = errors.New("btcgw::apikey_deletion_failed")
	ErrAPIKeyDeletionFailedDesc = "Could not delete API Key. There may be a system error."

	ErrCouldNotClose = errors.New("ErrCouldNotClose")
)
