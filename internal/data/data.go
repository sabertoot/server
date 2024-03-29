package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	ID           uid.TootID
	UserID       uid.UserID
	CreatedAt    time.Time
	TextOriginal string
	TextHTML     string
	SourceType   uid.SourceType
	SourceID     string
	SourceData   string
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
			created_at INTEGER NOT NULL,
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
		t.ID.String(),
		t.UserID.Int(),
		t.CreatedAt.Unix(),
		t.TextOriginal,
		t.TextHTML,
		t.SourceType.Int(),
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
		tootsTable), uid.Twitter, userID.Int()).Scan(&id)

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
		tootsTable), userID.Int()).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("error querying toot count: %w", err)
	}

	return count, nil
}

func (svc *Service) Toots(
	ctx context.Context,
	userID uid.UserID,
	after int64,
	limit int,
) (
	[]*Toot, error,
) {
	rows, err := svc.db.QueryContext(ctx, fmt.Sprintf(
		"SELECT * FROM %s WHERE user_id=? AND created_at>? ORDER BY created_at DESC LIMIT ?",
		tootsTable), userID.Int(), after, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying toots: %w", err)
	}
	defer rows.Close()

	toots := []*Toot{}
	for rows.Next() {
		t := &Toot{}
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.CreatedAt,
			&t.TextOriginal,
			&t.TextHTML,
			&t.SourceType,
			&t.SourceID,
			&t.SourceData)
		if err != nil {
			return nil, fmt.Errorf("error scanning toot: %w", err)
		}

		toots = append(toots, t)
	}

	return toots, nil
}
