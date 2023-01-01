package activitypub

import (
	"fmt"
)

type Username string

func (u Username) String() string {
	return string(u)
}

func (u Username) IDPath() string {
	return "/users/" + string(u)
}

func (u Username) ProfilePath() string {
	return "/@" + string(u)
}

func (u Username) InboxPath() string {
	return fmt.Sprintf("%s/inbox", u.IDPath())
}

func (u Username) OutboxPath() string {
	return fmt.Sprintf("%s/outbox", u.IDPath())
}

func (u Username) FollowersPath() string {
	return fmt.Sprintf("%s/followers", u.IDPath())
}

func (u Username) FollowingPath() string {
	return fmt.Sprintf("%s/following", u.IDPath())
}

func (u Username) LikedPath() string {
	return fmt.Sprintf("%s/liked", u.IDPath())
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

func NewActor(
	username Username,
	name string,
	summary string,
	iconURL string,
	publicBaseURL string) *Actor {
	return &Actor{
		Context:           "https://www.w3.org/ns/activitystreams",
		Type:              "Person",
		ID:                publicBaseURL + username.IDPath(),
		PreferredUsername: username.String(),
		Name:              name,
		Summary:           summary,
		Icon:              iconURL,
		Inbox:             publicBaseURL + username.InboxPath(),
		Outbox:            publicBaseURL + username.OutboxPath(),
		Followers:         publicBaseURL + username.FollowersPath(),
		Following:         publicBaseURL + username.FollowingPath(),
		Liked:             publicBaseURL + username.LikedPath(),
		URL:               publicBaseURL + username.ProfilePath(),
	}
}
