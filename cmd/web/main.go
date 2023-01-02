package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sabertoot/server/cmd/web/handler"
	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/plog"
)

func main() {
	plog.Info("Sabertoot server starting...")

	plog.Debug("Loading settings...")
	settings, err := config.Load()
	if err != nil {
		plog.Fatal(err.Error())
		return
	}

	plog.Debug("Initialising handler...")
	webHandler := handler.New(settings)

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
