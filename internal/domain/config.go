// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

type Config struct {
	ConfigPath          string
	LogLevel            string `toml:"logLevel"`
	LogPath             string `toml:"logPath"`
	LogMaxSize          int    `toml:"logMaxSize"`
	LogMaxBackups       int    `toml:"logMaxBackups"`
	CustomDefinitions   string `toml:"customDefinitions"`
	DatabaseDSN         string `toml:"databaseDSN"`
	DatabaseMaxBackups  int    `toml:"databaseMaxBackups"`
	DatabaseAutoMigrate bool   `toml:"databaseAutoMigrate"`
}

type ConfigUpdate struct {
	LogLevel *string `json:"log_level,omitempty"`
	LogPath  *string `json:"log_path,omitempty"`
}
