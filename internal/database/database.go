// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package database

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog"
)

const slowQueryThreshold = 500 * time.Millisecond

type SQLDB interface {
	Open() error
	Migrate() error
	Close() error
	ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Ping() error
	//BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error)
}

type DB struct {
	log     zerolog.Logger
	Handler *sql.DB
	lock    sync.RWMutex
	ctx     context.Context
	cfg     *domain.Config

	cancel func()

	DSN string

	squirrel sq.StatementBuilderType
}

func NewDB(cfg *domain.Config, log zerolog.Logger) (*DB, error) {
	db := &DB{
		squirrel: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		log:      log.With().Str("module", "database").Logger(),
		cfg:      cfg,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())

	switch {
	case cfg.DatabaseDSN == "":
		if os.Getenv("IS_TEST_ENV") == "true" {
			db.DSN = ":memory:"
		} else {
			db.DSN = dataSourceName(cfg.ConfigPath, "autobrr.db")
		}
	case strings.HasPrefix(cfg.DatabaseDSN, "file:") || cfg.DatabaseDSN == ":memory:" || strings.HasSuffix(cfg.DatabaseDSN, ".db"):
		if dsn, ok := strings.CutPrefix(cfg.DatabaseDSN, "file:"); ok && strings.HasSuffix(dsn, ".db") {
			db.DSN = dsn
		} else {
			db.DSN = cfg.DatabaseDSN
		}
	default:
		return nil, errors.New("unsupported database DSN: %s", cfg.DatabaseDSN)
	}

	return db, nil
}

func (db *DB) Open() error {
	if err := db.openSQLite(); err != nil {
		return errors.Wrap(err, "could not open sqlite db connection")
	}

	return nil
}

func (db *DB) Migrate() error {
	if err := db.migrateSQLite(); err != nil {
		return errors.Wrap(err, "could not migrate sqlite db")
	}

	return nil
}

func (db *DB) Close() error {
	db.log.Debug().Msg("closing database..")

	shutdownErr := db.closingSQLite()

	db.cancel()

	if db.Handler != nil {
		if err := db.Handler.Close(); err != nil {
			return errors.Wrap(err, "could not close database")
		}
	}

	if shutdownErr != nil {
		return errors.Wrap(shutdownErr, "could not run sqlite shutdown tasks")
	}

	return nil
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	start := time.Now()

	defer func() {
		duration := time.Since(start)

		logEvent := db.log.Trace()

		if err != nil {
			logEvent = db.log.Error().Err(err)
		} else if duration > slowQueryThreshold {
			logEvent = db.log.Warn().Bool("slow_query", true)
		}

		logEvent.Str("query", query).Interface("args", args).Dur("duration", duration).Msg("database query")
	}()

	rows, err = db.Handler.QueryContext(ctx, query, args...)
	return rows, err
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	start := time.Now()

	defer func() {
		duration := time.Since(start)

		logEvent := db.log.Trace()

		if duration > slowQueryThreshold {
			logEvent = db.log.Warn().Bool("slow_query", true)
		}

		logEvent.Str("query", query).Interface("args", args).Dur("duration", duration).Msg("database query")
	}()

	row = db.Handler.QueryRowContext(ctx, query, args...)
	return row
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	start := time.Now()

	defer func() {
		duration := time.Since(start)

		logEvent := db.log.Trace()

		if err != nil {
			logEvent = db.log.Error()
		} else if duration > slowQueryThreshold {
			logEvent = db.log.Warn().Bool("slow_query", true)
		}

		logEvent.Str("query", query).Interface("args", args).Dur("duration", duration).Msg("database query")
	}()

	result, err = db.Handler.ExecContext(ctx, query, args...)
	return result, err
}

func (db *DB) BeginTX(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.Handler.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:      tx,
		handler: db,
	}, nil
}

func (db *DB) Ping() error {
	return db.Handler.Ping()
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.Handler.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:      tx,
		handler: db,
	}, nil
}

type Tx struct {
	*sql.Tx
	handler *DB
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	start := time.Now()

	defer func() {
		duration := time.Since(start)

		logEvent := tx.handler.log.Trace()

		if err != nil {
			logEvent = tx.handler.log.Error()
		} else if duration > slowQueryThreshold {
			logEvent = tx.handler.log.Warn().Bool("slow_query", true)
		}

		logEvent.Str("query", query).Interface("args", args).Dur("duration", duration).Bool("in_transaction", true).Msg("database query")
	}()

	rows, err = tx.Tx.QueryContext(ctx, query, args...)
	return rows, err
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	start := time.Now()

	defer func() {
		duration := time.Since(start)

		logEvent := tx.handler.log.Trace()

		if duration > slowQueryThreshold {
			logEvent = tx.handler.log.Warn().Bool("slow_query", true)
		}

		logEvent.Str("query", query).Interface("args", args).Dur("duration", duration).Bool("in_transaction", true).Msg("database query")
	}()

	row = tx.Tx.QueryRowContext(ctx, query, args...)
	return row
}
