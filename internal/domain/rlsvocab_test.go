// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// useVocab swaps in a parser for one test and puts the default one back after it.
func useVocab(t *testing.T, custom []VocabEntry, removed []VocabKey) {
	t.Helper()

	p, err := NewVocabParser(custom, removed)
	require.NoError(t, err)
	UseVocabParser(p)

	t.Cleanup(func() {
		p, err := NewVocabParser(nil, nil)
		require.NoError(t, err)
		UseVocabParser(p)
	})
}

func TestVocabParser_Custom(t *testing.T) {
	useVocab(t, []VocabEntry{
		{Kind: "group", Tag: "ZqFansubs", Aliases: []string{}},
		{Kind: "collection", Tag: "AMZN", Aliases: []string{"Prime Video"}},
		{Kind: "collection", Tag: "ZQ-TV", Title: "Zq Television", Aliases: []string{}},
	}, nil)

	rel := parseRelease("[ZqFansubs] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv")
	assert.Equal(t, "ZqFansubs", rel.Group, "a custom group in leading brackets is the group")
	assert.Empty(t, rel.Site)

	rel = parseRelease("Show.S01E01.1080p.PrimeVideo.WEB-DL.DDP5.1.H.264-GRP")
	assert.Equal(t, "AMZN", rel.Collection, "an alias is reported as its tag")

	rel = parseRelease("Show.S01E01.1080p.ZQ.TV.WEB-DL.DDP5.1.H.264-GRP")
	assert.Equal(t, "ZQ-TV", rel.Collection, "a separator in a tag matches any separator")

	rel = parseRelease("Show.S01E01.1080p.ZQTV.WEB-DL.DDP5.1.H.264-GRP")
	assert.Equal(t, "ZQ-TV", rel.Collection, "or none")
}

func TestVocabParser_Removed(t *testing.T) {
	useVocab(t, nil, []VocabKey{{Kind: "group", Tag: "SubsPlease"}})

	rel := parseRelease("[SubsPlease] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv")
	assert.Empty(t, rel.Group, "a removed group is not taken from leading brackets")
	assert.Equal(t, "SubsPlease", rel.Site)

	rel = parseRelease("[ToonsHub] HORROR COLLECTOR S01E03 1080p CR WEB-DL AAC2.0 H.264")
	assert.Equal(t, "ToonsHub", rel.Group, "other listed groups stay")
}

func TestVocabParser_Errors(t *testing.T) {
	_, err := NewVocabParser([]VocabEntry{{Kind: "collection", Tag: "AMZN"}}, nil)
	assert.ErrorContains(t, err, "already known")
}

func TestVocabEntry_Validate(t *testing.T) {
	for _, tt := range []struct {
		entry VocabEntry
		err   string
	}{
		{entry: VocabEntry{Kind: "group", Tag: "GRP"}},
		{entry: VocabEntry{Kind: "codec", Tag: "VVC", ReleaseType: "movie", Exclusive: true}},
		{entry: VocabEntry{Kind: "codec", Tag: "VVC", Pattern: `vvc|h\.?266`}},
		{entry: VocabEntry{Kind: "nope", Tag: "GRP"}, err: "unknown kind"},
		{entry: VocabEntry{Kind: "group"}, err: "a tag is required"},
		{entry: VocabEntry{Kind: "group", Tag: "GRP", ReleaseType: "film"}, err: "unknown release type"},
		{entry: VocabEntry{Kind: "group", Tag: "GRP", Exclusive: true}, err: "needs a release type"},
		{entry: VocabEntry{Kind: "group", Tag: "GRP", Pattern: "(grp"}, err: "not a valid regular expression"},
	} {
		err := tt.entry.Validate()
		if tt.err == "" {
			assert.NoError(t, err, tt.entry.Tag)
			continue
		}

		var vocabErr *VocabError
		require.ErrorAs(t, err, &vocabErr)
		assert.ErrorContains(t, err, tt.err)
	}
}

func TestVocabEntry_Normalize(t *testing.T) {
	entry := VocabEntry{Kind: " group ", Tag: " GRP ", Aliases: []string{" G R P ", "", "  "}}
	entry.Normalize()

	assert.Equal(t, VocabEntry{Kind: "group", Tag: "GRP", Aliases: []string{"G R P"}}, entry)
}

