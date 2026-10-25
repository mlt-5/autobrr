// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"

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
	leadingBracket = regexp.MustCompile(`^\s*\[([^\]]+)\]`)
	tagSeparators  = regexp.MustCompile(`[-._ ]+`)

	// autobrrTagInfos are autobrr's additions to rls's tags, set up in releasetags.go.
	autobrrTagInfos map[string][]*taginfo.Taginfo

	currentParser atomic.Pointer[VocabParser]
)

// VocabSource says where a vocabulary entry comes from.
type VocabSource string

const (
	// VocabSourceBuiltin is rls's and autobrr's own tags. They can't be changed.
	VocabSourceBuiltin VocabSource = "builtin"
	// VocabSourceList is the new_parser lists in rlsvocab.csv. They can be removed and restored.
	VocabSourceList VocabSource = "list"
	// VocabSourceCustom is entries added in the app.
	VocabSourceCustom VocabSource = "custom"
)

// VocabKinds are the kinds of tag the parser reads from its vocabulary.
var VocabKinds = []string{
	"arch", "audio", "bitdepth", "channels", "codec", "collection", "container", "cut", "edition", "ext", "genre",
	"group", "hdr", "language", "other", "platform", "region", "resolution", "size", "source",
}

// VocabKindTitles are readable names for the kinds.
var VocabKindTitles = map[string]string{
	"arch": "Architecture", "audio": "Audio", "bitdepth": "Bit depth", "channels": "Audio channels",
	"codec": "Codec", "collection": "Streaming service", "container": "Container", "cut": "Cut",
	"edition": "Edition", "ext": "File extension", "genre": "Genre", "group": "Release group", "hdr": "HDR",
	"language": "Language", "other": "Other", "platform": "Platform", "region": "Region",
	"resolution": "Resolution", "size": "Size", "source": "Source",
}

// VocabReleaseTypes are the release types a vocabulary entry can point to.
var VocabReleaseTypes = []string{
	"movie", "episode", "series", "music", "game", "app", "book", "audiobook", "comic", "education", "magazine",
}

// VocabEntry is one tag in the release-name parser's vocabulary.
type VocabEntry struct {
	ID          int64       `json:"id,omitempty"`
	Source      VocabSource `json:"source"`
	Kind        string      `json:"kind"`
	Tag         string      `json:"tag"`
	Title       string      `json:"title"`
	Aliases     []string    `json:"aliases"`
	Pattern     string      `json:"pattern"`
	ReleaseType string      `json:"release_type"`
	Exclusive   bool        `json:"exclusive"`
	Removed     bool        `json:"removed"`
}

// Matches reports whether search is in the entry's tag, title, pattern or other spellings, ignoring case. An empty
// search matches every entry.
func (e *VocabEntry) Matches(search string) bool {
	search = strings.ToLower(search)
	return slices.ContainsFunc(append([]string{e.Tag, e.Title, e.Pattern}, e.Aliases...), func(s string) bool {
		return strings.Contains(strings.ToLower(s), search)
	})
}

// VocabKey names an entry in the new_parser lists.
type VocabKey struct {
	Kind string `json:"kind"`
	Tag  string `json:"tag"`
}

// VocabError is a problem with a vocabulary entry that was given to the app.
type VocabError struct {
	message string
}

func (e *VocabError) Error() string {
	return e.message
}

// VocabErrorf formats a VocabError.
func VocabErrorf(format string, args ...any) error {
	return &VocabError{message: fmt.Sprintf(format, args...)}
}

// Normalize trims the entry's text fields and drops empty aliases.
func (e *VocabEntry) Normalize() {
	e.Kind = strings.TrimSpace(e.Kind)
	e.Tag = strings.TrimSpace(e.Tag)
	e.Title = strings.TrimSpace(e.Title)
	e.Pattern = strings.TrimSpace(e.Pattern)
	e.ReleaseType = strings.TrimSpace(e.ReleaseType)

	aliases := []string{}
	for _, alias := range e.Aliases {
		if alias = strings.TrimSpace(alias); alias != "" {
			aliases = append(aliases, alias)
		}
	}
	e.Aliases = aliases
}

// Validate checks the fields a custom entry is given.
func (e *VocabEntry) Validate() error {
	switch {
	case !slices.Contains(VocabKinds, e.Kind):
		return VocabErrorf("unknown kind %q", e.Kind)
	case e.Tag == "":
		return VocabErrorf("a tag is required")
	case e.ReleaseType != "" && !slices.Contains(VocabReleaseTypes, e.ReleaseType):
		return VocabErrorf("unknown release type %q", e.ReleaseType)
	case e.Exclusive && e.ReleaseType == "":
		return VocabErrorf("an exclusive tag needs a release type")
	}

	if e.Pattern != "" {
		if _, err := regexp.Compile(e.Pattern); err != nil {
			return VocabErrorf("the pattern is not a valid regular expression: %v", err)
		}
	}

	return nil
}

