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

	ErrInvalidParam     = errors.New("btcgw::invalid_param")
	ErrInvalidParamDesc = "Parameter should be a 32 bytes binary in hexadecimal string."

	ErrDigestNotFound     = errors.New("btcgw::digest_not_found")
	ErrDigestNotFoundDesc = "Digest not found."

	ErrRegisterFailed     = errors.New("btcgw::register_failed")
	ErrRegisterFailedDesc = "Could not register. There may be a system error."

	ErrDigestAlreadyExists     = errors.New("btcgw::digest_already_exists")
	ErrDigestAlreadyExistsDesc = "Digest already exists."

	ErrAPIKeyCreationFailed     = errors.New("btcgw::apikey_creation_failed")
	ErrAPIKeyCreationFailedDesc = "Could not create API Key. There may be a system error."

	ErrAPIKeyDeletionFailed     = errors.New("btcgw::apikey_deletion_failed")
	ErrAPIKeyDeletionFailedDesc = "Could not delete API Key. There may be a system error."

	ErrCouldNotClose = errors.New("ErrCouldNotClose")
)
