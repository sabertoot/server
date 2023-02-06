package main

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/sabertoot/server/cmd/web/handler"
	"github.com/sabertoot/server/internal/activitypub"
	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/data"
	"github.com/sabertoot/server/internal/plog"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()
	plog.Info("Sabertoot server starting...")

	plog.Debug("Loading settings...")
	settings, err := config.Load()
	if err != nil {
		plog.Fatal(err.Error())
		return
	}

	plog.Infof("Initialising SQLite database: %s", settings.SQLite.DSN)
	db, err := sql.Open("sqlite3", settings.SQLite.DSN)
	if err != nil {
		plog.Fatal(err.Error())
		return
	}
	defer db.Close()

	var version string
	err = db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)
	if err != nil {
		plog.Fatal(err.Error())
		return
	}
	plog.Debugf("SQLite version: %s", version)

	dataService := data.NewService(db)
	if err = dataService.InitTables(ctx); err != nil {
		plog.Fatal(err.Error())
		return
	}

	plog.Debug("Initialising handler...")
	pubFactory := activitypub.NewFactory(settings.Server.PublicBaseURL)
	webHandler := handler.New(settings, dataService, pubFactory)

	plog.Debug("Creating server...")
	httpServer := &http.Server{
		Addr:              ":" + strconv.Itoa(settings.Server.Port),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       620 * time.Second,
		Handler:           webHandler,
		MaxHeaderBytes:    settings.Server.MaxHeaderBytes,
	}
	plog.Infof("Listening on port %d", settings.Server.Port)
	err = httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
