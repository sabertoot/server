package activitypub

import (
	"fmt"
)

type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) IDPath() string {
	return "/" + string(id)
}

func (id ID) InboxPath() string {
	return fmt.Sprintf("%s/inbox", id.IDPath())
}

func (id ID) OutboxPath() string {
	return fmt.Sprintf("%s/outbox", id.IDPath())
}

func (id ID) FollowersPath() string {
	return fmt.Sprintf("%s/followers", id.IDPath())
}

func (id ID) FollowingPath() string {
	return fmt.Sprintf("%s/following", id.IDPath())
}

func (id ID) LikedPath() string {
	return fmt.Sprintf("%s/liked", id.IDPath())
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
}

func NewActor(
	id ID,
	name string,
	summary string,
	iconURL string,
	publicBaseURL string) *Actor {
	return &Actor{
		Context:           "https://www.w3.org/ns/activitystreams",
		Type:              "Person",
		ID:                publicBaseURL + id.IDPath(),
		PreferredUsername: id.String(),
		Name:              name,
		Summary:           summary,
		Icon:              iconURL,
		Inbox:             publicBaseURL + id.InboxPath(),
		Outbox:            publicBaseURL + id.OutboxPath(),
		Followers:         publicBaseURL + id.FollowersPath(),
		Following:         publicBaseURL + id.FollowingPath(),
		Liked:             publicBaseURL + id.LikedPath(),
	}
}
