package data

import (
	"context"
	"database/sql"
	"fmt"
)

// tweets
// ---
// id
// created_at
// data

// toots
// ---
// id
// twitter_id
// created_at
// text_original
// text_html
// attachments

// TEXT, NUMERIC, INTEGER, REAL, BLOB

type Tweet struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Handle    string `json:"handle"`
	Data      string `json:"data"`
}

func InitTables(ctx context.Context, db *sql.DB) error {
	statement, err := db.PrepareContext(ctx, "CREATE TABLE IF NOT EXISTS tweets (id INTEGER PRIMARY KEY, created_at TEXT NOT NULL, handle TEXT NOT NULL, data TEXT NOT NULL)")
	if err != nil {
		return fmt.Errorf("error preparing create statement for 'tweets' table: %w", err)
	}

	_, err = statement.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("error creating 'tweets' table: %w", err)
	}

	return nil
}

func SaveTweet(ctx context.Context, db *sql.DB, tweet *Tweet) error {
	statement, err := db.PrepareContext(ctx, "INSERT INTO tweets (id, created_at, handle, data) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing insert statement for 'tweets' table: %w", err)
	}

	_, err = statement.ExecContext(ctx, tweet.ID, tweet.CreatedAt, tweet.Handle, tweet.Data)
	if err != nil {
		return fmt.Errorf("error inserting into 'tweets' table: %w", err)
	}

	return nil
}

func LatestTweetID(ctx context.Context, db *sql.DB, handle string) (string, error) {
	var id string
	err := db.QueryRowContext(
		ctx, "SELECT id FROM tweets WHERE handle=? ORDER BY id DESC LIMIT 1", handle).Scan(&id)

	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("error querying latest tweet id: %w", err)
	}

	return id, nil
}
