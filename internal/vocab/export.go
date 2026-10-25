// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package vocab

import (
	"context"
	"io"
	"strings"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/xlsx"
)

var exportColumns = []xlsx.Column{
	{Name: "Kind", Width: 18},
	{Name: "Tag", Width: 18},
	{Name: "Title", Width: 30},
	{Name: "Other spellings", Width: 30},
	{Name: "Pattern", Width: 50},
	{Name: "Release type", Width: 13},
	{Name: "Only for this type", Width: 17},
	{Name: "Source", Width: 10},
	{Name: "Removed", Width: 9},
}

var sourceLabels = map[domain.VocabSource]string{
	domain.VocabSourceBuiltin: "Built in",
	domain.VocabSourceList:    "List",
	domain.VocabSourceCustom:  "Added",
}

// Export writes the vocabulary entries of kind (or every kind when empty) that match search as an Excel workbook.
func (s *Service) Export(ctx context.Context, w io.Writer, kind, search string) error {
	entries, err := s.List(ctx)
	if err != nil {
		return err
	}

	book := xlsx.NewWriter(w)
	if err := book.AddSheet("Vocabulary", exportColumns); err != nil {
		return err
	}

	for _, entry := range entries {
		if (kind != "" && entry.Kind != kind) || !entry.Matches(search) {
			continue
		}

		if err := book.WriteRow(exportRow(entry)...); err != nil {
			return err
		}
	}

	return book.Close()
}

func exportRow(e domain.VocabEntry) []xlsx.Cell {
	kind := domain.VocabKindTitles[e.Kind]
	if kind == "" {
		kind = e.Kind
	}

	return []xlsx.Cell{
		xlsx.String(kind),
		xlsx.String(e.Tag),
		xlsx.String(e.Title),
		xlsx.String(strings.Join(e.Aliases, ", ")),
		xlsx.String(e.Pattern),
		xlsx.String(e.ReleaseType),
		xlsx.String(yes(e.Exclusive)),
		xlsx.String(sourceLabels[e.Source]),
		xlsx.String(yes(e.Removed)),
	}
}

func yes(b bool) string {
	if b {
		return "Yes"
	}
	return ""
}
