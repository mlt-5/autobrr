// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package release

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/autobrr/autobrr/internal/database"
	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readSheets returns each sheet's cells by reference, e.g. sheets[0]["B2"].
func readSheets(t *testing.T, data []byte) []map[string]string {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	type cell struct {
		Ref    string `xml:"r,attr"`
		Value  string `xml:"v"`
		Inline string `xml:"is>t"`
	}
	type worksheet struct {
		Rows []struct {
			Cells []cell `xml:"c"`
		} `xml:"sheetData>row"`
	}

	var sheets []map[string]string
	for i := 1; ; i++ {
		f, err := r.Open(fmt.Sprintf("xl/worksheets/sheet%d.xml", i))
		if err != nil {
			return sheets
		}

		content, err := io.ReadAll(f)
		require.NoError(t, err)

		var ws worksheet
		require.NoError(t, xml.Unmarshal(content, &ws))

		cells := map[string]string{}
		for _, row := range ws.Rows {
			for _, c := range row.Cells {
				cells[c.Ref] = c.Value + c.Inline
			}
		}
		sheets = append(sheets, cells)
	}
}

func TestService_Export(t *testing.T) {
	ctx := t.Context()
	log := logger.New(&domain.Config{LogLevel: "ERROR"}, nil)

	db, err := database.NewDB(&domain.Config{LogLevel: "ERROR", DatabaseDSN: ":memory:"}, log)
	require.NoError(t, err)
	require.NoError(t, db.Open())
	require.NoError(t, db.Migrate())
	t.Cleanup(func() { _ = db.Close() })

	repo := database.NewReleaseRepo(log, db)
	service := &Service{log: log, repo: repo}

	stamp := time.Date(2026, 10, 23, 19, 14, 7, 0, time.UTC)

	torrentLeech := domain.IndexerMinimal{Identifier: "tl"}

	// Enough older releases to need more than one page.
	for i := range exportPageSize + 5 {
		old := domain.NewRelease(torrentLeech)
		old.FilterStatus = domain.ReleaseStatusFilterApproved
		old.Implementation = domain.ReleaseImplementationRSS
		old.Timestamp = stamp.Add(-time.Duration(i+1) * time.Hour)
		old.TorrentName = fmt.Sprintf("Old.Release.%d-GRP", i)
		require.NoError(t, repo.Store(ctx, old))
	}

	penguin := domain.NewRelease(torrentLeech)
	penguin.FilterStatus = domain.ReleaseStatusFilterApproved
	penguin.FilterName = "TV 1080p"
	penguin.Implementation = domain.ReleaseImplementationRSS
	penguin.Timestamp = stamp
	penguin.TorrentName = "The.Penguin.S01E06.1080p.WEB.h264-ETHEL"
	penguin.Title = "The Penguin"
	penguin.Size = 2_814_600_000
	penguin.Codec = []string{"H.264"}
	penguin.Season = 1
	penguin.Episode = 6
	penguin.Group = "ETHEL"
	penguin.InfoURL = "https://example.com/torrent/240917?a=1&b=2"
	require.NoError(t, repo.Store(ctx, penguin))

	for _, status := range []domain.ReleaseActionStatus{
		{Status: domain.ReleasePushStatusApproved, Action: "Save to watch folder", Type: domain.ActionTypeWatchFolder, Filter: "TV 1080p", Rejections: []string{}, Timestamp: stamp.Add(2 * time.Second)},
		{Status: domain.ReleasePushStatusErr, Action: "Notify Sonarr", Type: domain.ActionTypeWebhook, Filter: "TV 1080p", Rejections: []string{"webhook returned 502 Bad Gateway"}, Timestamp: stamp.Add(3 * time.Second)},
	} {
		status.ReleaseID = penguin.ID
		require.NoError(t, repo.StoreReleaseActionStatus(ctx, &status))
	}

	stockholm, err := time.LoadLocation("Europe/Stockholm")
	require.NoError(t, err)

	t.Run("all", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, service.Export(ctx, &buf, domain.ReleaseQueryParams{}, stockholm))

		sheets := readSheets(t, buf.Bytes())
		require.Len(t, sheets, 2)
		releases, actions := sheets[0], sheets[1]

		assert.Equal(t, "ID", releases["A1"])
		assert.Equal(t, "Info URL", releases["V1"])

		assert.Equal(t, fmt.Sprint(penguin.ID), releases["A2"], "newest first")
		assert.Equal(t, "46318.88480324074", releases["B2"], "21:14:07 in Stockholm")
		assert.Equal(t, "The.Penguin.S01E06.1080p.WEB.h264-ETHEL", releases["C2"])
		assert.Equal(t, "tl", releases["E2"], "the identifier stands in when no indexer is configured")
		assert.Equal(t, "Approved", releases["G2"])
		assert.Equal(t, "Approved 1, Error 1", releases["H2"])
		assert.Equal(t, "2814.6", releases["K2"])
		assert.Equal(t, "H.264", releases["N2"])
		assert.Empty(t, releases["O2"], "no HDR")
		assert.Equal(t, "1", releases["Q2"])
		assert.Equal(t, "6", releases["R2"])
		assert.Empty(t, releases["S2"], "no year")
		assert.Equal(t, "RSS", releases["U2"])
		assert.Equal(t, "https://example.com/torrent/240917?a=1&b=2", releases["V2"])

		last := exportPageSize + 7
		assert.Equal(t, "Old.Release.0-GRP", releases[fmt.Sprintf("C%d", last)], "every page is exported, by id")
		assert.Empty(t, releases[fmt.Sprintf("C%d", last+1)])

		assert.Equal(t, fmt.Sprint(penguin.ID), actions["A2"])
		assert.Equal(t, "Save to watch folder", actions["E2"])
		assert.Equal(t, "Watch folder", actions["F2"])
		assert.Equal(t, "Approved", actions["G2"])
		assert.Equal(t, "Notify Sonarr", actions["E3"])
		assert.Equal(t, "Error", actions["G3"])
		assert.Equal(t, "webhook returned 502 Bad Gateway", actions["H3"])
		assert.Empty(t, actions["A4"])
	})

	t.Run("filtered", func(t *testing.T) {
		var buf bytes.Buffer
		params := domain.ReleaseQueryParams{}
		params.Filters.PushStatus = string(domain.ReleasePushStatusErr)
		require.NoError(t, service.Export(ctx, &buf, params, time.UTC))

		sheets := readSheets(t, buf.Bytes())
		assert.Equal(t, "The.Penguin.S01E06.1080p.WEB.h264-ETHEL", sheets[0]["C2"])
		assert.Equal(t, "46318.801469907405", sheets[0]["B2"], "19:14:07 UTC")
		assert.Empty(t, sheets[0]["C3"])
		assert.Equal(t, "Notify Sonarr", sheets[1]["E3"], "all of a matched release's actions are listed")
	})

	t.Run("empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, service.Export(ctx, &buf, domain.ReleaseQueryParams{Search: "Nothing.Matches"}, time.UTC))

		sheets := readSheets(t, buf.Bytes())
		require.Len(t, sheets, 2)
		assert.Empty(t, sheets[0]["A2"])
		assert.Empty(t, sheets[1]["A2"])
	})
}

func TestPushStatusSummary(t *testing.T) {
	approved := domain.ReleaseActionStatus{Status: domain.ReleasePushStatusApproved}
	rejected := domain.ReleaseActionStatus{Status: domain.ReleasePushStatusRejected}
	failed := domain.ReleaseActionStatus{Status: domain.ReleasePushStatusErr}

	assert.Empty(t, pushStatusSummary(nil))
	assert.Equal(t, "Approved", pushStatusSummary([]domain.ReleaseActionStatus{approved}))
	assert.Equal(t, "Approved (2)", pushStatusSummary([]domain.ReleaseActionStatus{approved, approved}))
	assert.Equal(t, "Approved 1, Rejected 1, Error 2", pushStatusSummary([]domain.ReleaseActionStatus{failed, approved, rejected, failed}))
}
