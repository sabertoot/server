package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sabertoot/server/internal/uid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

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

func (svc *Service) InitTables(ctx context.Context) error {
	// SQLite supported data types:
	// TEXT, NUMERIC, INTEGER, REAL, BLOB

	statement, err := svc.db.PrepareContext(ctx,
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

func (svc *Service) SaveToot(ctx context.Context, t *Toot) error {
	statement, err := svc.db.PrepareContext(ctx,
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

func (svc *Service) LatestTweetID(ctx context.Context, userID uid.UserID) (string, error) {
	var id string
	err := svc.db.QueryRowContext(ctx, fmt.Sprintf(
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

func (svc *Service) TootCount(ctx context.Context, userID uid.UserID) (int, error) {
	var count int
	err := svc.db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE user_id=?",
		tootsTable), userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("error querying toot count: %w", err)
	}

	return count, nil
}
