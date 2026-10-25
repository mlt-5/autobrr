// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package database

import (
	"context"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
)

type VocabRepo struct {
	log zerolog.Logger
	db  *DB
}

func NewVocabRepo(log zerolog.Logger, db *DB) *VocabRepo {
	return &VocabRepo{
		log: log.With().Str("repo", "vocab").Logger(),
		db:  db,
	}
}

// List returns the custom vocabulary entries in the order they were added.
func (r *VocabRepo) List(ctx context.Context) ([]domain.VocabEntry, error) {
	query, args, err := r.db.squirrel.
		Select("id", "kind", "tag", "title", "aliases", "pattern", "release_type", "exclusive").
		From("release_vocab").
		OrderBy("id").
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	rows, err := r.db.Handler.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}
	defer rows.Close()

	entries := []domain.VocabEntry{}
	for rows.Next() {
		entry := domain.VocabEntry{Source: domain.VocabSourceCustom}
		if err := rows.Scan(&entry.ID, &entry.Kind, &entry.Tag, &entry.Title, pq.Array(&entry.Aliases), &entry.Pattern, &entry.ReleaseType, &entry.Exclusive); err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}
		if entry.Aliases == nil {
			entry.Aliases = []string{}
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading rows")
	}

	return entries, nil
}

// Store adds a custom vocabulary entry and sets its ID.
func (r *VocabRepo) Store(ctx context.Context, entry *domain.VocabEntry) error {
	err := r.db.squirrel.
		Insert("release_vocab").
		Columns("kind", "tag", "title", "aliases", "pattern", "release_type", "exclusive").
		Values(entry.Kind, entry.Tag, entry.Title, pq.Array(entry.Aliases), entry.Pattern, entry.ReleaseType, entry.Exclusive).
		Suffix("RETURNING id").
		RunWith(r.db.Handler).
		QueryRowContext(ctx).
		Scan(&entry.ID)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}

	return nil
}

// Update replaces a custom vocabulary entry.
func (r *VocabRepo) Update(ctx context.Context, entry *domain.VocabEntry) error {
	query, args, err := r.db.squirrel.
		Update("release_vocab").
		Set("kind", entry.Kind).
		Set("tag", entry.Tag).
		Set("title", entry.Title).
		Set("aliases", pq.Array(entry.Aliases)).
		Set("pattern", entry.Pattern).
		Set("release_type", entry.ReleaseType).
		Set("exclusive", entry.Exclusive).
		Where(sq.Eq{"id": entry.ID}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	result, err := r.db.Handler.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}

	if n, err := result.RowsAffected(); err != nil {
		return errors.Wrap(err, "error getting rows affected")
	} else if n == 0 {
		return domain.ErrRecordNotFound
	}

	return nil
}

// Delete removes a custom vocabulary entry.
func (r *VocabRepo) Delete(ctx context.Context, id int64) error {
	query, args, err := r.db.squirrel.
		Delete("release_vocab").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	result, err := r.db.Handler.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}

	if n, err := result.RowsAffected(); err != nil {
		return errors.Wrap(err, "error getting rows affected")
	} else if n == 0 {
		return domain.ErrRecordNotFound
	}

	return nil
}

// ListRemoved returns the new_parser list entries that were removed.
func (r *VocabRepo) ListRemoved(ctx context.Context) ([]domain.VocabKey, error) {
	query, args, err := r.db.squirrel.
		Select("kind", "tag").
		From("release_vocab_removed").
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	rows, err := r.db.Handler.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}
	defer rows.Close()

	keys := []domain.VocabKey{}
	for rows.Next() {
		var key domain.VocabKey
		if err := rows.Scan(&key.Kind, &key.Tag); err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading rows")
	}

	return keys, nil
}

// Remove marks a new_parser list entry as removed.
func (r *VocabRepo) Remove(ctx context.Context, key domain.VocabKey) error {
	query, args, err := r.db.squirrel.
		Insert("release_vocab_removed").
		Columns("kind", "tag").
		Values(key.Kind, key.Tag).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	if _, err := r.db.Handler.ExecContext(ctx, query, args...); err != nil {
		return errors.Wrap(err, "error executing query")
	}

	return nil
}

// Restore brings back a removed new_parser list entry.
func (r *VocabRepo) Restore(ctx context.Context, key domain.VocabKey) error {
	query, args, err := r.db.squirrel.
		Delete("release_vocab_removed").
		Where(sq.Eq{"kind": key.Kind, "tag": key.Tag}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	if _, err := r.db.Handler.ExecContext(ctx, query, args...); err != nil {
		return errors.Wrap(err, "error executing query")
	}

	return nil
}
