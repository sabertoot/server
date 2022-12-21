package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sabertoot/server/internal/activitypub"
	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/plog"
)

const (
	mediaTypeJSON     = "application/json; charset=utf-8"
	mediaTypeActivity = "application/activity+json"
	mediaTypeJRD      = "application/jrd+json"
)

type Handler struct {
	Settings *config.Settings
}

func clearHeaders(w http.ResponseWriter) {
	for k := range w.Header() {
		w.Header().Del(k)
	}
}

func (h *Handler) error404(w http.ResponseWriter) {
	clearHeaders(w)
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", mediaTypeJSON)
	w.Write([]byte(`{ "error": "Resource does not exist or has been moved" }`))
}

func (h *Handler) error405(w http.ResponseWriter, r *http.Request) {
	clearHeaders(w)
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Header().Set("Content-Type", mediaTypeJSON)
	w.Write([]byte(
		fmt.Sprintf(
			`{ "error": "This endpoint does not allow HTTP %s requests" }`,
			r.Method)))
}

func (h *Handler) error500(w http.ResponseWriter, err error) {
	clearHeaders(w)
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", mediaTypeJSON)
	w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err.Error())))
}

func (h *Handler) serveWebFinger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.error405(w, r)
		return
	}

	resource := r.URL.Query().Get("resource")
	acctPrefix := "acct:"
	if resource == "" || !strings.HasPrefix(resource, acctPrefix) {
		h.error404(w)
		return
	}

	subject := resource[len(acctPrefix):]
	hostSuffix := "@" + h.Settings.Server.PublicHost
	if !strings.HasSuffix(subject, hostSuffix) {
		h.error404(w)
		return
	}

	userID := activitypub.ID(subject[:len(subject)-len(hostSuffix)])
	for id, _ := range h.Settings.Accounts {
		if id == userID.String() {
			w.Header().Set("Content-Type", mediaTypeJRD)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(
				fmt.Sprintf(
					`{ "subject": "%s", "links": [ { "rel": "self", "type": "%s", "href": "%s" } ] }`,
					resource,
					mediaTypeActivity,
					h.Settings.Server.PublicBaseURL+userID.IDPath())))
			return
		}
	}

	h.error404(w)
}

func (h *Handler) serveObject(w http.ResponseWriter, obj any) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		plog.Errorf("error marshalling actor: %v", err)
		h.error500(w, err)
		return
	}

	w.Header().Set("Content-Type", mediaTypeActivity)
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (h *Handler) serveActor(
	w http.ResponseWriter,
	r *http.Request,
	userID activitypub.ID,
) {
	if r.Method != http.MethodGet {
		h.error405(w, r)
		return
	}

	accountDetails := h.Settings.Accounts[userID.String()]
	h.serveObject(w, activitypub.NewActor(
		userID,
		accountDetails.Name,
		accountDetails.Summary,
		"ToDo",
		h.Settings.Server.PublicBaseURL))
}

func (h *Handler) serveInbox(
	w http.ResponseWriter,
	r *http.Request,
	userID activitypub.ID,
) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		h.error405(w, r)
		return
	}
}

func (h *Handler) serveOutbox(
	w http.ResponseWriter,
	r *http.Request,
	userID activitypub.ID,
) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		h.error405(w, r)
		return
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/.well-known/webfinger" {
		h.serveWebFinger(w, r)
		return
	}

	for id := range h.Settings.Accounts {
		userID := activitypub.ID(id)

		if r.URL.Path == userID.IDPath() {
			h.serveActor(w, r, userID)
			return
		}

		if r.URL.Path == userID.InboxPath() {
			h.serveInbox(w, r, userID)
			return
		}

		if r.URL.Path == userID.OutboxPath() {
			h.serveOutbox(w, r, userID)
			return
		}
	}

	h.error404(w)
}
