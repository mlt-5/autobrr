// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumberingRanges(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{name: "The.Penguin.S01E06.1080p.WEB.h264-ETHEL", want: []string{"S01E06"}},
		{name: "Dune.Part.Two.2024.2160p.UHD.BluRay.x265-GROUP", want: []string{"2024"}},
		{name: "Blade.Runner.2049.2017.1080p.BluRay.x264-GRP", want: []string{"2017"}},
		{name: "[ASW] Yowaki Max - 04 [1080p HEVC x265 10Bit][AAC]", want: []string{"04"}},
		{name: "[ANi] Hotel Inhumans / 殺手旅店 - 17 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]", want: []string{"17"}},
		{name: "【喵萌奶茶屋】★10月新番★[学生会也有洞 / Seitokai ni mo Ana wa Aru!][04][1080p][简日双语]", want: nil}, // the parser reads no episode from it
		{name: "Show.S02.1080p.WEB-DL.x264-GRP", want: []string{"S02"}},
		{name: "Show.2x05.720p.HDTV.x264-GRP", want: []string{"2x05"}},
		{name: "Movie.2020.1080p.WEB.x264-GRP", want: []string{"2020"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRelease(IndexerMinimal{})
			r.ParseString(tt.name)

			var got []string
			runes := []rune(tt.name)
			for _, rg := range NumberingRanges(tt.name, r.Year, r.Season, r.Episode) {
				got = append(got, string(runes[rg[0]:rg[1]]))
			}
			assert.Equal(t, tt.want, got, "year %d season %d episode %d", r.Year, r.Season, r.Episode)
		})
	}
}
