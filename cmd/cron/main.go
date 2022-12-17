package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sabertoot/server/internal/config"
	"github.com/sabertoot/server/internal/data"
	"github.com/sabertoot/server/internal/plog"
	"github.com/sabertoot/server/internal/twitter"

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
	go scheduledTask(ctx, settings)

	select {}
}

func scheduledTask(ctx context.Context, settings *config.Settings) {
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

	if err = data.InitTables(ctx, db); err != nil {
		plog.Fatal(err.Error())
		return
	}

	plog.Info("Successfully initialised SQL tables.")

	harvest(ctx, db, settings)

	interval := settings.Cron.Interval()
	for {
		select {
		case <-ctx.Done():
			plog.Info("Scheduled task stopped.")
		case <-time.After(interval):
			harvest(ctx, db, settings)
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

func parseTweet(tweet map[string]any) (*data.Tweet, error) {
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

	return &data.Tweet{
		ID:        id,
		CreatedAt: createdAt,
		Data:      string(buffer),
	}, nil
}

func harvest(ctx context.Context, db *sql.DB, settings *config.Settings) {
	for _, account := range settings.Accounts {
		plog.Infof("Collecting tweets for %s", account.Twitter.Handle)

		sinceId, err := data.LatestTweetID(ctx, db, account.Twitter.Handle)
		if err != nil {
			plog.Error(err.Error())
			continue
		}
		plog.Infof("Latest tweet ID: %s", sinceId)

		nextToken := ""
		for {
			result, err := twitter.GetTweetsByUser(
				ctx,
				account.Twitter.Handle,
				account.Twitter.Token,
				account.StartDate,
				sinceId,
				nextToken)
			if err != nil {
				plog.Error(err.Error())
				continue
			}

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

			elements, err := mustGet[[]any](result, "data")
			if err != nil {
				plog.Error(err.Error())
				break
			}

			for _, elem := range elements {
				tweet, err := parseTweet(elem.(map[string]any))
				if err != nil {
					plog.Error(err.Error())
					continue
				}

				tweet.Handle = account.Twitter.Handle

				err = data.SaveTweet(ctx, db, tweet)
				if err != nil {
					plog.Errorf("Error saving tweet: %s", err.Error())
					continue
				}
				plog.Infof("Tweet saved: %s", tweet.ID)
			}

			// toots
			// ---
			// id
			// twitter_id
			// created_at
			// text_original
			// text_html
			// attachments

			tokenValue, ok := tryGet[string](meta, "next_token")
			if !ok {
				plog.Debug("No more tweets to collect.")
				break
			}
			nextToken = tokenValue
		}
	}
}
