package db

// history.go holds the resolved-notification history: listing past entries,
// recording new ones, and purging entries past the retention window.

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
	VALUES ($1, $2, strftime('%Y-%m-%d %H:%M:%S', $3))
`
	historyPurgeSQL = `
	DELETE FROM history
	WHERE resolved_at < datetime('now', ?)
`
)

// HistoryEntry is a record of a notification that has been resolved, capturing
// when it was originally delivered and when it was cleared from the queue.
type HistoryEntry struct {
	ID         int64
	Session    string
	Message    string
	NotifiedAt time.Time
	ResolvedAt time.Time
}

// ListHistory returns all history entries, ordered by resolution time with the
// most recently resolved first.
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
		e.NotifiedAt, _ = time.Parse(time.RFC3339, notifiedAt)
		e.ResolvedAt, _ = time.Parse(time.RFC3339, resolvedAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// AddHistory inserts a history record for the given session, storing notifiedAt
// as the time the notification was originally delivered.
func AddHistory(q Querier, session, message string, notifiedAt time.Time) error {
	_, err := q.Exec(historyAddSQL, session, message, notifiedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("add history session=%q: %w", session, err)
	}
	return nil
}

// PurgeHistory deletes history entries that were resolved more than
// retentionDays ago.
func PurgeHistory(q Querier, retentionDays int) error {
	_, err := q.Exec(
		historyPurgeSQL,
		fmt.Sprintf("-%d days", retentionDays),
	)
	if err != nil {
		return fmt.Errorf("purge history retention=%d days: %w", retentionDays, err)
	}
	return nil
}
