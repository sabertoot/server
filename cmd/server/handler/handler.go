package handler

import (
	"net/http"

	"github.com/sabertoot/server/internal/config"
)

type Handler struct {
	Settings *config.Settings
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World!"))
}
