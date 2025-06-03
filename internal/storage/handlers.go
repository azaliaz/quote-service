package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

func (db *DB) AddQuote(ctx context.Context, quote *Quote) (int64, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			db.log.Error("rollback error", slog.String("err", err.Error()))
		}
	}()

	var id int64
	err = tx.QueryRow(ctx,
		`INSERT INTO quotes (author, quote, created_at)
		 VALUES ($1, $2, NOW())
		 RETURNING id`,
		quote.Author, quote.Quote).Scan(&id)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DB) GetAllQuotes(ctx context.Context) ([]*Quote, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx,
		`SELECT id, author, quote, created_at FROM quotes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []*Quote
	for rows.Next() {
		var q Quote
		if err := rows.Scan(&q.ID, &q.Author, &q.Quote, &q.CreatedAt); err != nil {
			return nil, err
		}
		quotes = append(quotes, &q)
	}
	return quotes, rows.Err()
}

func (db *DB) GetRandomQuote(ctx context.Context) (*Quote, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var q Quote
	err = conn.QueryRow(ctx,
		`SELECT id, author, quote, created_at FROM quotes ORDER BY RANDOM() LIMIT 1`).Scan(
		&q.ID, &q.Author, &q.Quote, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (db *DB) GetQuotesByAuthor(ctx context.Context, author string) ([]*Quote, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx,
		`SELECT id, author, quote, created_at FROM quotes WHERE author = $1 ORDER BY created_at DESC`, author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []*Quote
	for rows.Next() {
		var q Quote
		if err := rows.Scan(&q.ID, &q.Author, &q.Quote, &q.CreatedAt); err != nil {
			return nil, err
		}
		quotes = append(quotes, &q)
	}
	return quotes, rows.Err()
}

func (db *DB) DeleteQuote(ctx context.Context, id int64) error {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx,
		`DELETE FROM quotes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("quote with id %d not found", id)
	}
	return nil
}