func TestTagPattern(t *testing.T) {
	assert.Equal(t, `WEB[\-\._ ]?DL`, tagPattern("WEB-DL"))
	assert.Equal(t, `H[\-\._ ]?264\+`, tagPattern("H.264+"))
}

func TestVocab(t *testing.T) {
	custom := []VocabEntry{{ID: 1, Kind: "group", Tag: "ZqFansubs", Aliases: []string{}}}
	entries := Vocab(custom, []VocabKey{{Kind: "group", Tag: "SubsPlease"}})

	count := map[VocabSource]int{}
	seen := map[string]bool{}
	for _, entry := range entries {
		count[entry.Source]++

		id := string(entry.Source) + "/" + entry.Kind + "/" + entry.Tag
		assert.False(t, seen[id], "%s is listed once", id)
		seen[id] = true

		switch {
		case entry.Kind == "collection" && entry.Tag == "AMZN" && entry.Source == VocabSourceBuiltin:
			assert.Equal(t, "Amazon", entry.Title)
		case entry.Kind == "group" && entry.Tag == "SubsPlease":
			assert.Equal(t, VocabSourceList, entry.Source)
			assert.True(t, entry.Removed)
		case entry.Kind == "group" && entry.Tag == "ToonsHub":
			assert.False(t, entry.Removed)
		case entry.Kind == "group" && entry.Tag == "ZqFansubs":
			assert.Equal(t, VocabSourceCustom, entry.Source)
		}
	}

	assert.Equal(t, 175, count[VocabSourceList])
	assert.Equal(t, 1, count[VocabSourceCustom])
	assert.Greater(t, count[VocabSourceBuiltin], 800)

	assert.True(t, IsListedVocab(VocabKey{Kind: "group", Tag: "SubsPlease"}))
	assert.False(t, IsListedVocab(VocabKey{Kind: "group", Tag: "CODEX"}), "built-in tags are not in the lists")
}

func TestVocabParser_ShortLeadingGroup(t *testing.T) {
	useVocab(t, []VocabEntry{{Kind: "group", Tag: "ASW", Aliases: []string{}}}, nil)

	rel := parseRelease("[ASW] Yowaki Max - 04 [1080p HEVC x265 10Bit][AAC]")
	assert.Equal(t, "ASW", rel.Group)
	assert.Equal(t, "Yowaki Max", rel.Title, "a short group in brackets is not taken as the title")
	assert.Equal(t, 4, rel.Episode)
	assert.Empty(t, rel.Site)
}

func TestPreviewVocab(t *testing.T) {
	const name = "[ASW] Yowaki Max - 04 [1080p HEVC x265 10Bit][AAC]"

	p, err := NewVocabParser(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, VocabPreview{
		Fields: []VocabPreviewField{
			{Label: "Type", Value: "episode"},
			{Label: "Title", Value: "Yowaki Max"},
			{Label: "Episode", Value: "4"},
			{Label: "Resolution", Value: "1080p"},
			{Label: "Codec", Value: "HEVC, x265"},
			{Label: "Audio", Value: "AAC"},
			{Label: "Site", Value: "ASW"},
		},
		Unused: []string{},
	}, PreviewVocab(p, name))

	p, err = NewVocabParser([]VocabEntry{{Kind: "group", Tag: "ASW"}}, nil)
	require.NoError(t, err)
	preview := PreviewVocab(p, name)
	assert.Contains(t, preview.Fields, VocabPreviewField{Label: "Group", Value: "ASW"})
	assert.NotContains(t, preview.Fields, VocabPreviewField{Label: "Site", Value: "ASW"})

	p, err = NewVocabParser(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Zqfoo"}, PreviewVocab(p, "Movie.2020.1080p.BluRay.Zqfoo.x264-GRP").Unused, "words the parser could not place")
}

func TestVocabParser_ListedTag(t *testing.T) {
	useVocab(t, []VocabEntry{{Kind: "group", Tag: "SubsPlease", Aliases: []string{"Subs Please"}}}, nil)

	rel := parseRelease("[Subs Please] Kyoufu Collector - 03 (1080p) [ABCD1234].mkv")
	assert.Equal(t, "Subs Please", rel.Group, "an entry for a listed tag adds spellings to it")

	_, err := NewVocabParser([]VocabEntry{{Kind: "group", Tag: "SubsPlease"}}, nil)
	assert.ErrorContains(t, err, "already known")
}
