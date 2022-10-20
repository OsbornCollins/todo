// Filename: cmd/api/routes.go

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// Create a new httprouter router instance
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/todoitems", app.listTODOItemsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/todoitems", app.createTODOItemHandler)
	router.HandlerFunc(http.MethodGet, "/v1/todoitems/:id", app.showTODOItemHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/todoitems/:id", app.updateTODOItemHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/todoitems/:id", app.deleteTODOItemHandler)

	return router
}
