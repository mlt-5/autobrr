// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package vocab

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"testing"

	"github.com/autobrr/autobrr/internal/database"
	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parse(name string) *domain.Release {
	r := domain.NewRelease(domain.IndexerMinimal{})
	r.ParseString(name)
	return r
}

func find(entries []domain.VocabEntry, kind, tag string, source domain.VocabSource) *domain.VocabEntry {
	for i := range entries {
		if entries[i].Kind == kind && entries[i].Tag == tag && entries[i].Source == source {
			return &entries[i]
		}
	}
	return nil
}

func TestService(t *testing.T) {
	ctx := t.Context()
	log := logger.New(&domain.Config{LogLevel: "ERROR"}, nil)

	db, err := database.NewDB(&domain.Config{LogLevel: "ERROR", DatabaseDSN: ":memory:"}, log)
	require.NoError(t, err)
	require.NoError(t, db.Open())
	require.NoError(t, db.Migrate())
	t.Cleanup(func() { _ = db.Close() })

	service := NewService(log.With().Logger(), database.NewVocabRepo(log.With().Logger(), db))
	t.Cleanup(func() {
		p, err := domain.NewVocabParser(nil, nil)
		require.NoError(t, err)
		domain.UseVocabParser(p)
	})

	require.NoError(t, service.Start(ctx))

	const fansub = "[ZqFansubs] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv"
	assert.Empty(t, parse(fansub).Group)

	entry := domain.VocabEntry{Kind: " group ", Tag: "ZqFansubs", Aliases: []string{"Zq Subs", ""}}
	require.NoError(t, service.Store(ctx, &entry))
	assert.NotZero(t, entry.ID)
	assert.Equal(t, domain.VocabSourceCustom, entry.Source)
	assert.Equal(t, "ZqFansubs", parse(fansub).Group, "an added entry is used at once")

	err = service.Store(ctx, &domain.VocabEntry{Kind: "group", Tag: "ZqFansubs"})
	assert.ErrorContains(t, err, "already added")

	err = service.Store(ctx, &domain.VocabEntry{Kind: "collection", Tag: "AMZN"})
	assert.ErrorContains(t, err, "already built in")

	entries, err := service.List(ctx)
	require.NoError(t, err)
	stored := find(entries, "group", "ZqFansubs", domain.VocabSourceCustom)
	require.NotNil(t, stored)
	assert.Equal(t, []string{"Zq Subs"}, stored.Aliases)

	require.NoError(t, service.Remove(ctx, domain.VocabKey{Kind: "group", Tag: "SubsPlease"}))
	assert.Empty(t, parse("[SubsPlease] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv").Group)

	assert.ErrorIs(t, service.Remove(ctx, domain.VocabKey{Kind: "group", Tag: "CODEX"}), domain.ErrRecordNotFound, "built-in tags can't be removed")

	entries, err = service.List(ctx)
	require.NoError(t, err)
	assert.True(t, find(entries, "group", "SubsPlease", domain.VocabSourceList).Removed)

	// A restart builds the parser from what is stored.
	p, err := domain.NewVocabParser(nil, nil)
	require.NoError(t, err)
	domain.UseVocabParser(p)
	require.NoError(t, service.Start(ctx))
	assert.Equal(t, "ZqFansubs", parse(fansub).Group)
	assert.Empty(t, parse("[SubsPlease] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv").Group)

	require.NoError(t, service.Restore(ctx, domain.VocabKey{Kind: "group", Tag: "SubsPlease"}))
	assert.Equal(t, "SubsPlease", parse("[SubsPlease] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv").Group)

	require.NoError(t, service.Delete(ctx, entry.ID))
	assert.Empty(t, parse(fansub).Group)
	assert.ErrorIs(t, service.Delete(ctx, entry.ID), domain.ErrRecordNotFound)
}

// readSheet returns the first sheet's rows as text.
func readSheet(t *testing.T, data []byte) [][]string {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	f, err := r.Open("xl/worksheets/sheet1.xml")
	require.NoError(t, err)
	content, err := io.ReadAll(f)
	require.NoError(t, err)

	var sheet struct {
		Rows []struct {
			Cells []struct {
				Ref  string `xml:"r,attr"`
				Text string `xml:"is>t"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	require.NoError(t, xml.Unmarshal(content, &sheet))

	var rows [][]string
	for _, row := range sheet.Rows {
		cells := make([]string, len(exportColumns))
		for _, c := range row.Cells {
			cells[int(c.Ref[0]-'A')] = c.Text
		}
		rows = append(rows, cells)
	}
	return rows
}

func TestService_Export(t *testing.T) {
	ctx := t.Context()
	log := logger.New(&domain.Config{LogLevel: "ERROR"}, nil)

	db, err := database.NewDB(&domain.Config{LogLevel: "ERROR", DatabaseDSN: ":memory:"}, log)
	require.NoError(t, err)
	require.NoError(t, db.Open())
	require.NoError(t, db.Migrate())
	t.Cleanup(func() { _ = db.Close() })

	service := NewService(log.With().Logger(), database.NewVocabRepo(log.With().Logger(), db))
	t.Cleanup(func() {
		p, err := domain.NewVocabParser(nil, nil)
		require.NoError(t, err)
		domain.UseVocabParser(p)
	})

	require.NoError(t, service.Store(ctx, &domain.VocabEntry{Kind: "group", Tag: "ZqSubs", Aliases: []string{"Zq Fansubs"}}))
	require.NoError(t, service.Remove(ctx, domain.VocabKey{Kind: "group", Tag: "SubsPlease"}))

	var buf bytes.Buffer
	require.NoError(t, service.Export(ctx, &buf, "group", "subs"))

	rows := readSheet(t, buf.Bytes())
	require.Len(t, rows, 3, "the header, SubsPlease and ZqSubs")
	assert.Equal(t, []string{"Kind", "Tag", "Title", "Other spellings", "Pattern", "Release type", "Only for this type", "Source", "Removed"}, rows[0])
	assert.Equal(t, []string{"Release group", "SubsPlease", "", "", `\QSubsPlease\E`, "", "", "List", "Yes"}, rows[1])
	assert.Equal(t, []string{"Release group", "ZqSubs", "", "Zq Fansubs", "", "", "", "Added", ""}, rows[2])

	buf.Reset()
	require.NoError(t, service.Export(ctx, &buf, "", ""))
	rows = readSheet(t, buf.Bytes())
	assert.Greater(t, len(rows), 1000, "every entry without a filter")
}
