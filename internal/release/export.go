// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package release

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"
	"github.com/autobrr/autobrr/pkg/xlsx"
)

// ExportLimit is the most releases one export holds.
const ExportLimit = 100_000

const exportPageSize = 1000

var releaseColumns = []xlsx.Column{
	{Name: "ID", Width: 8},
	{Name: "Time", Width: 19},
	{Name: "Name", Width: 60},
	{Name: "Title", Width: 24},
	{Name: "Indexer", Width: 16},
	{Name: "Filter", Width: 18},
	{Name: "Filter status", Width: 13},
	{Name: "Push status", Width: 20},
	{Name: "Rejections", Width: 40},
	{Name: "Category", Width: 14},
	{Name: "Size (MB)", Width: 11},
	{Name: "Resolution", Width: 11},
	{Name: "Source", Width: 12},
	{Name: "Codec", Width: 10},
	{Name: "HDR", Width: 10},
	{Name: "Group", Width: 14},
	{Name: "Season", Width: 8},
	{Name: "Episode", Width: 8},
	{Name: "Year", Width: 7},
	{Name: "Protocol", Width: 9},
	{Name: "Source type", Width: 11},
	{Name: "Info URL", Width: 42},
}

var actionColumns = []xlsx.Column{
	{Name: "Release ID", Width: 11},
	{Name: "Release name", Width: 60},
	{Name: "Time", Width: 19},
	{Name: "Filter", Width: 18},
	{Name: "Action", Width: 22},
	{Name: "Type", Width: 13},
	{Name: "Status", Width: 10},
	{Name: "Rejections", Width: 40},
}

// Export writes the releases matching params, newest first, as an Excel workbook with a Releases sheet and an
// Actions sheet holding each release's action results. Times are shown in loc.
func (s *Service) Export(ctx context.Context, w io.Writer, params domain.ReleaseQueryParams, loc *time.Location) error {
	book := xlsx.NewWriter(w)

	if err := book.AddSheet("Releases", releaseColumns); err != nil {
		return err
	}

	// The Actions sheet is filled by a second pass over the same id range, so new releases stored while exporting
	// do not get action rows without a release row.
	var newest, oldest int64
	count := 0

	err := s.exportPages(ctx, params, 0, func(release *domain.Release) (bool, error) {
		if count == ExportLimit {
			return false, book.WriteRow(xlsx.Empty, xlsx.String(fmt.Sprintf("Export stopped at %d releases. Narrow the search to export older releases.", ExportLimit)))
		}

		if newest == 0 {
			newest = release.ID
		}
		oldest = release.ID
		count++

		return true, book.WriteRow(releaseRow(release, loc)...)
	})
	if err != nil {
		return err
	}

	if err := book.AddSheet("Actions", actionColumns); err != nil {
		return err
	}

	if count > 0 {
		err = s.exportPages(ctx, params, uint64(newest)+1, func(release *domain.Release) (bool, error) {
			if release.ID < oldest {
				return false, nil
			}

			for _, status := range release.ActionStatus {
				if err := book.WriteRow(actionRow(release, status, loc)...); err != nil {
					return false, err
				}
			}

			return true, nil
		})
		if err != nil {
			return err
		}
	}

	return book.Close()
}

// exportPages calls fn for each release matching params with an id below cursor (or any id when cursor is 0),
// newest first, until fn returns false or the releases run out.
func (s *Service) exportPages(ctx context.Context, params domain.ReleaseQueryParams, cursor uint64, fn func(*domain.Release) (bool, error)) error {
	params.Limit = exportPageSize
	params.Offset = 0
	params.Cursor = cursor

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		page, err := s.repo.Find(ctx, params)
		if err != nil {
			return errors.Wrap(err, "could not find releases to export")
		}

		for _, release := range page.Data {
			more, err := fn(release)
			if err != nil || !more {
				return err
			}
		}

		if len(page.Data) < exportPageSize || page.NextCursor <= 0 {
			return nil
		}

		params.Cursor = uint64(page.NextCursor)
	}
}

