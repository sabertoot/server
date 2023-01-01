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
	mediaTypeHTML     = "text/html"
)

type Handler struct {
	Settings *config.Settings
}

func clearHeaders(w http.ResponseWriter) {
	for k := range w.Header() {
		w.Header().Del(k)
	}
}

func (h *Handler) error404(w http.ResponseWriter, msg string) {
	clearHeaders(w)
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", mediaTypeJSON)
	w.Write([]byte(`{ "error": "` + msg + `" }`))
}

func (h *Handler) error404Generic(w http.ResponseWriter) {
	h.error404(w, "Resource does not exist or has been moved")
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

func (h *Handler) error400(w http.ResponseWriter, msg string) {
	clearHeaders(w)
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", mediaTypeJSON)
	w.Write([]byte(`{ "error": "` + msg + `" }`))
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
		h.error400(w, "Invalid resource query parameter")
		return
	}

	subject := resource[len(acctPrefix):]
	domainSuffix := "@" + h.Settings.Server.Domain
	if !strings.HasSuffix(subject, domainSuffix) {
		h.error404(w, "User does not exist or has been moved")
		return
	}

	username := activitypub.Username(subject[:len(subject)-len(domainSuffix)])
	for _, user := range h.Settings.Users {
		if user.Username == username.String() {

			actorURL := h.Settings.Server.PublicBaseURL + username.IDPath()
			profileURL := h.Settings.Server.PublicBaseURL + username.ProfilePath()

			w.Header().Set("Content-Type", mediaTypeJRD)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(
				fmt.Sprintf(
					`{ "subject": "%s", "aliases": [ "%s", "%s" ], "links": [ { "rel": "self", "type": "%s", "href": "%s" }, { "rel": "http://webfinger.net/rel/profile-page", "type": "%s", "href": "%s" } ] }`,
					resource,
					actorURL,
					profileURL,
					mediaTypeActivity,
					actorURL,
					mediaTypeHTML,
					profileURL)))
			return
		}
	}

	h.error404(w, "User does not exist or has been moved")
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
	username activitypub.Username,
	user config.User,
) {
	if r.Method != http.MethodGet {
		h.error405(w, r)
		return
	}

	h.serveObject(w, activitypub.NewActor(
		username,
		user.FullName,
		user.Summary,
		"ToDo",
		h.Settings.Server.PublicBaseURL))
}

func (h *Handler) serveOutbox(
	w http.ResponseWriter,
	r *http.Request,
	user config.User,
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

	for _, user := range h.Settings.Users {
		username := activitypub.Username(user.Username)

		if r.URL.Path == username.IDPath() {
			h.serveActor(w, r, username, user)
			return
		}

		if r.URL.Path == username.OutboxPath() {
			h.serveOutbox(w, r, user)
			return
		}
	}

	h.error404Generic(w)
}
