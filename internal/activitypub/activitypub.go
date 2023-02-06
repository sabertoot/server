package activitypub

import (
	"time"

	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/uid"
)

const (
	activityStreamsContext = "https://www.w3.org/ns/activitystreams"
)

type Factory struct {
	publicBaseURL string
}

func NewFactory(publicBaseURL string) *Factory {
	return &Factory{
		publicBaseURL: publicBaseURL,
	}
}

type Actor struct {
	Context           string `json:"@context"`
	Type              string `json:"type"`
	ID                string `json:"id"`
	PreferredUsername string `json:"preferredUsername"`
	Name              string `json:"name"`
	Summary           string `json:"summary"`
	Icon              string `json:"icon"`
	Inbox             string `json:"inbox"`
	Outbox            string `json:"outbox"`
	Followers         string `json:"followers"`
	Following         string `json:"following"`
	Liked             string `json:"liked"`
	URL               string `json:"url"`
}

func (f *Factory) NewActor(
	user *config.User) *Actor {
	profileImageURL :=
		f.publicBaseURL +
			h.settings.Storage.ProfileImageRelativeURLPath(user.UserID())
	return &Actor{
		Context:           activityStreamsContext,
		Type:              "Person",
		ID:                f.publicBaseURL + user.IDPath(),
		PreferredUsername: user.Username,
		Name:              user.FullName,
		Summary:           user.Summary,
		Icon:              profileImageURL,
		Inbox:             f.publicBaseURL + user.InboxPath(),
		Outbox:            f.publicBaseURL + user.OutboxPath(),
		Followers:         f.publicBaseURL + user.FollowersPath(),
		Following:         f.publicBaseURL + user.FollowingPath(),
		Liked:             f.publicBaseURL + user.LikedPath(),
		URL:               f.publicBaseURL + user.ProfilePath(),
	}
}

type OrderedCollection struct {
	Context    string `json:"@context"`
	Type       string `json:"type"`
	ID         string `json:"id"`
	TotalItems int    `json:"totalItems"`
	First      string `json:"first"`
	Last       string `json:"last"`
}

func NewOrderedCollection(
	id string,
	totalItems int,
	first string,
	last string) *OrderedCollection {
	return &OrderedCollection{
		Context:    activityStreamsContext,
		Type:       "OrderedCollection",
		ID:         id,
		TotalItems: totalItems,
		First:      first,
		Last:       last,
	}
}

type Object struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Summary      string   `json:"summary,omitempty"`
	InReplyTo    string   `json:"inReplyTo,omitempty"`
	Published    string   `json:"published"`
	URL          string   `json:"url"`
	AttributedTo string   `json:"attributedTo"`
	To           []string `json:"to"`
	CC           []string `json:"cc"`
}

func NewNote(
	tootID uid.TootID,
	user *config.User,
	published time.Time,
) *Object {
	return &Object{
		Type:      "Note",
		Summary:   "",
		InReplyTo: "",
		Published: published.Format(time.RFC3339),
	}
}

type OrderedItem struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Actor     string   `json:"actor"`
	Published string   `json:"published"`
	To        []string `json:"to"`
	CC        []string `json:"cc"`
	Object    *Object  `json:"object"`
}

type OrderedCollectionPage struct {
	Context      string         `json:"@context"`
	Type         string         `json:"type"`
	ID           string         `json:"id"`
	Next         string         `json:"next"`
	Prev         string         `json:"prev"`
	PartOf       string         `json:"partOf"`
	OrderedItems []*OrderedItem `json:"orderedItems"`
}

func NewOrderedCollectionPage(
	id string,
	partOf string,
	next string,
	prev string) *OrderedCollectionPage {
	return &OrderedCollectionPage{
		Context:      activityStreamsContext,
		Type:         "OrderedCollectionPage",
		ID:           id,
		Next:         next,
		Prev:         prev,
		PartOf:       partOf,
		OrderedItems: []*OrderedItem{},
	}
}
