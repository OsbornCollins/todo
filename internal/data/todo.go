// Filename: internal/data/todo.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"todo.osborncollins.net/internal/validator"
)

type Todo struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Task_Name   string    `json:"task_name"`
	Description string    `json:"desription"`
	Notes       string    `json:"notes"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Status      []string  `json:"status"`
	Version     int32     `json:"version"`
}

func ValidateTodo(v *validator.Validator, todo *Todo) {

	// Use the check() method to execute our validation checks
	// Task_Name validation
	v.Check(todo.Task_Name != "", "task_name", "must be provided")
	v.Check(len(todo.Task_Name) <= 300, "task_name", "must not be more than 300 bytes long")

	// Description Validation
	v.Check(todo.Description != "", "description", "must be provided")
	v.Check(len(todo.Description) <= 800, "level", "must not be more than 800 bytes long")

	// Notes validation
	v.Check(todo.Notes != "", "notes", "must be provided")
	v.Check(len(todo.Notes) <= 500, "notes", "must not be more than 500 bytes long")

	// Category validation
	v.Check(todo.Category != "", "category", "must be provided")
	v.Check(len(todo.Category) <= 200, "category", "must not be more than 200 bytes long")

	// Priority validation
	v.Check(todo.Priority != "", "priority", "must be provided")
	v.Check(len(todo.Priority) <= 100, "priority", "must not be more than 100 bytes long")

	//Staus validation
	v.Check(todo.Status != nil, "status", "must be provided")
	v.Check(len(todo.Status) >= 1, "status", "must contain atleast 1 entry")
	v.Check(len(todo.Status) <= 5, "status", "must contain less than 6 entries")
	v.Check(validator.Unique(todo.Status), "status", "must not contain duplicate entries")
}

// Define a TodoModel which wraps a sql.DB connection pool
type TodoModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new todo item
func (m TodoModel) Insert(todo *Todo) error {
	query := `
	INSERT INTO todotbl (task_name, description, notes, category, priority, status)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Collect the data fields into a slice
	args := []interface{}{todo.Task_Name, todo.Description, todo.Notes,
		todo.Category, todo.Priority, pq.Array(todo.Status),
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&todo.ID, &todo.CreatedAt, &todo.Version)
}

// GET() allows us to retrieve a specific todo item
func (m TodoModel) Get(id int64) (*Todo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create query
	query := `
		SELECT id, created_at, task_name, description, notes, category, priority, status, version
		FROM todotbl
		WHERE id = $1
	`
	// Declare a Todo variable to hold the return data
	var todo Todo
	// Execute Query using the QueryRow
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.CreatedAt,
		&todo.Task_Name,
		&todo.Description,
		&todo.Notes,
		&todo.Category,
		&todo.Priority,
		pq.Array(&todo.Status),
		&todo.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &todo, nil
}

// Update() allows us to edit/alter a todo item in the list
func (m TodoModel) Update(todo *Todo) error {
	query := `
		UPDATE todotbl 
		set task_name = $1, description = $2, 
		notes = $3, category = $4, 
		priority = $5, status = $6, 
		version = version + 1
		WHERE id = $7
		AND version = $8
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	args := []interface{}{
		todo.Task_Name,
		todo.Description,
		todo.Notes,
		todo.Category,
		todo.Priority,
		pq.Array(todo.Status),
		todo.ID,
		todo.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&todo.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific todo item from the list
func (m TodoModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM todotbl
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	results, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operations. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// The GetAll() returns a list of all the todo items sorted by ID
func (m TodoModel) GetAll(task_name string, priority string, status []string, filters Filters) ([]*Todo, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, task_name, description, notes, category, priority, status, version
		FROM todotbl
		WHERE (to_tsvector('simple',task_name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple',priority) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (status @> $3 OR $3 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-second-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{task_name, priority, pq.Array(status), filters.limit(), filters.offset()}
	// Execute query
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the result set
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the todo data
	todos := []*Todo{}
	// Iterate over the rows in the results set
	for rows.Next() {
		var todo Todo
		// Scan the values from the row in to the Todo struct
		err := rows.Scan(
			&totalRecords,
			&todo.ID,
			&todo.CreatedAt,
			&todo.Task_Name,
			&todo.Description,
			&todo.Notes,
			&todo.Category,
			&todo.Priority,
			pq.Array(&todo.Status),
			&todo.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the Todo to our slice
		todos = append(todos, &todo)
	}
	// Check for errors after looping through the results set
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of Todos
	return todos, metadata, nil
}
