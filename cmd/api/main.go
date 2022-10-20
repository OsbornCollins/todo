// Filename: cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"todo.osborncollins.net/internal/data"
)

// The Application Version Number
const version = "1.0.0"

// The Configuration Settings
type config struct {
	port int
	env  string // Development, Staging, Production, ETC.
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Dependency Injection
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config

	// Read in flags that are needed to populate our config
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port") // When using a struct we must use IntVar
	flag.StringVar(&cfg.env, "env", "development", "Environment( Development | Staging | Production )")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("TODO_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	//Create a logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	// Create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	//Create an instance of our application struct
	// We are using the application struct for dependecy injection
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	// If anything happens we would like to close connection
	defer db.Close()
	//Log the sucessful connection pool
	logger.Println("Database connection pool established")

	// Create our new servemux
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Create HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start our Server
	logger.Printf("Starting %s Server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

}

//The openDB() function returns a pointer to an sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// Test the connection pool
	// Create a conteext with a 5 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
