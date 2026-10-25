// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package database

import (
	"database/sql/driver"

	"github.com/autobrr/autobrr/internal/domain"

	"modernc.org/sqlite"
)

func init() {
	// vocab_group(group) is 1 when the vocabulary has the release group. It asks the parser in use, so it follows
	// vocabulary changes without storing anything.
	sqlite.MustRegisterScalarFunction("vocab_group", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		group, _ := args[0].(string)
		if domain.IsVocabGroup(group) {
			return int64(1), nil
		}
		return int64(0), nil
	})
}