func releaseRow(r *domain.Release, loc *time.Location) []xlsx.Cell {
	size := xlsx.Empty
	if r.Size > 0 {
		size = xlsx.Decimal(float64(r.Size) / 1_000_000)
	}

	return []xlsx.Cell{
		xlsx.Int(int(r.ID)),
		xlsx.Date(r.Timestamp.In(loc)),
		xlsx.String(r.TorrentName),
		xlsx.String(r.Title),
		xlsx.String(r.Indexer.Name),
		xlsx.String(r.FilterName),
		xlsx.String(filterStatusLabel(r.FilterStatus)),
		xlsx.String(pushStatusSummary(r.ActionStatus)),
		xlsx.String(strings.Join(nonEmpty(r.Rejections), "; ")),
		xlsx.String(r.Category),
		size,
		xlsx.String(r.Resolution),
		xlsx.String(r.Source),
		xlsx.String(strings.Join(nonEmpty(r.Codec), ", ")),
		xlsx.String(strings.Join(nonEmpty(r.HDR), ", ")),
		xlsx.String(r.Group),
		xlsx.Int(r.Season),
		xlsx.Int(r.Episode),
		xlsx.Int(r.Year),
		xlsx.String(string(r.Protocol)),
		xlsx.String(implementationLabel(r.Implementation)),
		xlsx.Link(r.InfoURL),
	}
}

func actionRow(r *domain.Release, status domain.ReleaseActionStatus, loc *time.Location) []xlsx.Cell {
	return []xlsx.Cell{
		xlsx.Int(int(r.ID)),
		xlsx.String(r.TorrentName),
		xlsx.Date(status.Timestamp.In(loc)),
		xlsx.String(status.Filter),
		xlsx.String(status.Action),
		xlsx.String(actionTypeLabel(status.Type)),
		xlsx.String(status.Status.String()),
		xlsx.String(strings.Join(nonEmpty(status.Rejections), "; ")),
	}
}

// pushStatusSummary sums up a release's action results, e.g. "Approved (2)" or "Approved 1, Error 1".
func pushStatusSummary(statuses []domain.ReleaseActionStatus) string {
	order := []domain.ReleasePushStatus{domain.ReleasePushStatusApproved, domain.ReleasePushStatusRejected, domain.ReleasePushStatusErr, domain.ReleasePushStatusPending}

	counts := map[domain.ReleasePushStatus]int{}
	for _, status := range statuses {
		counts[status.Status]++
	}

	var parts []string
	for _, status := range order {
		if n := counts[status]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", status.String(), n))
		}
	}

	switch {
	case len(parts) == 0:
		return ""
	case len(parts) == 1 && len(statuses) == 1:
		return statuses[0].Status.String()
	case len(parts) == 1:
		return fmt.Sprintf("%s (%d)", statuses[0].Status.String(), len(statuses))
	default:
		return strings.Join(parts, ", ")
	}
}

func filterStatusLabel(status domain.ReleaseFilterStatus) string {
	switch status {
	case domain.ReleaseStatusFilterApproved:
		return "Approved"
	case domain.ReleaseStatusFilterRejected:
		return "Rejected"
	case domain.ReleaseStatusFilterPending:
		return "Pending"
	default:
		return string(status)
	}
}

func implementationLabel(implementation domain.ReleaseImplementation) string {
	switch implementation {
	case domain.ReleaseImplementationIRC:
		return "IRC"
	case domain.ReleaseImplementationTorznab:
		return "Torznab"
	case domain.ReleaseImplementationRSS:
		return "RSS"
	default:
		return string(implementation)
	}
}

func actionTypeLabel(actionType domain.ActionType) string {
	switch actionType {
	case domain.ActionTypeWatchFolder:
		return "Watch folder"
	case domain.ActionTypeWebhook:
		return "Webhook"
	case domain.ActionTypeTest:
		return "Test"
	default:
		return string(actionType)
	}
}

func nonEmpty(values []string) []string {
	var out []string
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
