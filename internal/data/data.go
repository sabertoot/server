package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sabertoot/server/internal/uid"
)

type Toot struct {
	ID           string `json:"id"`
	UserID       int    `json:"user_id"`
	CreatedAt    string `json:"created_at"`
	TextOriginal string `json:"text_original"`
	TextHTML     string `json:"text_html"`
	SourceType   int    `json:"source_type"`
	SourceID     string `json:"source_id"`
	SourceData   string `json:"source_data"`
}

const (
	tootsTable = "toots"
)

func InitTables(ctx context.Context, db *sql.DB) error {
	// SQLite supported data types:
	// TEXT, NUMERIC, INTEGER, REAL, BLOB

	statement, err := db.PrepareContext(ctx,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
		(
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			text_original TEXT NOT NULL,
			text_html TEXT NOT NULL,
			source_type INTEGER NOT NULL,
			source_id TEXT NOT NULL,
			source_data TEXT NOT NULL
		)`, tootsTable))
	if err != nil {
		return fmt.Errorf("error preparing create statement for '%s' table: %w", tootsTable, err)
	}

	_, err = statement.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("error creating '%s' table: %w", tootsTable, err)
	}

	return nil
}

func SaveToot(ctx context.Context, db *sql.DB, t *Toot) error {
	statement, err := db.PrepareContext(ctx,
		fmt.Sprintf(`INSERT INTO %s
		(
			id,
			user_id,
			created_at,
			text_original,
			text_html,
			source_type,
			source_id,
			source_data
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, tootsTable))
	if err != nil {
		return fmt.Errorf("error preparing insert statement for '%s' table: %w", tootsTable, err)
	}

	_, err = statement.ExecContext(
		ctx,
		t.ID,
		t.UserID,
		t.CreatedAt,
		t.TextOriginal,
		t.TextHTML,
		t.SourceType,
		t.SourceID,
		t.SourceData)
	if err != nil {
		return fmt.Errorf("error inserting into '%s' table: %w", tootsTable, err)
	}

	return nil
}

func LatestTweetID(ctx context.Context, db *sql.DB, userID uid.UserID) (string, error) {
	var id string
	err := db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT source_id FROM %s WHERE source_type=? AND user_id=? ORDER BY id DESC LIMIT 1",
		tootsTable), uid.Twitter, userID).Scan(&id)

	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("error querying latest tweet id: %w", err)
	}

	return id, nil
}
