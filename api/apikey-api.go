//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package apikey -generate "types,chi-server,spec" -include-tags "API Key" -o apikey/api.gen.go openapi.yml

package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ebiiim/btcgw/api/apikey"
	"github.com/ebiiim/btcgw/auth"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
)

// export

var APIKeyHandlerFromMux = apikey.HandlerFromMux

type APIKeyService struct {
	d *auth.DocstoreAuth
}

var _ apikey.ServerInterface = (*APIKeyService)(nil)

func NewAPIKeyService(d *auth.DocstoreAuth) *APIKeyService {
	a := &APIKeyService{
		d: d,
	}
	return a
}

func sendAPIKeyServiceError(w http.ResponseWriter, code int, err error, desc string) {
	var pDesc *string = nil
	if desc != "" {
		pDesc = &desc
	}
	gsErr := apikey.Error{
		Error:            err.Error(),
		ErrorDescription: pDesc,
	}
	WriteJSON(w, code, gsErr)
}

func (a *APIKeyService) PostApikeysCreate(w http.ResponseWriter, r *http.Request) {
	var rdom apikey.BBc1Domain
	if err := json.NewDecoder(r.Body).Decode(&rdom); err != nil {
		sendAPIKeyServiceError(w, http.StatusBadRequest, ErrInvalidRequestBody, ErrInvalidRequestBodyDesc)
		return
	}
	if _, err := hex.DecodeString(rdom.Domain); err != nil {
		sendAPIKeyServiceError(w, http.StatusBadRequest, ErrInvalidParam, ErrInvalidParamDesc)
		return
	}
	ctx := r.Context()
	k, err := a.d.Generate(ctx, rdom.Domain, false, fmt.Sprintf("Created by API at: %s", time.Now().Format(time.RFC3339)))
	if err != nil {
		sendAPIKeyServiceError(w, http.StatusInternalServerError, ErrAPIKeyCreationFailed, ErrAPIKeyCreationFailedDesc)
		return
	}
	WriteJSON(w, http.StatusOK, &apikey.APIKey{Key: k.Key})
}

func (a *APIKeyService) PostApikeysDelete(w http.ResponseWriter, r *http.Request) {
	var rkey apikey.APIKey
	if err := json.NewDecoder(r.Body).Decode(&rkey); err != nil {
		sendAPIKeyServiceError(w, http.StatusBadRequest, ErrInvalidRequestBody, ErrInvalidRequestBodyDesc)
		return
	}
	ctx := r.Context()
	if err := a.d.Delete(ctx, rkey.Key); err != nil {
		sendAPIKeyServiceError(w, http.StatusInternalServerError, ErrAPIKeyDeletionFailed, ErrAPIKeyDeletionFailedDesc)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// OAPIValidator sets up OpenAPI validator and must be set as a middleware.
func (a *APIKeyService) OAPIValidator() func(next http.Handler) http.Handler {
	swagger, err := apikey.GetSwagger()
	if err != nil {
		panic(fmt.Sprintf("could not load swagger spec: %s", err))
	}
	// Skips validating server names.
	swagger.Servers = nil

	validatorOpts := &oapimiddleware.Options{}
	return oapimiddleware.OapiRequestValidatorWithOptions(swagger, validatorOpts)
}

func (a *APIKeyService) Close() error {
	if err := a.d.Close(); err != nil {
		return fmt.Errorf("%w (DocstoreAuth: %v)", ErrCouldNotClose, err)
	}
	return nil
}
