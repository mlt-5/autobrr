// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"fmt"
	"regexp"
	"slices"
	"unicode/utf8"
)

// NumberingRanges returns where a release name holds its year, season and episode, as [start, end) offsets in
// runes. A number on its own is taken from its last place in the name, since a title can hold the same number
// earlier, as in "Blade.Runner.2049.2017".
func NumberingRanges(name string, year, season, episode int) [][2]int {
	var ranges [][2]int
	add := func(pattern string, all bool) bool {
		re := regexp.MustCompile(`(?i)(?:^|[^\pL\pN])(` + pattern + `)(?:[^\pL\pN]|$)`)

		var found [][2]int
		for start := 0; start < len(name); {
			m := re.FindStringSubmatchIndex(name[start:])
			if m == nil {
				break
			}
			found = append(found, [2]int{runeOffset(name, start+m[2]), runeOffset(name, start+m[3])})
			start += max(m[3], 1)
		}
		if len(found) == 0 {
			return false
		}

		if !all {
			found = found[len(found)-1:]
		}
		ranges = append(ranges, found...)
		return true
	}

	switch {
	case season > 0 && episode > 0 && (add(fmt.Sprintf(`S0*%d[ ._-]?E0*%d`, season, episode), true) || add(fmt.Sprintf(`0*%dx0*%d`, season, episode), true)):
	default:
		if season > 0 {
			add(fmt.Sprintf(`S0*%d|Season[ ._-]?0*%d`, season, season), true)
		}
		if episode > 0 {
			add(fmt.Sprintf(`(?:E|EP|Episode[ ._-]?)?0*%d(?:v\d+)?`, episode), false)
		}
	}

	if year > 0 {
		add(fmt.Sprintf(`%d`, year), false)
	}

	slices.SortFunc(ranges, func(a, b [2]int) int { return a[0] - b[0] })

	return ranges
}

func runeOffset(s string, byteOffset int) int {
	return utf8.RuneCountInString(s[:byteOffset])
}
