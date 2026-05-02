// internal/db/queue.go
package db

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	sqliteTimeLayout = "2006-01-02T15:04:05Z"

	queueAddSQL = `
	INSERT INTO queue (session, message, pane)
	VALUES ($1, $2, $3)
	ON CONFLICT(session) DO UPDATE SET
		message    = excluded.message,
	 	pane       = excluded.pane,
		created_at = CURRENT_TIMESTAMP
`

	queueListSQL = `
	SELECT
		id,
		session,
		pane,
		message,
		created_at
	FROM queue
	ORDER BY created_at ASC
`

	queueDeleteSQL = `
	DELETE FROM queue
	WHERE session = $1
`
)

type QueueEntry struct {
	ID        int64
	Session   string
	Pane      string
	Message   string
	CreatedAt time.Time
}

func (e QueueEntry) Target() string {
	return e.Session
}

func Enqueue(database *sql.DB, session, message, pane string) error {
	_, err := database.Exec(queueAddSQL, session, message, pane)
	if err != nil {
		return fmt.Errorf("enqueue session=%q: %w", session, err)
	}
	return nil
}

func List(database *sql.DB) ([]QueueEntry, error) {
	rows, err := database.Query(queueListSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []QueueEntry
	for rows.Next() {
		var (
			e         QueueEntry
			createdAt string
		)
		if err := rows.Scan(&e.ID, &e.Session, &e.Pane, &e.Message, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(sqliteTimeLayout, createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func Delete(database *sql.DB, session string) (bool, error) {
	res, err := database.Exec(queueDeleteSQL, session)
	if err != nil {
		return false, fmt.Errorf("delete session=%q: %w", session, err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