// taginfo builds the parser's tag info for a custom entry. For a tag rls already has, the entry adds spellings and
// keeps rls's title, release type and exclusive unless it gives its own.
func (e *VocabEntry) taginfo(builtin map[string][]*taginfo.Taginfo) (*taginfo.Taginfo, error) {
	var base *taginfo.Taginfo
	for _, info := range builtin[e.Kind] {
		if info.Tag() == e.Tag {
			base = info
			break
		}
	}

	var patterns []string
	switch {
	case e.Pattern != "":
		patterns = append(patterns, e.Pattern)
	case base == nil:
		patterns = append(patterns, tagPattern(e.Tag))
	}
	for _, alias := range e.Aliases {
		patterns = append(patterns, tagPattern(alias))
	}
	if len(patterns) == 0 {
		return nil, VocabErrorf("%s %q is already built in; give other spellings or a pattern", e.Kind, e.Tag)
	}

	title, releaseType, exclusive, other := e.Title, e.ReleaseType, e.Exclusive, ""
	if base != nil {
		if title == "" && base.Title() != base.Tag() {
			title = base.Title()
		}
		if releaseType == "" {
			releaseType = rls.Type(base.Type()).String()
			exclusive = base.Excl()
		}
		other = base.Other()
	}

	excl := ""
	if exclusive {
		excl = "1"
	}

	info, err := taginfo.New(e.Tag, title, strings.Join(patterns, "|"), other, releaseType, excl)
	if err != nil {
		return nil, VocabErrorf("%v", err)
	}

	return info, nil
}

// tagPattern matches a tag with any of - . _ or space, or nothing, at each separator, so "WEB-DL" also finds "WEB.DL"
// and "WEBDL".
func tagPattern(tag string) string {
	words := tagSeparators.Split(tag, -1)
	for i, word := range words {
		words[i] = regexp.QuoteMeta(word)
	}
	return strings.Join(words, `[\-\._ ]?`)
}

// VocabParser is the release-name parser built from a vocabulary.
type VocabParser struct {
	parser rls.Parser
	groups taginfo.FindFunc
}

// NewVocabParser builds a parser from rls's and autobrr's tags, the new_parser lists without the removed entries,
// and the custom entries.
func NewVocabParser(custom []VocabEntry, removed []VocabKey) (*VocabParser, error) {
	builtin := taginfo.All(autobrrTagInfos)

	list := map[string][]*taginfo.Taginfo{}
	for kind, infos := range rlsVocab {
		for _, info := range infos {
			if !slices.Contains(removed, VocabKey{Kind: kind, Tag: info.Tag()}) {
				list[kind] = append(list[kind], info)
			}
		}
	}

	own := map[string][]*taginfo.Taginfo{}
	for _, entry := range custom {
		info, err := entry.taginfo(builtin)
		if err != nil {
			return nil, err
		}
		own[entry.Kind] = append(own[entry.Kind], info)
	}

	groups := slices.Concat(list["group"], own["group"])

	return &VocabParser{
		parser: rls.NewTagParser(taginfo.All(autobrrTagInfos, list, own), rls.DefaultLexers()...),
		groups: taginfo.Find(groups...),
	}, nil
}

// UseVocabParser makes p the parser for every release name parsed from now on.
func UseVocabParser(p *VocabParser) {
	currentParser.Store(p)
}

// Vocab lists every vocabulary entry: the built-in tags, the new_parser lists marked when removed, and the custom
// entries, ordered by kind.
func Vocab(custom []VocabEntry, removed []VocabKey) []VocabEntry {
	var entries []VocabEntry

	builtin := taginfo.All(autobrrTagInfos)
	for _, kind := range VocabKinds {
		for _, info := range builtin[kind] {
			entries = append(entries, vocabEntry(VocabSourceBuiltin, kind, info))
		}

		for _, info := range rlsVocab[kind] {
			entry := vocabEntry(VocabSourceList, kind, info)
			entry.Removed = slices.Contains(removed, VocabKey{Kind: kind, Tag: info.Tag()})
			entries = append(entries, entry)
		}

		for _, entry := range custom {
			if entry.Kind == kind {
				entry.Source = VocabSourceCustom
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// IsListedVocab reports whether the new_parser lists have the tag.
func IsListedVocab(key VocabKey) bool {
	return slices.ContainsFunc(rlsVocab[key.Kind], func(info *taginfo.Taginfo) bool { return info.Tag() == key.Tag })
}

func vocabEntry(source VocabSource, kind string, info *taginfo.Taginfo) VocabEntry {
	title := info.Title()
	if title == info.Tag() {
		title = ""
	}

	return VocabEntry{
		Source:      source,
		Kind:        kind,
		Tag:         info.Tag(),
		Title:       title,
		Aliases:     []string{},
		Pattern:     info.Regexp(),
		ReleaseType: rls.Type(info.Type()).String(),
		Exclusive:   info.Excl(),
	}
}

func parseRelease(name string) rls.Release {
	p := currentParser.Load()
	return p.leadingGroup(name, p.parser.ParseRelease([]byte(name)))
}

// leadingGroup takes a listed group in brackets at the start of a fansub name, such as
// "[SubsPlease] Title - 04 (1080p)", as the group. rls reads that bracket as the site and looks for the group after
// the last "-".
func (p *VocabParser) leadingGroup(name string, rel rls.Release) rls.Release {
	m := leadingBracket.FindStringSubmatch(name)
	if m == nil {
		return rel
	}

	group := strings.Trim(m[1], "\t\n\f\r ")
	if group == "" || p.groups(group) == nil {
		return rel
	}

	rel.Group = group
	if rel.Site == group {
		rel.Site = ""
	}

	return rel
}
