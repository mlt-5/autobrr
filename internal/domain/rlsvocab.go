// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/autobrr/rls"
	"github.com/autobrr/rls/taginfo"
)

// rlsvocab.csv is go/vocab.csv from new_parser, generated there from its vocab/ lists; copy it again after
// changing them. Its entries are tried after rls's own, so they only add tags rls doesn't already read.
//
//go:embed rlsvocab.csv
var rlsVocabCsv []byte

var (
	rlsVocab       = taginfo.MustLoadBytes(rlsVocabCsv)
	listedGroup    = taginfo.Find(rlsVocab["group"]...)
	leadingBracket = regexp.MustCompile(`^\s*\[([^\]]+)\]`)
)

func parseRelease(name string) rls.Release {
	return leadingGroup(name, rls.ParseString(name))
}

// leadingGroup takes a listed group in brackets at the start of a fansub name, such as
// "[SubsPlease] Title - 04 (1080p)", as the group. rls reads that bracket as the site and looks for the group after
// the last "-".
func leadingGroup(name string, rel rls.Release) rls.Release {
	m := leadingBracket.FindStringSubmatch(name)
	if m == nil {
		return rel
	}

	group := strings.Trim(m[1], "\t\n\f\r ")
	if group == "" || listedGroup(group) == nil {
		return rel
	}

	rel.Group = group
	if rel.Site == group {
		rel.Site = ""
	}

	return rel
}
