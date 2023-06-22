// Package oapigen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.13.0 DO NOT EDIT.
package oapigen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/go-chi/chi/v5"
)

// Defines values for SuggestJSONBodySort.
const (
	SuggestJSONBodySortAsc  SuggestJSONBodySort = "asc"
	SuggestJSONBodySortDesc SuggestJSONBodySort = "desc"
)

// Error default error
type Error struct {
	Message string `json:"message"`
}

// SuggestJSONBody defines parameters for Suggest.
type SuggestJSONBody struct {
	Limit  *int                `json:"limit,omitempty"`
	Prefix string              `json:"prefix"`
	Sort   SuggestJSONBodySort `json:"sort"`
}

// SuggestJSONBodySort defines parameters for Suggest.
type SuggestJSONBodySort string

// TranslateJSONBody defines parameters for Translate.
type TranslateJSONBody struct {
	Text string `json:"text"`
}

// SuggestJSONRequestBody defines body for Suggest for application/json ContentType.
type SuggestJSONRequestBody SuggestJSONBody

// TranslateJSONRequestBody defines body for Translate for application/json ContentType.
type TranslateJSONRequestBody TranslateJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /emoji/suggest)
	Suggest(w http.ResponseWriter, r *http.Request)

	// (POST /emoji/translate)
	Translate(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// Suggest operation middleware
func (siw *ServerInterfaceWrapper) Suggest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Suggest(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// Translate operation middleware
func (siw *ServerInterfaceWrapper) Translate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Translate(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshallingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshallingParamError) Error() string {
	return fmt.Sprintf("Error unmarshalling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshallingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/emoji/suggest", wrapper.Suggest)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/emoji/translate", wrapper.Translate)
	})

	return r
}

type SuggestRequestObject struct {
	Body *SuggestJSONRequestBody
}

type SuggestResponseObject interface {
	VisitSuggestResponse(w http.ResponseWriter) error
}

type Suggest200JSONResponse []struct {
	Alias string `json:"alias"`
	Char  string `json:"char"`
}

func (response Suggest200JSONResponse) VisitSuggestResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type SuggestdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response SuggestdefaultJSONResponse) VisitSuggestResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type TranslateRequestObject struct {
	Body *TranslateJSONRequestBody
}

type TranslateResponseObject interface {
	VisitTranslateResponse(w http.ResponseWriter) error
}

type Translate200JSONResponse string

func (response Translate200JSONResponse) VisitTranslateResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type TranslatedefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response TranslatedefaultJSONResponse) VisitTranslateResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (POST /emoji/suggest)
	Suggest(ctx context.Context, request SuggestRequestObject) (SuggestResponseObject, error)

	// (POST /emoji/translate)
	Translate(ctx context.Context, request TranslateRequestObject) (TranslateResponseObject, error)
}

type StrictHandlerFunc = runtime.StrictHttpHandlerFunc
type StrictMiddlewareFunc = runtime.StrictHttpMiddlewareFunc

type StrictHTTPServerOptions struct {
	RequestErrorHandlerFunc  func(w http.ResponseWriter, r *http.Request, err error)
	ResponseErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}}
}

func NewStrictHandlerWithOptions(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc, options StrictHTTPServerOptions) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: options}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
	options     StrictHTTPServerOptions
}

// Suggest operation middleware
func (sh *strictHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	var request SuggestRequestObject

	var body SuggestJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.Suggest(ctx, request.(SuggestRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "Suggest")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(SuggestResponseObject); ok {
		if err := validResponse.VisitSuggestResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// Translate operation middleware
func (sh *strictHandler) Translate(w http.ResponseWriter, r *http.Request) {
	var request TranslateRequestObject

	var body TranslateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.Translate(ctx, request.(TranslateRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "Translate")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(TranslateResponseObject); ok {
		if err := validResponse.VisitTranslateResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}