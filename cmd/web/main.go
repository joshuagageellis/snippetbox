package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joshuagageellis/snippetbox.git/internal/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Application dependencies.
type application struct {
	logger   *slog.Logger
	snippets *models.SnippetModel
}

func main() {
	// Logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Init new application.
	app := &application{
		logger: logger,
	}

	enverr := godotenv.Load()
	if enverr != nil {
		app.logger.Error("Error loading .env file")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		app.logger.Error("PORT must be set in env file")
		os.Exit(1)
	}

	host := os.Getenv("HOST")
	if host == "" {
		app.logger.Error("HOST must be set in env file")
		os.Exit(1)
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		app.logger.Error("DSN must be set in env file")
		os.Exit(1)
	}
	// Init DB pool.
	db, err := openDB(dsn)
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	// Initialize a new instance of SnippetModel and add it to the application
	// dependencies.
	app.snippets = &models.SnippetModel{DB: db}

	// Init table.
	err = app.snippets.CreateSnippetTable()
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}

	// Init index.
	app.snippets.CreateSnippetIndex()
	if err != nil {
		app.logger.Warn(err.Error(), "msg", "index already exists")
	}

	app.logger.Info(fmt.Sprintf("Start on %s:%s", host, port))

	errServer := http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), app.routes())
	log.Fatal(errServer)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
