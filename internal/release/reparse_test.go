// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package release

import (
	"testing"

	"github.com/autobrr/autobrr/internal/database"
	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Reparse(t *testing.T) {
	ctx := t.Context()
	log := logger.New(&domain.Config{LogLevel: "ERROR"}, nil)

	db, err := database.NewDB(&domain.Config{LogLevel: "ERROR", DatabaseDSN: ":memory:"}, log)
	require.NoError(t, err)
	require.NoError(t, db.Open())
	require.NoError(t, db.Migrate())
	t.Cleanup(func() { _ = db.Close() })

	repo := database.NewReleaseRepo(log, db)
	service := &Service{log: log, repo: repo}

	store := func(name, announcedTitle string) int64 {
		r := domain.NewRelease(domain.IndexerMinimal{Identifier: "nyaa"})
		r.Title = announcedTitle
		r.ParseString(name)
		require.NoError(t, repo.Store(ctx, r))
		return r.ID
	}
	get := func(id int64) *domain.Release {
		r, err := repo.Get(ctx, &domain.GetReleaseRequest{Id: int(id)})
		require.NoError(t, err)
		return r
	}

	asw := store("[ASW] Yowaki Max - 04 [1080p HEVC x265 10Bit][AAC]", "")
	store("The.Penguin.S01E06.1080p.WEB.h264-ETHEL", "")
	announced := store("Dune.Part.Two.2024.2160p.UHD.BluRay.x265-GROUP", "Dune: Part Two")

	result, err := service.Reparse(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.ReparseResult{Checked: 3, Updated: 1}, *result, "only the announced title differs from the name")
	assert.Equal(t, "Dune Part Two", get(announced).Title, "announced values are replaced too")

	result, err = service.Reparse(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.ReparseResult{Checked: 3, Updated: 0}, *result)

	before := get(asw)
	assert.Empty(t, before.Group)

	p, err := domain.NewVocabParser([]domain.VocabEntry{{Kind: "group", Tag: "ASW"}}, nil)
	require.NoError(t, err)
	domain.UseVocabParser(p)
	t.Cleanup(func() {
		p, err := domain.NewVocabParser(nil, nil)
		require.NoError(t, err)
		domain.UseVocabParser(p)
	})

	result, err = service.Reparse(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.ReparseResult{Checked: 3, Updated: 1}, *result)

	after := get(asw)
	assert.Equal(t, "ASW", after.Group)
	assert.Equal(t, "Yowaki Max", after.Title)
	assert.Equal(t, 4, after.Episode)
	assert.NotEqual(t, before.NormalizedHash, after.NormalizedHash)
}
