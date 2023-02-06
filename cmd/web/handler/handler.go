package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sabertoot/server/internal/activitypub"
	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/data"
	"github.com/sabertoot/server/internal/plog"
)

const (
	mediaTypeJSON     = "application/json; charset=utf-8"
	mediaTypeActivity = "application/activity+json"
	mediaTypeJRD      = "application/jrd+json"
	mediaTypeHTML     = "text/html"
	mediaTypePNG      = "image/png"
	mediaTypeJPEG     = "image/jpeg"

	pageSize = 20
)

type Handler struct {
	settings    *config.Settings
	dataService *data.Service
	pubFactory  *activitypub.Factory
}

func New(
	settings *config.Settings,
	dataService *data.Service,
	pubFactory *activitypub.Factory,
) *Handler {
	return &Handler{
		settings:    settings,
		dataService: dataService,
		pubFactory:  pubFactory,
	}
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
	domainSuffix := "@" + h.settings.Server.Domain
	if !strings.HasSuffix(subject, domainSuffix) {
		h.error404(w, "User does not exist or has been moved")
		return
	}

	username := subject[:len(subject)-len(domainSuffix)]
	for _, user := range h.settings.Users {
		if user.Username == username {

			actorURL := h.settings.Server.PublicBaseURL + user.IDPath()
			profileURL := h.settings.Server.PublicBaseURL + user.ProfilePath()

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
	user *config.User,
) {
	if r.Method != http.MethodGet {
		h.error405(w, r)
		return
	}

	h.serveObject(w, h.pubFactory.NewActor(user))
}

func (h *Handler) serveProfileImage(
	w http.ResponseWriter,
	r *http.Request,
	user *config.User,
) {
	if r.Method != http.MethodGet {
		h.error405(w, r)
		return
	}

	entries, err := os.ReadDir(h.settings.Storage.ProfileImageDirectory())
	if err != nil {
		plog.Errorf("error reading profile image directory: %v", err)
		h.error500(w, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		ext := filepath.Ext(entry.Name())
		fileID := fileName[:len(fileName)-len(ext)]

		if fileID == user.ID.String() {
			file, err := os.Open(filepath.Join(h.settings.Storage.ProfileImageDirectory(), fileName))
			if err != nil {
				plog.Errorf("Error opening profile image file: %v", err)
				h.error500(w, err)
				return
			}
			defer file.Close()

			if ext == ".png" {
				w.Header().Set("Content-Type", mediaTypePNG)
			} else if ext == ".jpg" || ext == ".jpeg" {
				w.Header().Set("Content-Type", mediaTypeJPEG)
			}
			w.WriteHeader(http.StatusOK)
			io.Copy(w, file)
			return
		}
	}

	h.error404(w, "Profile image not found")
}

func (h *Handler) serveOutbox(
	w http.ResponseWriter,
	r *http.Request,
	user *config.User,
) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		h.error405(w, r)
		return
	}

	query := r.URL.Query()
	before := query.Get("before")
	after := query.Get("after")

	if len(before) > 0 && len(after) > 0 {
		h.error400(w, "Only the 'before' or 'after' query parameter may be set, not both")
		return
	}

	if len(before) > 0 {
		// Query toots before the given timestamp
		return
	}

	ctx := r.Context()

	if len(after) > 0 {
		after, err := strconv.ParseInt(after, 10, 64)
		if err != nil {
			h.error400(w, "The query parameter 'after' must be a valid int64 value")
			return
		}

		_, err = h.dataService.Toots(ctx, user.ID, after, pageSize)
		if err != nil {
			plog.Errorf("error getting toots: %v", err)
			h.error500(w, err)
			return
		}

		// for _, toot := range toots {
		// 	obj := activitypub.Object
		// }

		return
	}

	totalItems, err := h.dataService.TootCount(ctx, user.ID)
	if err != nil {
		plog.Errorf("error getting toot count: %v", err)
		h.error500(w, err)
		return
	}

	// Let's create a Y99k problem in memory of Jay-Z
	y9999 := time.Date(99000, 12, 31, 23, 59, 59, 0, time.UTC)
	maxEpoch := y9999.Unix()

	id := h.settings.Server.PublicBaseURL + user.OutboxPath()
	first := fmt.Sprintf("%s?after=0", id)
	last := fmt.Sprintf("%s?before=%d", id, maxEpoch)

	orderedCollection := activitypub.NewOrderedCollection(
		id,
		totalItems,
		first,
		last,
	)

	h.serveObject(w, orderedCollection)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/.well-known/webfinger" {
		h.serveWebFinger(w, r)
		return
	}

	for _, user := range h.settings.Users {

		if r.URL.Path == user.IDPath() {
			h.serveActor(w, r, user)
			return
		}

		if r.URL.Path == user.OutboxPath() {
			h.serveOutbox(w, r, user)
			return
		}

		if r.URL.Path == user.ProfileImagePath() {
			h.serveProfileImage(w, r, user)
			return
		}
	}

	h.error404Generic(w)
}
