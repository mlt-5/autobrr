// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strconv"
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

// taginfo builds the parser's tag info for a custom entry. For a tag the parser already knows, the entry adds
// spellings and keeps the known title, release type and exclusive unless it gives its own.
func (e *VocabEntry) taginfo(known map[string][]*taginfo.Taginfo) (*taginfo.Taginfo, error) {
	var base *taginfo.Taginfo
	for _, info := range known[e.Kind] {
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
		return nil, VocabErrorf("%s %q is already known; give other spellings or a pattern", e.Kind, e.Tag)
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
	list := map[string][]*taginfo.Taginfo{}
	for kind, infos := range rlsVocab {
		for _, info := range infos {
			if !slices.Contains(removed, VocabKey{Kind: kind, Tag: info.Tag()}) {
				list[kind] = append(list[kind], info)
			}
		}
	}

	known := taginfo.All(autobrrTagInfos, list)
	own := map[string][]*taginfo.Taginfo{}
	for _, entry := range custom {
		info, err := entry.taginfo(known)
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

// Parse parses a release name.
func (p *VocabParser) Parse(name string) rls.Release {
	rel := p.parser.ParseRelease([]byte(name))

	m := leadingBracket.FindStringSubmatch(name)
	if m == nil {
		return rel
	}

	group := strings.Trim(m[1], "\t\n\f\r ")
	if group == "" || p.groups(group) == nil {
		return rel
	}

	// rls reads a bracket holding a known tag under 5 characters as plain text, which then becomes the title
	if rel.Site != group {
		rel = p.parser.ParseRelease([]byte(name[len(m[0]):]))
	}

	return leadingGroup(rel, group)
}

func parseRelease(name string) rls.Release {
	return currentParser.Load().Parse(name)
}

// leadingGroup takes a listed group in brackets at the start of a fansub name, such as
// "[SubsPlease] Title - 04 (1080p)", as the group. rls reads that bracket as the site and looks for the group after
// the last "-".
func leadingGroup(rel rls.Release, group string) rls.Release {
	rel.Group = group
	if rel.Site == group {
		rel.Site = ""
	}

	return rel
}

// VocabPreview is what the parser reads from a release name.
type VocabPreview struct {
	Fields []VocabPreviewField `json:"fields"`
	Unused []string            `json:"unused"`
}

// VocabPreviewField is one value the parser read.
type VocabPreviewField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// PreviewVocab parses name with p into the fields a release shows, and the words p could not place.
func PreviewVocab(p *VocabParser, name string) VocabPreview {
	rel := p.Parse(name)

	r := NewRelease(IndexerMinimal{})
	r.ParseStringWith(p, name)

	number := func(n int) string {
		if n == 0 {
			return ""
		}
		return strconv.Itoa(n)
	}

	preview := VocabPreview{Fields: []VocabPreviewField{}, Unused: []string{}}
	for _, field := range []VocabPreviewField{
		{Label: "Type", Value: r.Type.String()},
		{Label: "Title", Value: r.Title},
		{Label: "Subtitle", Value: r.SubTitle},
		{Label: "Year", Value: number(r.Year)},
		{Label: "Season", Value: number(r.Season)},
		{Label: "Episode", Value: number(r.Episode)},
		{Label: "Resolution", Value: r.Resolution},
		{Label: "Source", Value: r.Source},
		{Label: "Codec", Value: strings.Join(r.Codec, ", ")},
		{Label: "HDR", Value: strings.Join(r.HDR, ", ")},
		{Label: "Audio", Value: strings.Join(r.Audio, ", ")},
		{Label: "Channels", Value: r.AudioChannels},
		{Label: "Container", Value: r.Container},
		{Label: "Streaming service", Value: r.Website},
		{Label: "Language", Value: strings.Join(r.Language, ", ")},
		{Label: "Region", Value: r.Region},
		{Label: "Edition", Value: strings.Join(r.Edition, ", ")},
		{Label: "Cut", Value: strings.Join(r.Cut, ", ")},
		{Label: "Other", Value: strings.Join(r.Other, ", ")},
		{Label: "Group", Value: r.Group},
		{Label: "Site", Value: rel.Site},
	} {
		if field.Value != "" {
			preview.Fields = append(preview.Fields, field)
		}
	}

	for _, tag := range rel.Unused() {
		if text := strings.TrimSpace(tag.Text()); text != "" {
			preview.Unused = append(preview.Unused, text)
		}
	}

	return preview
}
