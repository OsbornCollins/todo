// Filename: cmd/api/errors.go

package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

// We want to send JSON formatted error messages
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	// Create the JSON response
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Server error response
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// We log the error
	app.logError(r, err)
	// Prepare a message with the error
	message := "the server encountered a problem and could not process the request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// The not found response
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	// Create our message
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// Method not allowed response
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	// Create our message
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// User provided a bad request
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// Validation errors
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// Edit conflict errors
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}
