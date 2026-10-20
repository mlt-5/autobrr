// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

const RedactedStr = "<redacted>"

func RedactString(s string) string {
	if len(s) == 0 {
		return ""
	}

	return RedactedStr
}

func IsRedactedString(s string) bool {
	if s == "" {
		return false
	}
	return s == RedactedStr
}
