// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelease_ParseString_Vocab(t *testing.T) {
	tests := []struct {
		name     string
		group    string
		website  string
		channels string
		audio    []string
	}{
		{name: "[SubsPlease] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv", group: "SubsPlease"},
		{name: "[SubsPlease] Some Show - Part-Two - 03 (1080p) [ABCD1234].mkv", group: "SubsPlease"},
		{name: "[ToonsHub] HORROR COLLECTOR S01E03 1080p CR WEB-DL AAC2.0 H.264 (Kyoufu Collector, Multi-Subs)", group: "ToonsHub", website: "CR", channels: "2.0", audio: []string{"AAC"}},
		{name: "Movie.2020.1080p.FAND.WEB-DL.DDP5.1.H.264-Fight-BB", group: "Fight-BB", website: "FAND", channels: "5.1", audio: []string{"DDP"}},
		{name: "Show.S01E01.1080p.WEB-DL.STEREO.H264-GRP", group: "GRP", channels: "2.0", audio: []string{}},
		{name: "Album.VLS2004.FLAC-GRP", group: "GRP", audio: []string{"FLAC"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRelease(IndexerMinimal{})
			r.ParseString(tt.name)

			assert.Equal(t, tt.group, r.Group)
			assert.Equal(t, tt.website, r.Website)
			assert.Equal(t, tt.channels, r.AudioChannels)
			if tt.audio != nil {
				assert.ElementsMatch(t, tt.audio, r.Audio)
			}
		})
	}
}

func TestLeadingGroup(t *testing.T) {
	rel := parseRelease("[ZqNotListed] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv")
	assert.Empty(t, rel.Group, "an unlisted bracket stays the site")
	assert.Equal(t, "ZqNotListed", rel.Site)

	rel = parseRelease("[SubsPlease] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv")
	assert.Equal(t, "SubsPlease", rel.Group)
	assert.Empty(t, rel.Site)
}
