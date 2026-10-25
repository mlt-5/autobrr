// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package release

import (
	"context"

	"github.com/autobrr/autobrr/internal/domain"
)

const reparseBatchSize = 500

// Reparse parses every stored release's name again with the vocabulary in use and stores the fields that changed.
// It replaces values an indexer announced as well, and leaves filter results and actions as they were.
func (s *Service) Reparse(ctx context.Context) (*domain.ReparseResult, error) {
	var result domain.ReparseResult

	var afterID int64
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		batch, err := s.repo.ListReparse(ctx, afterID, reparseBatchSize)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		var changed []domain.ReleaseReparse
		for _, r := range batch {
			fields := domain.ParseNameFields(r.TorrentName)
			if !fields.Equal(r.Fields) {
				r.Fields = fields
				changed = append(changed, r)
			}
		}

		if err := s.repo.UpdateParsedFields(ctx, changed); err != nil {
			return nil, err
		}

		result.Checked += len(batch)
		result.Updated += len(changed)
		afterID = batch[len(batch)-1].ID
	}

	s.log.Info().Int("checked", result.Checked).Int("updated", result.Updated).Msg("re-parsed stored releases")

	return &result, nil
}
