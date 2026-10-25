// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package database

import (
	"context"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

// ListReparse returns up to limit stored releases with an ID above afterID, lowest first, with their name and
// parsed fields.
func (repo *ReleaseRepo) ListReparse(ctx context.Context, afterID int64, limit uint64) ([]domain.ReleaseReparse, error) {
	query, args, err := repo.db.squirrel.
		Select(
			"id", "torrent_name",
			"COALESCE(title, '')", "COALESCE(sub_title, '')",
			"COALESCE(season, 0)", "COALESCE(episode, 0)", "COALESCE(year, 0)", "COALESCE(month, 0)", "COALESCE(day, 0)",
			"COALESCE(resolution, '')", "COALESCE(source, '')", "COALESCE(codec, '')", "COALESCE(container, '')",
			"COALESCE(hdr, '')", "COALESCE(audio, '')", "COALESCE(audio_channels, '')", "COALESCE(release_group, '')",
			"COALESCE(region, '')", "COALESCE(language, '')", "COALESCE(edition, '')", "COALESCE(cut, '')",
			"COALESCE(website, '')", "COALESCE(media_processing, '')", "COALESCE(type, '')",
			"COALESCE(proper, FALSE)", "COALESCE(repack, FALSE)", "COALESCE(hybrid, FALSE)",
			"other", "COALESCE(normalized_hash, '')",
		).
		From("release").
		Where(sq.Gt{"id": afterID}).
		OrderBy("id").
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	rows, err := repo.db.Handler.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}
	defer rows.Close()

	var releases []domain.ReleaseReparse
	for rows.Next() {
		var r domain.ReleaseReparse
		f := &r.Fields
		if err := rows.Scan(
			&r.ID, &r.TorrentName, &f.Title, &f.SubTitle, &f.Season, &f.Episode, &f.Year, &f.Month, &f.Day,
			&f.Resolution, &f.Source, &f.Codec, &f.Container, &f.HDR, &f.Audio, &f.AudioChannels, &f.Group,
			&f.Region, &f.Language, &f.Edition, &f.Cut, &f.Website, &f.MediaProcessing, &f.Type,
			&f.Proper, &f.Repack, &f.Hybrid, pq.Array(&f.Other), &f.NormalizedHash,
		); err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}
		if f.Other == nil {
			f.Other = []string{}
		}
		releases = append(releases, r)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading rows")
	}

	return releases, nil
}

// UpdateParsedFields stores new parsed fields for the releases in one transaction.
func (repo *ReleaseRepo) UpdateParsedFields(ctx context.Context, releases []domain.ReleaseReparse) error {
	if len(releases) == 0 {
		return nil
	}

	tx, err := repo.db.Handler.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error begin transaction")
	}
	defer tx.Rollback()

	for _, r := range releases {
		f := r.Fields
		query, args, err := repo.db.squirrel.
			Update("release").
			Set("title", f.Title).
			Set("sub_title", f.SubTitle).
			Set("season", f.Season).
			Set("episode", f.Episode).
			Set("year", f.Year).
			Set("month", f.Month).
			Set("day", f.Day).
			Set("resolution", f.Resolution).
			Set("source", f.Source).
			Set("codec", f.Codec).
			Set("container", f.Container).
			Set("hdr", f.HDR).
			Set("audio", f.Audio).
			Set("audio_channels", f.AudioChannels).
			Set("release_group", f.Group).
			Set("region", f.Region).
			Set("language", f.Language).
			Set("edition", f.Edition).
			Set("cut", f.Cut).
			Set("website", f.Website).
			Set("media_processing", f.MediaProcessing).
			Set("type", f.Type).
			Set("proper", f.Proper).
			Set("repack", f.Repack).
			Set("hybrid", f.Hybrid).
			Set("other", pq.Array(f.Other)).
			Set("normalized_hash", f.NormalizedHash).
			Where(sq.Eq{"id": r.ID}).
			ToSql()
		if err != nil {
			return errors.Wrap(err, "error building query")
		}

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return errors.Wrap(err, "error updating release %d", r.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "error committing parsed fields")
	}

	return nil
}
