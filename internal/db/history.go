package db

import (
	"fmt"
	"time"
)

const (
	historyListSQL = `
	SELECT
		id,
		session,
		message,
		notified_at,
		resolved_at
	FROM history
	ORDER BY
		resolved_at DESC,
		id DESC
`
	historyAddSQL = `
	INSERT INTO history (session, message, notified_at)
	VALUES ($1, $2, $3)
`
)

type HistoryEntry struct {
	ID         int64
	Session    string
	Message    string
	NotifiedAt time.Time
	ResolvedAt time.Time
}

func ListHistory(q Querier) ([]HistoryEntry, error) {
	rows, err := q.Query(historyListSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var (
			e          HistoryEntry
			notifiedAt string
			resolvedAt string
		)
		if err := rows.Scan(&e.ID, &e.Session, &e.Message, &notifiedAt, &resolvedAt); err != nil {
			return nil, err
		}
		e.NotifiedAt, _ = time.Parse(sqliteTimeLayout, notifiedAt)
		e.ResolvedAt, _ = time.Parse(sqliteTimeLayout, resolvedAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func AddHistory(q Querier, session, message string, notifiedAt time.Time) error {
	_, err := q.Exec(historyAddSQL, session, message, notifiedAt.UTC().Format(sqliteTimeLayout))
	if err != nil {
		return fmt.Errorf("add history session=%q: %w", session, err)
	}
	return nil
}
