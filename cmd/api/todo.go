// Filename: cmd/api/todo.go

package main

import (
	"errors"
	"fmt"
	"net/http"

	"todo.osborncollins.net/internal/data"
	"todo.osborncollins.net/internal/validator"
)

// createTODOItemHandler for the "POST" /v1/todoitems" endpoint
func (app *application) createTODOItemHandler(w http.ResponseWriter, r *http.Request) {
	// Our Target decode destination
	var input struct {
		Task_Name   string   `json:"task_name"`
		Description string   `json:"description"`
		Notes       string   `json:"notes"`
		Category    string   `json:"category"`
		Priority    string   `json:"priority"`
		Status      []string `json:"status"`
	}
	// Initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//Copy the values from the input struct to a new todo struct
	todo := &data.Todo{
		Task_Name:   input.Task_Name,
		Description: input.Description,
		Notes:       input.Notes,
		Category:    input.Category,
		Priority:    input.Priority,
		Status:      input.Status,
	}
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateTodo(v, todo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create a Todo Object
	err = app.models.Todos.Insert(todo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// Create a location header for the newly created resource/Todo object
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/todoitems/%d", todo.ID))
	// Write the JSON response with 201 - created status code with the body
	// being the actual todo data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"todo": todo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showTODOItemHandlerfor the "GET" /v1/todoitems/:id" endpoint
func (app *application) showTODOItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the specific todo item
	todo, err := app.models.Todos.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the response by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"todo": todo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateTODOItemHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the todo item that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the original record from the database
	todo, err := app.models.Todos.Get(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold data read in from the client
	// We update the input struct to use pointers because pointers have a
	// default value of nil false
	// if a field remains nil then we know that the client did not update it
	var input struct {
		Task_Name   *string  `json:"task_name"`
		Description *string  `json:"description"`
		Notes       *string  `json:"notes"`
		Category    *string  `json:"category"`
		Priority    *string  `json:"priority"`
		Status      []string `json:"status"`
	}

	//Initalize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if input.Task_Name != nil {
		todo.Task_Name = *input.Task_Name
	}
	if input.Description != nil {
		todo.Description = *input.Description
	}
	if input.Notes != nil {
		todo.Notes = *input.Notes
	}
	if input.Category != nil {
		todo.Category = *input.Category
	}
	if input.Priority != nil {
		todo.Priority = *input.Priority
	}
	if input.Status != nil {
		todo.Status = input.Status
	}

	// Perform Validation on the updated todo item. If validation fails then
	// we send a 422 - unprocessable entity response to the client
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateTodo(v, todo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the update todo record to the Update() method
	err = app.models.Todos.Update(todo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"todo": todo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// The deleteTODOItemHandler() allows the user to delete a todo item from the databse by using the ID
func (app *application) deleteTODOItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the todo item from the database. Send a 404 Not Found status code to the
	// client if there is no matching record
	err = app.models.Todos.Delete(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return 200 Status OK to the client with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "todo item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// The listTODOItemsHandler() allows the client to see a listing of todo items
// based on a set criteria
func (app *application) listTODOItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameter
	var input struct {
		Task_Name string
		Priority  string
		Status    []string
		data.Filters
	}
	// Initialize a validator
	v := validator.New()
	// Get the URL values map
	qs := r.URL.Query()
	// use the helper methods to extract values
	input.Task_Name = app.readString(qs, "task_name", "")
	input.Priority = app.readString(qs, "priority", "")
	input.Status = app.readCSV(qs, "status", []string{})
	// Get the page information using the read int method
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Specify the allowed sort values
	input.Filters.SortList = []string{"id", "task_name", "priority", "-id", "-task_name", "-priority"}
	// Check for validation errors
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all todo items
	todos, metadata, err := app.models.Todos.GetAll(input.Task_Name, input.Priority, input.Status, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing all the todo items
	err = app.writeJSON(w, http.StatusOK, envelope{"todos": todos, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
