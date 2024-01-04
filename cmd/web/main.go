package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joshuagageellis/snippetbox.git/internal/models"
)

type Env struct {
	PORT string
	HOST string
	DSN  string
	ENV  string
}

// Application dependencies.
type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	env            *Env
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	// Load env.
	loadEnv(app)

	// Init DB pool.
	db, err := openDB(app.env.DSN)
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

	// Init form decoder.
	app.formDecoder = form.NewDecoder()

	// Seed database.
	if app.env.ENV == "dev" {
		err = app.snippets.SeedDatabase()
		if err != nil {
			app.logger.Error(err.Error())
			os.Exit(1)
		}
	}

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	app.sessionManager = sessionManager

	// Init template cache.
	app.templateCache, err = newTemplateCache()
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}

	app.logger.Info(fmt.Sprintf("Start on %s:%s", app.env.HOST, app.env.PORT))

	errServer := http.ListenAndServe(fmt.Sprintf("%s:%s", app.env.HOST, app.env.PORT), app.routes())
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
