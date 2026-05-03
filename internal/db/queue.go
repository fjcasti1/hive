// internal/db/queue.go
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
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
	RETURNING session, message, created_at
`
)

type queueEntry struct {
	ID        int64
	Session   string
	Pane      string
	Message   string
	CreatedAt time.Time
}

func (e queueEntry) Target() string {
	return e.Session
}

func Enqueue(q Querier, session, message, pane string) error {
	_, err := q.Exec(queueAddSQL, session, message, pane)
	if err != nil {
		return fmt.Errorf("enqueue session=%q: %w", session, err)
	}
	return nil
}

func List(q Querier) ([]queueEntry, error) {
	rows, err := q.Query(queueListSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []queueEntry
	for rows.Next() {
		var (
			e         queueEntry
			createdAt string
		)
		if err := rows.Scan(&e.ID, &e.Session, &e.Pane, &e.Message, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

type deletedQueueEntry struct {
	Session    string
	Message    string
	NotifiedAt time.Time
}

func Delete(q Querier, session string) (*deletedQueueEntry, error) {
	var (
		e         deletedQueueEntry
		createdAt string
	)
	err := q.QueryRow(queueDeleteSQL, session).Scan(&e.Session, &e.Message, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("delete session=%q: %w", session, err)
	}
	e.NotifiedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &e, nil
}
