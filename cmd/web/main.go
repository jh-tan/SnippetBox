package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/jh-tan/snippetbox/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "postgres://web:pass@/snippetbox", "Postgres data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	context := context.Background()
	db, err := openDB(*dsn, context)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &models.SnippetModel{DB: db, CONTEXT: context},
		templateCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string, context context.Context) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context, dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(context); err != nil {
		return nil, err
	}

	return db, nil
}
