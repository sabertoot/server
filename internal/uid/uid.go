package uid

import (
	"fmt"
	"strconv"
)

// Identifier to denote where an object came from.
type SourceType uint8

func (s SourceType) Int() int {
	return int(s)
}

const (
	Twitter SourceType = iota

	// Add more here
	// Instagram
	// Facebook
	// TikTok
	// RSSFeed
	// etc.
)

// Unique identifier for a user.
// Assumption is that there won't be more than 256 users
// configured in settings.json as it kind of goes against
// the idea of this project.
type UserID uint8

func (id UserID) Int() int {
	return int(id)
}

func (id UserID) String() string {
	return strconv.Itoa(id.Int())
}

// Deterministic global unique identifier for a toot.
// It will be a combination of user ID, source type and source ID.
type TootID string

func (id TootID) String() string {
	return string(id)
}

// New creates a new UID.
func New(
	userID UserID,
	sourceType SourceType,
	sourceID uint64,
) TootID {
	// Format:
	// First two characters: Base36 encoded user ID
	// Next two characters: Base36 encoded source type
	// Rest of the characters: Base36 encoded source ID (foreign ID)
	return TootID(fmt.Sprintf(
		"%02s%02s%s",
		strconv.FormatUint(uint64(userID), 36),
		strconv.FormatUint(uint64(sourceType), 36),
		strconv.FormatUint(sourceID, 36)))
}
