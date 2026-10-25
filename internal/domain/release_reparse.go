// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"reflect"
	"slices"
	"strings"
)

// ReleaseParsedFields are the stored release fields that come from parsing the release name, in the form they are
// stored in.
type ReleaseParsedFields struct {
	Title           string
	SubTitle        string
	Season          int
	Episode         int
	Year            int
	Month           int
	Day             int
	Resolution      string
	Source          string
	Codec           string
	Container       string
	HDR             string
	Audio           string
	AudioChannels   string
	Group           string
	Region          string
	Language        string
	Edition         string
	Cut             string
	Website         string
	MediaProcessing string
	Type            string
	Proper          bool
	Repack          bool
	Hybrid          bool
	Other           []string
	NormalizedHash  string
}

// ReleaseReparse is a stored release's name and parsed fields.
type ReleaseReparse struct {
	ID          int64
	TorrentName string
	Fields      ReleaseParsedFields
}

// ReparseResult says how many stored releases a re-parse checked and how many it changed.
type ReparseResult struct {
	Checked int `json:"checked"`
	Updated int `json:"updated"`
}

// ParseNameFields parses a release name with the vocabulary in use, as if the release had just arrived with no
// announced values.
func ParseNameFields(name string) ReleaseParsedFields {
	r := NewRelease(IndexerMinimal{})
	r.ParseString(name)

	other := r.Other
	if other == nil {
		other = []string{}
	}

	return ReleaseParsedFields{
		Title:           r.Title,
		SubTitle:        r.SubTitle,
		Season:          r.Season,
		Episode:         r.Episode,
		Year:            r.Year,
		Month:           r.Month,
		Day:             r.Day,
		Resolution:      r.Resolution,
		Source:          r.Source,
		Codec:           strings.Join(r.Codec, ","),
		Container:       r.Container,
		HDR:             strings.Join(r.HDR, ","),
		Audio:           strings.Join(r.Audio, ","),
		AudioChannels:   r.AudioChannels,
		Group:           r.Group,
		Region:          r.Region,
		Language:        strings.Join(r.Language, ","),
		Edition:         strings.Join(r.Edition, ","),
		Cut:             strings.Join(r.Cut, ","),
		Website:         r.Website,
		MediaProcessing: r.MediaProcessing,
		Type:            r.Type.String(),
		Proper:          r.Proper,
		Repack:          r.Repack,
		Hybrid:          r.Hybrid,
		Other:           other,
		NormalizedHash:  r.NormalizedHash,
	}
}

// Equal reports whether f and o hold the same values.
func (f ReleaseParsedFields) Equal(o ReleaseParsedFields) bool {
	other, oOther := f.Other, o.Other
	f.Other, o.Other = nil, nil
	return reflect.DeepEqual(f, o) && slices.Equal(other, oOther)
}
