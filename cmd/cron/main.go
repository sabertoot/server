package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/data"
	"github.com/sabertoot/server/internal/download"
	"github.com/sabertoot/server/internal/plog"
	"github.com/sabertoot/server/internal/twitter"
	"github.com/sabertoot/server/internal/uid"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	plog.Info("Sabertoot cron job starting...")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	plog.Debug("Loading settings...")
	settings, err := config.Load()
	if err != nil {
		plog.Fatal(err.Error())
		return
	}
	plog.Infof("Scheduled task interval: %d seconds", settings.Cron.IntervalSeconds)

	plog.Debug("Starting scheduled task...")
	go harvest(ctx, settings)

	select {}
}

func harvest(ctx context.Context, settings *config.Settings) {
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

	plog.Info("Successfully initialised SQL tables.")

	harvestTweets(ctx, dataService, settings)

	interval := settings.Cron.Interval()
	for {
		select {
		case <-ctx.Done():
			plog.Info("Scheduled task stopped.")
		case <-time.After(interval):
			harvestTweets(ctx, dataService, settings)
		}
	}
}

func tryGet[T any](m map[string]any, key string) (T, bool) {
	var parsed T
	v, ok := m[key]
	if !ok {
		return parsed, false
	}
	parsed, ok = v.(T)
	return parsed, ok
}

func mustGet[T any](m map[string]any, key string) (T, error) {
	var parsed T
	v, ok := m[key]
	if !ok {
		return parsed, fmt.Errorf("cannot find key '%s' in map", key)
	}
	parsed, ok = v.(T)
	if !ok {
		return parsed, fmt.Errorf("cannot parse value for key %s", key)
	}
	return parsed, nil
}

func parseTweet(userID uid.UserID, tweet map[string]any) (*data.Toot, error) {
	buffer, err := json.Marshal(tweet)
	if err != nil {
		return nil, fmt.Errorf("error serialising tweet data: %w", err)
	}

	id, err := mustGet[string](tweet, "id")
	if err != nil {
		return nil, fmt.Errorf("error retrieving `id` value: %w", err)
	}

	createdAt, err := mustGet[string](tweet, "created_at")
	if err != nil {
		return nil, fmt.Errorf("Error retrieving `created_at` value: %w", err)
	}

	text, err := mustGet[string](tweet, "text")
	if err != nil {
		return nil, fmt.Errorf("Error retrieving `text` value: %w", err)
	}

	sourceID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing tweet ID: %w", err)
	}

	tootID := uid.New(userID, uid.Twitter, sourceID)

	return &data.Toot{
		ID:           tootID.String(),
		UserID:       userID.Int(),
		CreatedAt:    createdAt,
		TextOriginal: text,
		TextHTML:     "ToDo",
		SourceType:   uid.Twitter.Int(),
		SourceID:     id,
		SourceData:   string(buffer),
	}, nil
}

func harvestTweets(ctx context.Context, dataService *data.Service, settings *config.Settings) {
	for _, user := range settings.Users {
		plog.Infof("Collecting tweets for %s", user.Twitter.Username)

		userID := user.UserID()
		sinceId, err := dataService.LatestTweetID(ctx, userID)
		if err != nil {
			plog.Error(err.Error())
			continue
		}
		plog.Infof("Latest tweet ID: %s", sinceId)

		nextToken := ""
		for {
			result, err := twitter.GetTweetsByUser(
				ctx,
				user.Twitter.Username,
				user.Twitter.Token,
				user.StartDate,
				sinceId,
				nextToken)
			if err != nil {
				plog.Error(err.Error())
				continue
			}

			// Parse and check meta data:
			// ---
			meta, err := mustGet[map[string]any](result, "meta")
			if err != nil {
				plog.Error(err.Error())
				break
			}

			resultCount, err := mustGet[float64](meta, "result_count")
			if err != nil {
				plog.Error(err.Error())
				break
			}

			plog.Debugf("Result count: %d", int(resultCount))
			if resultCount == 0 {
				break
			}

			// Parse and save tweets:
			// ---
			elements, err := mustGet[[]any](result, "data")
			if err != nil {
				plog.Error(err.Error())
				break
			}

			for _, elem := range elements {
				toot, err := parseTweet(userID, elem.(map[string]any))
				if err != nil {
					plog.Error(err.Error())
					continue
				}

				err = dataService.SaveToot(ctx, toot)
				if err != nil {
					plog.Errorf("Error saving toot: %s", err.Error())
					continue
				}
				plog.Infof("Toot saved: %s", toot.ID)
			}

			// Parse user data:
			// ---
			includes, err := mustGet[map[string]any](result, "includes")
			if err != nil {
				plog.Error(err.Error())
				break
			}

			users, err := mustGet[[]any](includes, "users")
			if err != nil {
				plog.Error(err.Error())
				break
			}

			// Find all profile images:
			// ---
			profileImageURLs := make(map[string]string)
			for _, elem := range users {
				twitterUser := elem.(map[string]any)

				profileImageURL, ok := tryGet[string](twitterUser, "profile_image_url")
				if !ok || len(profileImageURL) == 0 {
					continue
				}

				username, ok := tryGet[string](twitterUser, "username")
				if !ok {
					continue
				}

				if strings.Contains(profileImageURL, "_normal.") {
					profileImageURL = strings.Replace(profileImageURL, "_normal", "", 1)
				}

				profileImageURLs[username] = profileImageURL
			}

			// Download profile images:
			// ---
			for _, user := range settings.Users {
				if profileImageURL, ok := profileImageURLs[user.Twitter.Username]; ok {
					plog.Debugf("Profile image found for user %s: %s", user.Username, profileImageURL)

					ext := ".jpg"
					parsedURL, err := url.Parse(profileImageURL)
					if err == nil {
						actualExt := path.Ext(parsedURL.Path)
						if len(actualExt) > 0 {
							ext = actualExt
						}
					}

					profileImagePath := settings.Storage.ProfileImageFullFilePath(user.UserID(), ext)
					err = download.File(ctx, profileImageURL, profileImagePath)
					if err != nil {
						plog.Error(err.Error())
						continue
					}

					plog.Infof("Profile image downloaded for user %s: %s", user.Username, profileImagePath)
				}
			}

			// Get token to next page:
			// ---
			tokenValue, ok := tryGet[string](meta, "next_token")
			if !ok {
				plog.Debug("No more tweets to collect.")
				break
			}
			nextToken = tokenValue
		}
	}
}
